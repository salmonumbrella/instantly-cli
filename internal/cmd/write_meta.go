package cmd

import "github.com/spf13/cobra"

// printWriteResult prints a result and always includes the payload used in meta,
// so agents can replay/debug write operations without reconstructing the request.
func printWriteResult(cmd *cobra.Command, op string, resp any, meta map[string]any, payload any) error {
	outMeta := map[string]any{
		"payload_used": payload,
	}
	for k, v := range meta {
		outMeta[k] = v
	}
	return printResult(cmd, op, resp, outMeta)
}
