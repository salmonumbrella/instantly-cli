package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newLeadListsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lead-lists",
		Aliases: []string{"leadlist", "lists"},
		Short:   "Manage lead lists",
	}

	cmd.AddCommand(newLeadListsListCmd())
	cmd.AddCommand(newLeadListsGetCmd())
	cmd.AddCommand(newLeadListsCreateCmd())
	cmd.AddCommand(newLeadListsUpdateCmd())
	cmd.AddCommand(newLeadListsVerificationStatsCmd())
	cmd.AddCommand(newLeadListsDeleteCmd())

	return cmd
}

func newLeadListsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List lead lists",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_lists.list", err, nil)
			}

			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprintf("%d", limit))
			}
			if startingAfter != "" {
				q.Set("starting_after", startingAfter)
			}
			if search != "" {
				q.Set("search", search)
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "lead_lists.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/lead-lists", q)
			if err != nil {
				return printError(cmd, "lead_lists.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_lists.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination (use response next_starting_after)")
	cmd.Flags().StringVar(&search, "search", "", "Search lead lists")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")

	return cmd
}

func newLeadListsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <list_id>",
		Aliases: []string{"show"},
		Short:   "Get lead list by ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_lists.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_lists.get", fmt.Errorf("list_id is required"), nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/lead-lists/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "lead_lists.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_lists.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newLeadListsCreateCmd() *cobra.Command {
	var (
		name      string
		enrich    bool
		enrichSet bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a lead list",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("enrich") {
				enrichSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_lists.create", err, nil)
			}
			if strings.TrimSpace(name) == "" {
				return printError(cmd, "lead_lists.create", fmt.Errorf("--name is required"), nil)
			}

			body := map[string]any{"name": name}
			if enrichSet {
				body["has_enrichment_task"] = enrich
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/lead-lists", nil, body)
			if err != nil {
				return printError(cmd, "lead_lists.create", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_lists.create", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Lead list name")
	cmd.Flags().BoolVar(&enrich, "enrich", false, "Enable enrichment task")
	return cmd
}

func newLeadListsUpdateCmd() *cobra.Command {
	var (
		name      string
		enrich    bool
		enrichSet bool
	)

	cmd := &cobra.Command{
		Use:   "update <list_id>",
		Short: "Update a lead list (partial)",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("enrich") {
				enrichSet = true
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_lists.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_lists.update", fmt.Errorf("list_id is required"), nil)
			}

			body := map[string]any{}
			if cmd.Flags().Changed("name") {
				body["name"] = name
			}
			if enrichSet {
				body["has_enrichment_task"] = enrich
			}
			if len(body) == 0 {
				return printError(cmd, "lead_lists.update", fmt.Errorf("no fields to update"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/lead-lists/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "lead_lists.update", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_lists.update", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Lead list name")
	cmd.Flags().BoolVar(&enrich, "enrich", false, "Enable enrichment task")
	return cmd
}

func newLeadListsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <list_id>",
		Short: "Delete a lead list (requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "lead_lists.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_lists.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_lists.delete", fmt.Errorf("list_id is required"), nil)
			}

			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/lead-lists/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "lead_lists.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_lists.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newLeadListsVerificationStatsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "verification-stats <list_id>",
		Aliases: []string{"verify-stats", "stats"},
		Short:   "Get email verification stats for a lead list",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_lists.verification_stats", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_lists.verification_stats", fmt.Errorf("list_id is required"), nil)
			}

			path := "/lead-lists/" + url.PathEscape(id) + "/verification-stats"
			resp, meta, err := client.GetJSON(cmdContext(cmd), path, nil)
			if err != nil {
				return printError(cmd, "lead_lists.verification_stats", err, metaFrom(meta, nil))
			}

			return printResult(cmd, "lead_lists.verification_stats", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}
