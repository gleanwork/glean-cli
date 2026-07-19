// Package cmd implements the command-line interface for the Glean CLI.
// It provides commands for interacting with Glean's API, managing configuration,
// and performing various operations like search and chat.
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gleanwork/glean-cli/internal/auth"
	gleanClient "github.com/gleanwork/glean-cli/internal/client"
	"github.com/gleanwork/glean-cli/internal/config"
	"github.com/gleanwork/glean-cli/internal/debug"
	clierrors "github.com/gleanwork/glean-cli/internal/errors"
	"github.com/gleanwork/glean-cli/internal/tui"
	"github.com/gleanwork/glean-cli/internal/update"
	"github.com/spf13/cobra"
)

var authErrLog = debug.New("auth:login")

// cliVersion is set at startup via SetVersion from the ldflags-injected build version.
var cliVersion = "dev"

// SetVersion records the build-time version for use in --version and update checks.
func SetVersion(v string) {
	cliVersion = v
	if rootCmd != nil {
		rootCmd.Version = v
	}
}

// rootCmd represents the base command when called without any subcommands.
// It provides help information and coordinates all subcommands.
var rootCmd *cobra.Command

// NewCmdRoot creates and returns the root command for the Glean CLI.
// It sets up the base command structure and adds all subcommands.
func NewCmdRoot() *cobra.Command {
	var verbosity int
	var continueSession bool

	cmd := &cobra.Command{
		Use:     "glean",
		Short:   "Glean CLI - A command-line interface for Glean operations.",
		Version: cliVersion,
		Long: heredoc.Doc(`
			Work seamlessly with Glean from your command line.

			Running 'glean' with no arguments opens the full-screen chat TUI.
			Run 'glean --help' for other available commands.
		`),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbosity > 0 {
				debug.Enable()
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			// Skip update notice when the user is already running `glean update`.
			if cmd.Name() == "update" {
				return
			}
			noticeCh := update.CheckAsync(cliVersion)
			if notice := <-noticeCh; notice != "" {
				fmt.Fprintf(os.Stderr, "\n%s\n", notice)
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.LoadConfig()
			if err != nil {
				return err
			}

			// Validate credentials before opening the TUI.
			if _, err := gleanClient.NewFromConfig(); err != nil {
				return authError(err)
			}

			// Always start with a fresh session. Prior conversation context
			// flashing back on startup is jarring and unexpected. Users who
			// want to continue a prior conversation can use --continue.
			session := &tui.Session{}
			if continueSession {
				session = tui.LoadLatest()
			}

			// Resolve identity: "email · host" or just host.
			// Email comes from stored token or decoded from the JWT access token
			// (Glean's RFC 8414 access tokens are JWTs containing the email claim).
			identity := cfg.GleanServerURL
			if tok, err := auth.LoadTokens(cfg.GleanServerURL); err == nil && tok != nil {
				email := tok.Email
				if email == "" {
					email = auth.EmailFromJWT(tok.AccessToken)
				}
				if email != "" {
					identity = email + "  ·  " + cfg.GleanServerURL
				}
			}

			model, err := tui.New(cfg, session, identity, cliVersion, cmd.Context())
			if err != nil {
				return fmt.Errorf("failed to create TUI: %w", err)
			}

			p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithContext(cmd.Context()))
			finalModel, err := p.Run()
			// Print session stats to stdout after TUI exits (like Claude Code).
			if m, ok := finalModel.(*tui.Model); ok && len(m.Session().Turns) > 0 {
				fmt.Println(m.StatsLine())
			}
			return err
		},
		// Silence usage output when an error occurs
		SilenceUsage: true,
		// Ensure consistent error formatting
		SilenceErrors: true,
	}

	cmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "Increase verbosity level (can be used multiple times)")
	cmd.Flags().BoolVar(&continueSession, "continue", false, "Resume the most recent saved chat session")

	cmd.AddGroup(&cobra.Group{
		ID:    "core",
		Title: "Core Commands:",
	})
	cmd.AddGroup(&cobra.Group{
		ID:    "namespace",
		Title: "API Namespace Commands:",
	})

	for _, sub := range []*cobra.Command{
		NewCmdAuth(),
		NewCmdSearch(),
		NewCmdChat(),
		NewCmdAPI(),
		NewCmdAgentHelp(),
		NewCmdSchema(),
	} {
		sub.GroupID = "core"
		cmd.AddCommand(sub)
	}

	for _, sub := range []*cobra.Command{
		NewCmdActivity(),
		NewCmdAgents(),
		NewCmdAnnouncements(),
		NewCmdAnswers(),
		NewCmdCollections(),
		NewCmdDocuments(),
		NewCmdEntities(),
		NewCmdInsights(),
		NewCmdMessages(),
		NewCmdPins(),
		NewCmdShortcuts(),
		NewCmdTools(),
		NewCmdVerification(),
	} {
		sub.GroupID = "namespace"
		cmd.AddCommand(sub)
	}

	// Dev/maintenance commands (hidden from default help)
	genSkills := NewCmdGenerateSkills()
	genSkills.Hidden = true
	cmd.AddCommand(genSkills)

	cmd.AddCommand(NewCmdUpdate())

	// Propagate settings to all subcommands
	for _, subCmd := range cmd.Commands() {
		subCmd.SilenceUsage = true
		subCmd.SilenceErrors = true
	}

	return cmd
}

func authError(err error) error {
	authErrLog.Log("underlying auth error: %v", err)

	suggestion := "Run:  glean auth login\n\n" +
		"Or set environment variables:\n" +
		"  export GLEAN_SERVER_URL=<your Glean server URL>\n" +
		"  export GLEAN_API_TOKEN=your-token\n\n" +
		"See https://developers.glean.com/get-started/authentication for how to find your server URL."
	if !authErrLog.Enabled() {
		suggestion += "\n\n  Tip: re-run with -v or GLEAN_DEBUG=auth:* for details"
	}

	return &clierrors.CLIError{
		UserMessage: "You're not signed in to Glean.",
		Suggestion:  suggestion,
		ExitCode:    clierrors.ExitAuthError,
		Cause:       err,
	}
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		var cliErr *clierrors.CLIError
		if errors.As(err, &cliErr) {
			fmt.Fprintf(os.Stderr, "\n%s\n", cliErr.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		return err
	}
	return nil
}

func init() {
	rootCmd = NewCmdRoot()
}
