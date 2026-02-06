package cmd

import (
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type flagSchema struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Type      string `json:"type"`
	Default   string `json:"default,omitempty"`
	Usage     string `json:"usage,omitempty"`
}

type cmdSchema struct {
	Path         string       `json:"path"`
	Use          string       `json:"use"`
	Aliases      []string     `json:"aliases,omitempty"`
	Short        string       `json:"short,omitempty"`
	Long         string       `json:"long,omitempty"`
	Example      string       `json:"example,omitempty"`
	HTTPMethod   string       `json:"http_method,omitempty"`
	Endpoint     string       `json:"endpoint,omitempty"`
	IsWrite      bool         `json:"is_write,omitempty"`
	HasConfirm   bool         `json:"has_confirm,omitempty"`
	NeedsConfirm bool         `json:"needs_confirm,omitempty"`
	PayloadFlags []string     `json:"payload_flags,omitempty"`
	Flags        []flagSchema `json:"flags,omitempty"`
	Subcommands  []cmdSchema  `json:"subcommands,omitempty"`
}

type rootSchema struct {
	Name            string       `json:"name"`
	PersistentFlags []flagSchema `json:"persistent_flags,omitempty"`
	Commands        []cmdSchema  `json:"commands"`
}

func newSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "schema",
		Short: "Print a machine-readable schema of commands/flags (for agents)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			root := cmd.Root()
			// Root persistent flags apply to every command; keep them separate to reduce duplication.
			out := rootSchema{
				Name:            root.Name(),
				PersistentFlags: flagsFromFlagSet(root.PersistentFlags()),
				Commands:        subcommandsSchema(root),
			}
			return printResult(cmd, "schema", out, nil)
		},
	}
	return cmd
}

func subcommandsSchema(cmd *cobra.Command) []cmdSchema {
	sub := cmd.Commands()
	out := make([]cmdSchema, 0, len(sub))
	for _, sc := range sub {
		// Hide non-runnable or hidden commands.
		if !sc.IsAvailableCommand() {
			continue
		}
		// Cobra help/completion are usually present; omit them to keep the schema focused.
		if sc.Name() == "help" || sc.Name() == "completion" {
			continue
		}
		out = append(out, commandSchema(sc))
	}
	return out
}

func commandSchema(cmd *cobra.Command) cmdSchema {
	out := cmdSchema{
		Path:        cmd.CommandPath(),
		Use:         cmd.Use,
		Aliases:     cmd.Aliases,
		Short:       cmd.Short,
		Long:        cmd.Long,
		Example:     cmd.Example,
		Flags:       flagsFromFlagSet(cmd.Flags()),
		Subcommands: subcommandsSchema(cmd),
	}

	// Parse method/path hints from Short strings like:
	// - "List leads (POST /leads/list)"
	// - "Get webhook event (GET /webhook-events/{id})"
	if m, p, ok := parseMethodAndEndpoint(cmd.Short); ok {
		out.HTTPMethod = m
		out.Endpoint = p
		out.IsWrite = m != "GET"
	}

	// Confirmation gating is a key agent-safety signal.
	out.HasConfirm = cmd.Flags() != nil && cmd.Flags().Lookup("confirm") != nil
	out.NeedsConfirm = out.HasConfirm && strings.Contains(cmd.Short, "requires --confirm")

	// Payload-related flags: a fast way for an agent to know how to supply input.
	out.PayloadFlags = payloadFlagsFrom(cmd.Flags())

	return out
}

func flagsFromFlagSet(fs *pflag.FlagSet) []flagSchema {
	if fs == nil {
		return nil
	}
	out := []flagSchema{}
	fs.VisitAll(func(f *pflag.Flag) {
		out = append(out, flagSchema{
			Name:      f.Name,
			Shorthand: f.Shorthand,
			Type:      f.Value.Type(),
			Default:   f.DefValue,
			Usage:     f.Usage,
		})
	})
	if len(out) == 0 {
		return nil
	}
	return out
}

// Match "(...)" suffixes in Short strings, extracting the HTTP method and endpoint path.
// Examples:
// - "List leads (POST /leads/list)"
// - "Delete lead (DELETE /leads/{id}, requires --confirm)"
var shortMethodEndpointRE = regexp.MustCompile(`\((GET|POST|PATCH|DELETE)\s+([^,\)]+)`)

func parseMethodAndEndpoint(short string) (method, endpoint string, ok bool) {
	m := shortMethodEndpointRE.FindStringSubmatch(short)
	if len(m) != 3 {
		return "", "", false
	}
	return m[1], m[2], true
}

func payloadFlagsFrom(fs *pflag.FlagSet) []string {
	if fs == nil {
		return nil
	}
	interesting := []string{
		"data-json",
		"data-file",
		"body-json",
		"body-file",
		"headers-json",
		"query",
	}
	var out []string
	for _, name := range interesting {
		if fs.Lookup(name) != nil {
			out = append(out, name)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
