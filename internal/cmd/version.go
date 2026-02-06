package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/instantly-cli/internal/buildinfo"
	"github.com/salmonumbrella/instantly-cli/internal/outfmt"
)

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version/build info",
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode := outfmt.ModeFrom(cmd.Context())
			v := map[string]any{
				"version": buildinfo.Version,
				"commit":  buildinfo.Commit,
				"date":    buildinfo.Date,
			}
			if mode == outfmt.Text {
				// Keep text output human-scannable.
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s (%s) %s\n", buildinfo.Version, buildinfo.Commit, buildinfo.Date)
				return nil
			}
			return printResult(cmd, "version", v, nil)
		},
	}
	return cmd
}
