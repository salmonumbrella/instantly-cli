package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newSupersearchEnrichmentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "supersearch-enrichment",
		Aliases: []string{"supersearch", "enrichment"},
		Short:   "Supersearch enrichment operations",
	}

	cmd.AddCommand(newSupersearchEnrichmentCreateCmd())
	cmd.AddCommand(newSupersearchEnrichmentGetCmd())
	cmd.AddCommand(newSupersearchEnrichmentHistoryCmd())
	cmd.AddCommand(newSupersearchEnrichmentUpdateSettingsCmd())

	cmd.AddCommand(newSupersearchEnrichmentRunCmd())
	cmd.AddCommand(newSupersearchEnrichmentAICmd())
	cmd.AddCommand(newSupersearchEnrichmentCountLeadsCmd())
	cmd.AddCommand(newSupersearchEnrichmentEnrichLeadsCmd())

	return cmd
}

func newSupersearchEnrichmentCreateCmd() *cobra.Command {
	return supersearchPostCmd(
		"create",
		"supersearch_enrichment.create",
		"Create a supersearch enrichment resource (POST /supersearch-enrichment, requires --confirm)",
		"/supersearch-enrichment",
	)
}

func newSupersearchEnrichmentGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <resource_id>",
		Short: "Get supersearch enrichment resource (GET /supersearch-enrichment/{resource_id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "supersearch_enrichment.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "supersearch_enrichment.get", fmt.Errorf("resource_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/supersearch-enrichment/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "supersearch_enrichment.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "supersearch_enrichment.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newSupersearchEnrichmentHistoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history <resource_id>",
		Short: "Get supersearch enrichment history (GET /supersearch-enrichment/history/{resource_id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "supersearch_enrichment.history", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "supersearch_enrichment.history", fmt.Errorf("resource_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/supersearch-enrichment/history/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "supersearch_enrichment.history", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "supersearch_enrichment.history", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newSupersearchEnrichmentUpdateSettingsCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "update-settings <resource_id>",
		Short: "Update supersearch enrichment settings (PATCH /supersearch-enrichment/{resource_id}/settings, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "supersearch_enrichment.update_settings", fmt.Errorf("refusing to update without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "supersearch_enrichment.update_settings", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "supersearch_enrichment.update_settings", fmt.Errorf("resource_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "supersearch_enrichment.update_settings", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "supersearch_enrichment.update_settings", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/supersearch-enrichment/"+url.PathEscape(id)+"/settings", nil, body)
			if err != nil {
				return printError(cmd, "supersearch_enrichment.update_settings", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "supersearch_enrichment.update_settings", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm updating settings")
	return cmd
}

func newSupersearchEnrichmentRunCmd() *cobra.Command {
	return supersearchPostCmd(
		"run",
		"supersearch_enrichment.run",
		"Run a supersearch enrichment operation (POST /supersearch-enrichment/run, requires --confirm)",
		"/supersearch-enrichment/run",
	)
}

func newSupersearchEnrichmentAICmd() *cobra.Command {
	return supersearchPostCmd(
		"ai",
		"supersearch_enrichment.ai",
		"Run supersearch enrichment AI (POST /supersearch-enrichment/ai, requires --confirm)",
		"/supersearch-enrichment/ai",
	)
}

func newSupersearchEnrichmentCountLeadsCmd() *cobra.Command {
	return supersearchPostCmd(
		"count-leads",
		"supersearch_enrichment.count_leads",
		"Count leads from supersearch (POST /supersearch-enrichment/count-leads-from-supersearch, requires --confirm)",
		"/supersearch-enrichment/count-leads-from-supersearch",
	)
}

func newSupersearchEnrichmentEnrichLeadsCmd() *cobra.Command {
	return supersearchPostCmd(
		"enrich-leads",
		"supersearch_enrichment.enrich_leads",
		"Enrich leads from supersearch (POST /supersearch-enrichment/enrich-leads-from-supersearch, requires --confirm)",
		"/supersearch-enrichment/enrich-leads-from-supersearch",
	)
}

func supersearchPostCmd(use, op, short, endpoint string) *cobra.Command {
	var (
		resourceID string
		dataJSON   string
		dataFile   string
		confirm    bool
	)
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, op, fmt.Errorf("refusing to run without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, op, err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, op, err, nil)
			}
			if len(body) == 0 {
				// Allow --resource-id alone to form the minimum body for these endpoints.
				if strings.TrimSpace(resourceID) == "" {
					return printError(cmd, op, fmt.Errorf("provide --data-json or --data-file"), nil)
				}
				body = map[string]any{"resource_id": strings.TrimSpace(resourceID)}
			}
			if strings.TrimSpace(resourceID) != "" && body["resource_id"] == nil {
				body["resource_id"] = strings.TrimSpace(resourceID)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), endpoint, nil, body)
			if err != nil {
				return printError(cmd, op, err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, op, resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&resourceID, "resource-id", "", "Resource ID (sets resource_id in JSON body if not already present)")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm action")
	return cmd
}
