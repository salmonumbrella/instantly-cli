package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newSubsequencesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "subsequences",
		Aliases: []string{"subsequence"},
		Short:   "Manage campaign subsequences",
	}

	cmd.AddCommand(newSubsequencesListCmd())
	cmd.AddCommand(newSubsequencesGetCmd())
	cmd.AddCommand(newSubsequencesCreateCmd())
	cmd.AddCommand(newSubsequencesUpdateCmd())
	cmd.AddCommand(newSubsequencesDeleteCmd())
	cmd.AddCommand(newSubsequencesPauseCmd())
	cmd.AddCommand(newSubsequencesResumeCmd())
	cmd.AddCommand(newSubsequencesDuplicateCmd())

	return cmd
}

func newSubsequencesListCmd() *cobra.Command {
	var (
		parentCampaign string
		limit          int
		startingAfter  string
		search         string
		queryPairs     []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subsequences (GET /subsequences, requires --parent-campaign)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.list", err, nil)
			}
			if strings.TrimSpace(parentCampaign) == "" {
				return printError(cmd, "subsequences.list", fmt.Errorf("--parent-campaign is required"), nil)
			}
			q := url.Values{}
			q.Set("parent_campaign", strings.TrimSpace(parentCampaign))
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
				return printError(cmd, "subsequences.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/subsequences", q)
			if err != nil {
				return printError(cmd, "subsequences.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "subsequences.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringVar(&parentCampaign, "parent-campaign", "", "Parent campaign ID (UUID)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringVar(&search, "search", "", "Search")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newSubsequencesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <subsequence_id>",
		Short: "Get subsequence (GET /subsequences/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "subsequences.get", fmt.Errorf("subsequence_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/subsequences/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "subsequences.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "subsequences.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newSubsequencesCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create subsequence (POST /subsequences)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "subsequences.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "subsequences.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/subsequences", nil, body)
			if err != nil {
				return printError(cmd, "subsequences.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "subsequences.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	return cmd
}

func newSubsequencesUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update <subsequence_id>",
		Short: "Update subsequence (PATCH /subsequences/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "subsequences.update", fmt.Errorf("subsequence_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "subsequences.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "subsequences.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/subsequences/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "subsequences.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "subsequences.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newSubsequencesDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <subsequence_id>",
		Short: "Delete subsequence (DELETE /subsequences/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "subsequences.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "subsequences.delete", fmt.Errorf("subsequence_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/subsequences/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "subsequences.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "subsequences.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newSubsequencesPauseCmd() *cobra.Command {
	var (
		confirm  bool
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "pause <subsequence_id>",
		Short: "Pause subsequence (POST /subsequences/{id}/pause, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "subsequences.pause", fmt.Errorf("refusing to pause without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.pause", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "subsequences.pause", fmt.Errorf("subsequence_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "subsequences.pause", err, nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/subsequences/"+url.PathEscape(id)+"/pause", nil, body)
			if err != nil {
				return printError(cmd, "subsequences.pause", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "subsequences.pause", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm pausing a subsequence")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged)")
	return cmd
}

func newSubsequencesResumeCmd() *cobra.Command {
	var (
		confirm  bool
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "resume <subsequence_id>",
		Short: "Resume subsequence (POST /subsequences/{id}/resume, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "subsequences.resume", fmt.Errorf("refusing to resume without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.resume", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "subsequences.resume", fmt.Errorf("subsequence_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "subsequences.resume", err, nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/subsequences/"+url.PathEscape(id)+"/resume", nil, body)
			if err != nil {
				return printError(cmd, "subsequences.resume", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "subsequences.resume", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm resuming a subsequence")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged)")
	return cmd
}

func newSubsequencesDuplicateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "duplicate <subsequence_id>",
		Short: "Duplicate subsequence (POST /subsequences/{id}/duplicate, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "subsequences.duplicate", fmt.Errorf("refusing to duplicate without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "subsequences.duplicate", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "subsequences.duplicate", fmt.Errorf("subsequence_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "subsequences.duplicate", err, nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/subsequences/"+url.PathEscape(id)+"/duplicate", nil, body)
			if err != nil {
				return printError(cmd, "subsequences.duplicate", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "subsequences.duplicate", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm duplicating a subsequence")
	return cmd
}
