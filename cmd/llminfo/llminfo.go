package llminfo

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//go:embed preamble.md
var preamble string

//go:embed postamble.md
var postamble string

// Cmd is the llm-info command.
var Cmd = &cobra.Command{
	Use:     "llm-info",
	Short:   "Print full CLI reference for LLM consumption",
	Long:    "Output a comprehensive markdown description of every command, flag, and workflow so an LLM agent can understand and use the Adapto CMS CLI. The command reference is generated from the live command tree, so it cannot diverge from the actual CLI.",
	Example: "adapto llm-info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(reference(cmd.Root()))
	},
}

func reference(root *cobra.Command) string {
	var b strings.Builder
	b.WriteString(preamble)
	writeGlobalFlags(&b, root)
	b.WriteString("## Commands\n\n")
	for _, c := range root.Commands() {
		writeCommandDocs(&b, c, 3)
	}
	b.WriteString(postamble)
	return b.String()
}

func writeGlobalFlags(b *strings.Builder, root *cobra.Command) {
	b.WriteString("## Global Flags\n\nEvery command accepts these flags:\n\n| Flag | Type | Description |\n|------|------|-------------|\n")
	root.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden {
			return
		}
		fmt.Fprintf(b, "| `--%s` | %s | %s |\n", f.Name, f.Value.Type(), f.Usage)
	})
	b.WriteString("\n---\n\n")
}

func writeCommandDocs(b *strings.Builder, c *cobra.Command, level int) {
	if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
		return
	}

	header := c.CommandPath()
	if parts := strings.SplitN(c.Use, " ", 2); len(parts) == 2 {
		header += " " + parts[1]
	}
	fmt.Fprintf(b, "%s %s\n\n", strings.Repeat("#", min(level, 4)), header)

	desc := c.Long
	if desc == "" {
		desc = c.Short
	}
	if desc != "" {
		b.WriteString(strings.TrimSpace(desc) + "\n\n")
	}

	if c.Example != "" {
		fmt.Fprintf(b, "```bash\n%s\n```\n\n", strings.TrimSpace(c.Example))
	}

	var rows [][2]string
	c.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if f.Hidden || f.Name == "help" {
			return
		}
		rows = append(rows, [2]string{f.Name, f.Usage})
	})
	if len(rows) > 0 {
		b.WriteString("| Flag | Description |\n|------|-------------|\n")
		for _, r := range rows {
			fmt.Fprintf(b, "| `--%s` | %s |\n", r[0], r[1])
		}
		b.WriteString("\n")
	}

	for _, sub := range c.Commands() {
		writeCommandDocs(b, sub, level+1)
	}

	if level == 3 {
		b.WriteString("---\n\n")
	}
}
