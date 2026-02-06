package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func normalizeLeadLabelInterestStatusLabel(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "positive", "neutral", "negative":
		return s, nil
	case "":
		return "", nil
	default:
		return "", fmt.Errorf("invalid interest status %q (allowed: positive|neutral|negative)", s)
	}
}

func newBlockListEntriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "block-list-entries",
		Aliases: []string{"blocklists", "blocklist-entries"},
		Short:   "Manage block list entries",
	}

	cmd.AddCommand(newBlockListEntriesListCmd())
	cmd.AddCommand(newBlockListEntriesGetCmd())
	cmd.AddCommand(newBlockListEntriesCreateCmd())
	cmd.AddCommand(newBlockListEntriesUpdateCmd())
	cmd.AddCommand(newBlockListEntriesDeleteCmd())

	return cmd
}

func newBlockListEntriesListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List block list entries (GET /block-lists-entries)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "block_list_entries.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if strings.TrimSpace(search) != "" {
				q.Set("search", strings.TrimSpace(search))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "block_list_entries.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/block-lists-entries", q)
			if err != nil {
				return printError(cmd, "block_list_entries.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "block_list_entries.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringVar(&search, "search", "", "Search")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newBlockListEntriesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <entry_id>",
		Short: "Get block list entry (GET /block-lists-entries/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "block_list_entries.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "block_list_entries.get", fmt.Errorf("entry_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/block-lists-entries/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "block_list_entries.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "block_list_entries.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newBlockListEntriesCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create block list entry (POST /block-lists-entries)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "block_list_entries.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "block_list_entries.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "block_list_entries.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			// Instantly expects "bl_value"; accept "value" as a convenient alias.
			if v, ok := body["value"]; ok && body["bl_value"] == nil {
				body["bl_value"] = v
				delete(body, "value")
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/block-lists-entries", nil, body)
			if err != nil {
				return printError(cmd, "block_list_entries.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "block_list_entries.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	return cmd
}

func newBlockListEntriesUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update <entry_id>",
		Short: "Update block list entry (PATCH /block-lists-entries/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "block_list_entries.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "block_list_entries.update", fmt.Errorf("entry_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "block_list_entries.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "block_list_entries.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			// Instantly expects "bl_value"; accept "value" as a convenient alias.
			if v, ok := body["value"]; ok && body["bl_value"] == nil {
				body["bl_value"] = v
				delete(body, "value")
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/block-lists-entries/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "block_list_entries.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "block_list_entries.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newBlockListEntriesDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <entry_id>",
		Short: "Delete block list entry (DELETE /block-lists-entries/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "block_list_entries.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "block_list_entries.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "block_list_entries.delete", fmt.Errorf("entry_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/block-lists-entries/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "block_list_entries.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "block_list_entries.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newLeadLabelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "lead-labels",
		Aliases: []string{"labels", "lead-label"},
		Short:   "Manage lead labels",
	}

	cmd.AddCommand(newLeadLabelsListCmd())
	cmd.AddCommand(newLeadLabelsGetCmd())
	cmd.AddCommand(newLeadLabelsCreateCmd())
	cmd.AddCommand(newLeadLabelsUpdateCmd())
	cmd.AddCommand(newLeadLabelsDeleteCmd())

	return cmd
}

func newLeadLabelsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List lead labels (GET /lead-labels)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_labels.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if strings.TrimSpace(search) != "" {
				q.Set("search", strings.TrimSpace(search))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "lead_labels.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/lead-labels", q)
			if err != nil {
				return printError(cmd, "lead_labels.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_labels.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringVar(&search, "search", "", "Search")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newLeadLabelsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <label_id>",
		Short: "Get lead label (GET /lead-labels/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_labels.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_labels.get", fmt.Errorf("label_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/lead-labels/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "lead_labels.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_labels.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newLeadLabelsCreateCmd() *cobra.Command {
	var (
		name           string
		interestStatus string
		dataJSON       string
		dataFile       string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create lead label (POST /lead-labels)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_labels.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "lead_labels.create", err, nil)
			}
			// Instantly expects "label"; accept "name" as a convenient alias.
			if v, ok := body["label"]; !ok || v == nil {
				if v2, ok := body["name"]; ok && v2 != nil {
					body["label"] = v2
					delete(body, "name")
				}
			}
			if strings.TrimSpace(name) != "" {
				body["label"] = strings.TrimSpace(name)
			}
			val, ok := body["label"].(string)
			if !ok || strings.TrimSpace(val) == "" {
				return printError(cmd, "lead_labels.create", fmt.Errorf("label is required (use --name or set label in --data-json/--data-file)"), nil)
			}

			// Required by the API: "interest_status_label" in {positive, neutral, negative}.
			// Agent-friendly default: neutral.
			if _, ok := body["interest_status_label"]; !ok || body["interest_status_label"] == nil {
				// Some agents may pass interest_status instead of interest_status_label.
				if v2, ok := body["interest_status"]; ok && v2 != nil {
					body["interest_status_label"] = v2
					delete(body, "interest_status")
				} else {
					body["interest_status_label"] = interestStatus
				}
			}
			if raw, ok := body["interest_status_label"]; ok && raw != nil {
				s, ok := raw.(string)
				if !ok {
					return printError(cmd, "lead_labels.create", fmt.Errorf("interest_status_label must be a string"), nil)
				}
				norm, err := normalizeLeadLabelInterestStatusLabel(s)
				if err != nil {
					return printError(cmd, "lead_labels.create", err, nil)
				}
				if norm == "" {
					return printError(cmd, "lead_labels.create", fmt.Errorf("interest_status_label is required (positive|neutral|negative)"), nil)
				}
				body["interest_status_label"] = norm
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/lead-labels", nil, body)
			if err != nil {
				return printError(cmd, "lead_labels.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "lead_labels.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Label name")
	cmd.Flags().StringVar(&interestStatus, "interest-status", "neutral", "Interest status: positive|neutral|negative (default neutral)")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newLeadLabelsUpdateCmd() *cobra.Command {
	var (
		name           string
		interestStatus string
		dataJSON       string
		dataFile       string
	)
	cmd := &cobra.Command{
		Use:   "update <label_id>",
		Short: "Update lead label (PATCH /lead-labels/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_labels.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_labels.update", fmt.Errorf("label_id is required"), nil)
			}

			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "lead_labels.update", err, nil)
			}
			// Instantly expects "label"; accept "name" as a convenient alias.
			if v, ok := body["label"]; !ok || v == nil {
				if v2, ok := body["name"]; ok && v2 != nil {
					body["label"] = v2
					delete(body, "name")
				}
			}
			if strings.TrimSpace(name) != "" {
				body["label"] = strings.TrimSpace(name)
			}
			if cmd.Flags().Changed("interest-status") {
				body["interest_status_label"] = interestStatus
			}
			if _, ok := body["interest_status_label"]; !ok || body["interest_status_label"] == nil {
				// Accept interest_status as a convenience alias.
				if v2, ok := body["interest_status"]; ok && v2 != nil {
					body["interest_status_label"] = v2
					delete(body, "interest_status")
				}
			}
			if raw, ok := body["interest_status_label"]; ok && raw != nil {
				s, ok := raw.(string)
				if !ok {
					return printError(cmd, "lead_labels.update", fmt.Errorf("interest_status_label must be a string"), nil)
				}
				norm, err := normalizeLeadLabelInterestStatusLabel(s)
				if err != nil {
					return printError(cmd, "lead_labels.update", err, nil)
				}
				if norm == "" {
					return printError(cmd, "lead_labels.update", fmt.Errorf("interest_status_label must be positive|neutral|negative"), nil)
				}
				body["interest_status_label"] = norm
			}
			if len(body) == 0 {
				return printError(cmd, "lead_labels.update", fmt.Errorf("no fields to update"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/lead-labels/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "lead_labels.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "lead_labels.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Label name")
	cmd.Flags().StringVar(&interestStatus, "interest-status", "", "Interest status: positive|neutral|negative")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newLeadLabelsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <label_id>",
		Short: "Delete lead label (DELETE /lead-labels/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "lead_labels.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "lead_labels.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "lead_labels.delete", fmt.Errorf("label_id is required"), nil)
			}
			// Instantly validates a JSON schema on this DELETE; include the id in the body.
			resp, meta, err := client.DeleteJSONWithBody(
				cmdContext(cmd),
				"/lead-labels/"+url.PathEscape(id),
				nil,
				map[string]any{"id": id},
			)
			if err != nil {
				return printError(cmd, "lead_labels.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "lead_labels.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}
