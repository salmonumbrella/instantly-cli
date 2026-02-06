package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newInboxPlacementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "inbox-placement",
		Aliases: []string{"deliverability", "ipt"},
		Short:   "Inbox placement tests and analytics",
	}

	cmd.AddCommand(newInboxPlacementTestsCmd())
	cmd.AddCommand(newInboxPlacementAnalyticsCmd())
	cmd.AddCommand(newInboxPlacementReportsCmd())

	return cmd
}

func newInboxPlacementTestsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "tests",
		Aliases: []string{"test"},
		Short:   "Inbox placement tests",
	}
	cmd.AddCommand(newInboxPlacementTestsListCmd())
	cmd.AddCommand(newInboxPlacementTestsGetCmd())
	cmd.AddCommand(newInboxPlacementTestsCreateCmd())
	cmd.AddCommand(newInboxPlacementTestsUpdateCmd())
	cmd.AddCommand(newInboxPlacementTestsDeleteCmd())
	cmd.AddCommand(newInboxPlacementTestsESPsCmd())
	return cmd
}

func newInboxPlacementTestsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inbox placement tests (GET /inbox-placement-tests)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.tests.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "inbox_placement.tests.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-tests", q)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.tests.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newInboxPlacementTestsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <test_id>",
		Short: "Get inbox placement test (GET /inbox-placement-tests/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.tests.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "inbox_placement.tests.get", fmt.Errorf("test_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-tests/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.tests.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newInboxPlacementTestsCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create inbox placement test (POST /inbox-placement-tests, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "inbox_placement.tests.create", fmt.Errorf("refusing to create test without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.tests.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "inbox_placement.tests.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/inbox-placement-tests", nil, body)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "inbox_placement.tests.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm creating a test")
	return cmd
}

func newInboxPlacementTestsUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update <test_id>",
		Short: "Update inbox placement test (PATCH /inbox-placement-tests/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.tests.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "inbox_placement.tests.update", fmt.Errorf("test_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "inbox_placement.tests.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/inbox-placement-tests/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "inbox_placement.tests.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newInboxPlacementTestsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <test_id>",
		Short: "Delete inbox placement test (DELETE /inbox-placement-tests/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "inbox_placement.tests.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.tests.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "inbox_placement.tests.delete", fmt.Errorf("test_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/inbox-placement-tests/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.tests.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newInboxPlacementTestsESPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "esps",
		Aliases: []string{"email-service-provider-options", "esp-options"},
		Short:   "List email service provider options (GET /inbox-placement-tests/email-service-provider-options)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.tests.esps", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-tests/email-service-provider-options", nil)
			if err != nil {
				return printError(cmd, "inbox_placement.tests.esps", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.tests.esps", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newInboxPlacementAnalyticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "analytics",
		Short:     "Inbox placement analytics",
		ValidArgs: []string{"stats-by-test-id", "deliverability-insights", "stats-by-date"},
	}
	cmd.AddCommand(newInboxPlacementAnalyticsListCmd())
	cmd.AddCommand(newInboxPlacementAnalyticsGetCmd())
	cmd.AddCommand(newInboxPlacementAnalyticsStatsByTestIDCmd())
	cmd.AddCommand(newInboxPlacementAnalyticsDeliverabilityInsightsCmd())
	cmd.AddCommand(newInboxPlacementAnalyticsStatsByDateCmd())
	return cmd
}

func newInboxPlacementAnalyticsListCmd() *cobra.Command {
	var (
		testID        string
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inbox placement analytics (GET /inbox-placement-analytics; requires test_id)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.list", err, nil)
			}
			q := url.Values{}
			if strings.TrimSpace(testID) != "" {
				q.Set("test_id", strings.TrimSpace(testID))
			}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "inbox_placement.analytics.list", err, nil)
			}
			if strings.TrimSpace(q.Get("test_id")) == "" {
				return printError(cmd, "inbox_placement.analytics.list", fmt.Errorf("missing required test_id (set --test-id or pass --query test_id=...)"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-analytics", q)
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.analytics.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringVar(&testID, "test-id", "", "Inbox placement test ID (required unless provided via --query test_id=...)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newInboxPlacementAnalyticsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <analytics_id>",
		Short: "Get inbox placement analytics record (GET /inbox-placement-analytics/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "inbox_placement.analytics.get", fmt.Errorf("analytics_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-analytics/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.analytics.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newInboxPlacementAnalyticsStatsByTestIDCmd() *cobra.Command {
	var (
		testID   string
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "stats-by-test-id",
		Short: "Stats by test ID (POST /inbox-placement-analytics/stats-by-test-id)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.stats_by_test_id", err, nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(dataJSON) != "" || strings.TrimSpace(dataFile) != "" {
				m, err := readJSONObjectInput(dataJSON, dataFile)
				if err != nil {
					return printError(cmd, "inbox_placement.analytics.stats_by_test_id", err, nil)
				}
				mergeMaps(body, m)
			}
			if strings.TrimSpace(testID) != "" && body["test_id"] == nil {
				body["test_id"] = strings.TrimSpace(testID)
			}
			if strings.TrimSpace(fmt.Sprint(body["test_id"])) == "" || body["test_id"] == nil {
				return printError(cmd, "inbox_placement.analytics.stats_by_test_id", fmt.Errorf("missing required test_id (set --test-id or set test_id in --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/inbox-placement-analytics/stats-by-test-id", nil, body)
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.stats_by_test_id", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.analytics.stats_by_test_id", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringVar(&testID, "test-id", "", "Inbox placement test ID (or provide test_id in the JSON body)")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with --test-id)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged with --test-id)")
	return cmd
}

func newInboxPlacementAnalyticsDeliverabilityInsightsCmd() *cobra.Command {
	var (
		testID   string
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "deliverability-insights",
		Short: "Deliverability insights (POST /inbox-placement-analytics/deliverability-insights)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.deliverability_insights", err, nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(dataJSON) != "" || strings.TrimSpace(dataFile) != "" {
				m, err := readJSONObjectInput(dataJSON, dataFile)
				if err != nil {
					return printError(cmd, "inbox_placement.analytics.deliverability_insights", err, nil)
				}
				mergeMaps(body, m)
			}
			if strings.TrimSpace(testID) != "" && body["test_id"] == nil {
				body["test_id"] = strings.TrimSpace(testID)
			}
			if strings.TrimSpace(fmt.Sprint(body["test_id"])) == "" || body["test_id"] == nil {
				return printError(cmd, "inbox_placement.analytics.deliverability_insights", fmt.Errorf("missing required test_id (set --test-id or set test_id in --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/inbox-placement-analytics/deliverability-insights", nil, body)
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.deliverability_insights", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.analytics.deliverability_insights", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringVar(&testID, "test-id", "", "Inbox placement test ID (or provide test_id in the JSON body)")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with --test-id)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged with --test-id)")
	return cmd
}

func newInboxPlacementAnalyticsStatsByDateCmd() *cobra.Command {
	var (
		testID    string
		startDate string
		endDate   string
		dataJSON  string
		dataFile  string
	)
	cmd := &cobra.Command{
		Use:   "stats-by-date",
		Short: "Stats by date (POST /inbox-placement-analytics/stats-by-date)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.stats_by_date", err, nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(dataJSON) != "" || strings.TrimSpace(dataFile) != "" {
				m, err := readJSONObjectInput(dataJSON, dataFile)
				if err != nil {
					return printError(cmd, "inbox_placement.analytics.stats_by_date", err, nil)
				}
				mergeMaps(body, m)
			}
			if strings.TrimSpace(testID) != "" && body["test_id"] == nil {
				body["test_id"] = strings.TrimSpace(testID)
			}
			if strings.TrimSpace(startDate) != "" && body["start_date"] == nil {
				body["start_date"] = strings.TrimSpace(startDate)
			}
			if strings.TrimSpace(endDate) != "" && body["end_date"] == nil {
				body["end_date"] = strings.TrimSpace(endDate)
			}
			if strings.TrimSpace(fmt.Sprint(body["test_id"])) == "" || body["test_id"] == nil {
				return printError(cmd, "inbox_placement.analytics.stats_by_date", fmt.Errorf("missing required test_id (set --test-id or set test_id in --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/inbox-placement-analytics/stats-by-date", nil, body)
			if err != nil {
				return printError(cmd, "inbox_placement.analytics.stats_by_date", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.analytics.stats_by_date", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringVar(&testID, "test-id", "", "Inbox placement test ID (or provide test_id in the JSON body)")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD) (sets start_date in JSON body)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD) (sets end_date in JSON body)")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newInboxPlacementReportsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "reports",
		Short:     "Inbox placement reports",
		ValidArgs: []string{"list", "get"},
	}
	cmd.AddCommand(newInboxPlacementReportsListCmd())
	cmd.AddCommand(newInboxPlacementReportsGetCmd())
	return cmd
}

func newInboxPlacementReportsListCmd() *cobra.Command {
	var (
		testID        string
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List inbox placement reports (GET /inbox-placement-reports; requires test_id)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.reports.list", err, nil)
			}
			q := url.Values{}
			if strings.TrimSpace(testID) != "" {
				q.Set("test_id", strings.TrimSpace(testID))
			}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "inbox_placement.reports.list", err, nil)
			}
			if strings.TrimSpace(q.Get("test_id")) == "" {
				return printError(cmd, "inbox_placement.reports.list", fmt.Errorf("missing required test_id (set --test-id or pass --query test_id=...)"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-reports", q)
			if err != nil {
				return printError(cmd, "inbox_placement.reports.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.reports.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringVar(&testID, "test-id", "", "Inbox placement test ID (required unless provided via --query test_id=...)")
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newInboxPlacementReportsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <report_id>",
		Short: "Get inbox placement report (GET /inbox-placement-reports/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "inbox_placement.reports.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "inbox_placement.reports.get", fmt.Errorf("report_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/inbox-placement-reports/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "inbox_placement.reports.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "inbox_placement.reports.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}
