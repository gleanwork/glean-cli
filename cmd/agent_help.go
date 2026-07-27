package cmd

import (
	"github.com/gleanwork/glean-cli/internal/agenthelp"
	"github.com/spf13/cobra"
)

// NewCmdAgentHelp creates the agent-help command: agent-facing usage
// generated from the installed binary, so it is version-accurate by
// construction. It must work unauthenticated — an agent's first command is
// often discovery, before credentials exist.
func NewCmdAgentHelp() *cobra.Command {
	var asJSON bool

	cmd := &cobra.Command{
		Use:   "agent-help [command...]",
		Short: "Show agent-facing usage that matches this binary",
		Long: `Show agent-facing usage generated from the installed glean binary.

With no arguments, prints the environment context (version, server, auth
state, API mode) and a map of all commands with when-to-use guidance. With a
command path, prints that command's flags, payload shapes, subcommands, and
examples.

Because the output comes from the binary itself, it cannot drift from the
installed version — use it as the source of truth when it differs from
static documentation or pre-installed skills.

Examples:
  glean agent-help                     # command map + environment context
  glean agent-help search              # full usage for search
  glean agent-help agents run --json   # machine-readable detail`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := agenthelp.BuildContext(cliVersion)
			docs := agenthelp.Collect(cmd.Root())
			return agenthelp.Render(cmd.OutOrStdout(), ctx, docs, args, asJSON)
		},
	}

	cmd.Flags().BoolVar(&asJSON, "json", false, "Emit machine-readable JSON")
	return cmd
}
