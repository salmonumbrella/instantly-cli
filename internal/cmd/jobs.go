package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newJobsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "jobs",
		Aliases: []string{"background-jobs", "bg"},
		Short:   "Track background jobs",
	}

	cmd.AddCommand(newJobsListCmd())
	cmd.AddCommand(newJobsGetCmd())

	return cmd
}

func newJobsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List background jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "jobs.list", err, nil)
			}

			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprintf("%d", limit))
			}
			if startingAfter != "" {
				q.Set("starting_after", startingAfter)
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "jobs.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/background-jobs", q)
			if err != nil {
				return printError(cmd, "jobs.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "jobs.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination (use response next_starting_after)")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")

	return cmd
}

func newJobsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <job_id>",
		Aliases: []string{"show"},
		Short:   "Get background job by ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "jobs.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "jobs.get", fmt.Errorf("job_id is required"), nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/background-jobs/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "jobs.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "jobs.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}
