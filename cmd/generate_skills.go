package cmd

import (
	"github.com/gleanwork/glean-cli/internal/skills"
	"github.com/spf13/cobra"
)

// NewCmdGenerateSkills creates the generate-skills command.
func NewCmdGenerateSkills() *cobra.Command {
	var outputDir string

	cmd := &cobra.Command{
		Use:   "generate-skills",
		Short: "Generate SKILL.md files from the CLI's schema registry",
		Long: `Generates a single Agent Skill (https://agentskills.io) from the CLI's own
schema registry. The skill lives at skills/glean-cli/SKILL.md as a thin router:
it holds a command map and points agents to glean agent-help for flags, payload
shapes, and examples, so per-command detail always matches the installed binary.

The generator is idempotent and removes directories from earlier layouts
(skills/glean-cli-<cmd>/ and skills/glean-cli/reference/).

Run this after adding or modifying commands to keep skills in sync.

  glean generate-skills                    # write to skills/
  glean generate-skills --output-dir out/  # write to out/`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return skills.Generate(outputDir)
		},
	}

	cmd.Flags().StringVar(&outputDir, "output-dir", "skills", "Output directory for generated skills")

	return cmd
}
