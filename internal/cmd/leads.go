package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newLeadsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "leads",
		Aliases: []string{"lead", "l"},
		Short:   "Manage leads",
	}

	cmd.AddCommand(newLeadsListCmd())
	cmd.AddCommand(newLeadsGetCmd())
	cmd.AddCommand(newLeadsCreateCmd())
	cmd.AddCommand(newLeadsUpdateCmd())
	cmd.AddCommand(newLeadsDeleteCmd())
	cmd.AddCommand(newLeadsBulkDeleteCmd())
	cmd.AddCommand(newLeadsMergeCmd())
	cmd.AddCommand(newLeadsUpdateInterestStatusCmd())

	return cmd
}

func newLeadsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		campaign      string
		listID        string
		search        string
		status        string
		distinct      bool
		distinctSet   bool
		bodyJSON      string
		bodyFile      string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List leads (POST /leads/list)",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("distinct") {
				distinctSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.list", err, nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(bodyJSON) != "" || strings.TrimSpace(bodyFile) != "" {
				m, err := readJSONObjectInput(bodyJSON, bodyFile)
				if err != nil {
					return printError(cmd, "leads.list", err, nil)
				}
				mergeMaps(body, m)
			}

			body["limit"] = limit
			if startingAfter != "" {
				body["starting_after"] = startingAfter
			}
			if campaign != "" {
				body["campaign"] = campaign
			}
			if listID != "" {
				body["list_id"] = listID
			}
			if status != "" {
				body["status"] = status
			}
			if search != "" {
				body["search"] = search
			}
			if distinctSet {
				body["distinct_contacts"] = distinct
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/leads/list", nil, body)
			if err != nil {
				return printError(cmd, "leads.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "leads.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination (use response next_starting_after)")
	cmd.Flags().StringVar(&campaign, "campaign", "", "Filter by campaign ID")
	cmd.Flags().StringVar(&listID, "list-id", "", "Filter by lead list ID")
	cmd.Flags().StringVar(&status, "status", "", "Filter by status")
	cmd.Flags().StringVar(&search, "search", "", "Search leads")
	cmd.Flags().BoolVar(&distinct, "distinct", false, "Deduplicate by email (distinct_contacts=true)")
	cmd.Flags().StringVar(&bodyJSON, "body-json", "", "Merge JSON object into request body")
	cmd.Flags().StringVar(&bodyFile, "body-file", "", "Merge JSON object file into request body (or '-' for stdin)")

	return cmd
}

func newLeadsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <lead_id>",
		Aliases: []string{"show"},
		Short:   "Get lead by ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "leads.get", fmt.Errorf("lead_id is required"), nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/leads/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "leads.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "leads.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newLeadsCreateCmd() *cobra.Command {
	var (
		email    string
		campaign string
		first    string
		last     string
		company  string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lead",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.create", err, nil)
			}
			if strings.TrimSpace(email) == "" {
				return printError(cmd, "leads.create", fmt.Errorf("--email is required"), nil)
			}

			body := map[string]any{
				"email": email,
				// Agent-friendly defaults.
				"skip_if_in_workspace": true,
				"skip_if_in_campaign":  true,
			}
			if campaign != "" {
				body["campaign"] = campaign
			}
			if first != "" {
				body["first_name"] = first
			}
			if last != "" {
				body["last_name"] = last
			}
			if company != "" {
				body["company_name"] = company
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/leads", nil, body)
			if err != nil {
				return printError(cmd, "leads.create", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "leads.create", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Lead email")
	cmd.Flags().StringVar(&campaign, "campaign", "", "Campaign ID")
	cmd.Flags().StringVar(&first, "first-name", "", "First name")
	cmd.Flags().StringVar(&last, "last-name", "", "Last name")
	cmd.Flags().StringVar(&company, "company-name", "", "Company name")

	return cmd
}

func newLeadsUpdateCmd() *cobra.Command {
	var (
		first   string
		last    string
		company string
	)

	cmd := &cobra.Command{
		Use:   "update <lead_id>",
		Short: "Update a lead (partial)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "leads.update", fmt.Errorf("lead_id is required"), nil)
			}

			body := map[string]any{}
			if cmd.Flags().Changed("first-name") {
				body["first_name"] = first
			}
			if cmd.Flags().Changed("last-name") {
				body["last_name"] = last
			}
			if cmd.Flags().Changed("company-name") {
				body["company_name"] = company
			}
			if len(body) == 0 {
				return printError(cmd, "leads.update", fmt.Errorf("no fields to update"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/leads/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "leads.update", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "leads.update", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&first, "first-name", "", "First name")
	cmd.Flags().StringVar(&last, "last-name", "", "Last name")
	cmd.Flags().StringVar(&company, "company-name", "", "Company name")

	return cmd
}

func newLeadsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <lead_id>",
		Short: "Delete a lead (requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "leads.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "leads.delete", fmt.Errorf("lead_id is required"), nil)
			}

			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/leads/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "leads.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "leads.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newLeadsBulkDeleteCmd() *cobra.Command {
	var (
		confirm    bool
		queryPairs []string
	)
	cmd := &cobra.Command{
		Use:   "bulk-delete",
		Short: "Bulk delete leads (DELETE /leads, requires --confirm; pass filters via --query)",
		Long: strings.TrimSpace(`
Bulk delete leads using server-side filters.

This is a potentially destructive operation. Use --query to provide filter parameters.

Example:
  instantly leads bulk-delete --confirm --query campaign_id=... --query status=...
`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "leads.bulk_delete", fmt.Errorf("refusing to bulk delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.bulk_delete", err, nil)
			}
			q := url.Values{}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "leads.bulk_delete", err, nil)
			}
			if len(q) == 0 {
				return printError(cmd, "leads.bulk_delete", fmt.Errorf("provide at least one --query filter"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/leads", q)
			if err != nil {
				return printError(cmd, "leads.bulk_delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "leads.bulk_delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Filter query param (repeatable): key=value")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newLeadsMergeCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "merge",
		Short: "Merge leads (POST /leads/merge, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "leads.merge", fmt.Errorf("refusing to merge leads without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.merge", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "leads.merge", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "leads.merge", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/leads/merge", nil, body)
			if err != nil {
				return printError(cmd, "leads.merge", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "leads.merge", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm action")
	return cmd
}

func newLeadsUpdateInterestStatusCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "update-interest-status",
		Short: "Update lead interest status (POST /leads/update-interest-status, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "leads.update_interest_status", fmt.Errorf("refusing to update interest status without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "leads.update_interest_status", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "leads.update_interest_status", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "leads.update_interest_status", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/leads/update-interest-status", nil, body)
			if err != nil {
				return printError(cmd, "leads.update_interest_status", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "leads.update_interest_status", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm action")
	return cmd
}
