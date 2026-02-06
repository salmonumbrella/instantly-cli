package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newWebhooksCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "webhooks",
		Aliases: []string{"webhook"},
		Short:   "Manage webhooks and webhook events",
	}

	cmd.AddCommand(newWebhooksListCmd())
	cmd.AddCommand(newWebhooksGetCmd())
	cmd.AddCommand(newWebhooksCreateCmd())
	cmd.AddCommand(newWebhooksUpdateCmd())
	cmd.AddCommand(newWebhooksDeleteCmd())
	cmd.AddCommand(newWebhooksEventTypesCmd())
	cmd.AddCommand(newWebhooksTestCmd())
	cmd.AddCommand(newWebhooksResumeCmd())

	cmd.AddCommand(newWebhookEventsCmd())

	return cmd
}

func newWebhooksListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		campaign      string
		eventType     string
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhooks (GET /webhooks)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.list", err, nil)
			}

			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if strings.TrimSpace(campaign) != "" {
				q.Set("campaign", strings.TrimSpace(campaign))
			}
			if strings.TrimSpace(eventType) != "" {
				q.Set("event_type", strings.TrimSpace(eventType))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "webhooks.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhooks", q)
			if err != nil {
				return printError(cmd, "webhooks.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhooks.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringVar(&campaign, "campaign", "", "Filter by campaign ID (UUID)")
	cmd.Flags().StringVar(&eventType, "event-type", "", "Filter by event type (e.g. all_events, email_sent)")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newWebhooksGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <webhook_id>",
		Short: "Get webhook (GET /webhooks/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "webhooks.get", fmt.Errorf("webhook_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhooks/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "webhooks.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhooks.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWebhooksCreateCmd() *cobra.Command {
	var (
		targetURL           string
		campaign            string
		name                string
		eventType           string
		customInterestValue float64
		customInterestSet   bool
		headersJSON         string
		dataJSON            string
		dataFile            string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create webhook (POST /webhooks)",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("custom-interest-value") {
				customInterestSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.create", err, nil)
			}

			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "webhooks.create", err, nil)
			}

			if strings.TrimSpace(targetURL) != "" {
				body["target_hook_url"] = strings.TrimSpace(targetURL)
			}
			if strings.TrimSpace(campaign) != "" {
				body["campaign"] = strings.TrimSpace(campaign)
			}
			if strings.TrimSpace(name) != "" {
				body["name"] = strings.TrimSpace(name)
			}
			if strings.TrimSpace(eventType) != "" {
				body["event_type"] = strings.TrimSpace(eventType)
			}
			if customInterestSet {
				body["custom_interest_value"] = customInterestValue
			}
			if strings.TrimSpace(headersJSON) != "" {
				var v any
				if err := json.Unmarshal([]byte(headersJSON), &v); err != nil {
					return printError(cmd, "webhooks.create", fmt.Errorf("invalid --headers-json: %w", err), nil)
				}
				m, ok := v.(map[string]any)
				if !ok {
					return printError(cmd, "webhooks.create", fmt.Errorf("--headers-json must be a JSON object"), nil)
				}
				body["headers"] = m
			}

			target, _ := body["target_hook_url"].(string)
			if strings.TrimSpace(target) == "" {
				return printError(cmd, "webhooks.create", fmt.Errorf("--target-url is required (or set target_hook_url in --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/webhooks", nil, body)
			if err != nil {
				return printError(cmd, "webhooks.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "webhooks.create", resp, metaFrom(meta, resp), body)
		},
	}

	cmd.Flags().StringVar(&targetURL, "target-url", "", "Target webhook URL (maps to target_hook_url)")
	cmd.Flags().StringVar(&campaign, "campaign", "", "Optional campaign ID (UUID)")
	cmd.Flags().StringVar(&name, "name", "", "Optional webhook name")
	cmd.Flags().StringVar(&eventType, "event-type", "all_events", "Event type (default all_events)")
	cmd.Flags().Float64Var(&customInterestValue, "custom-interest-value", 0, "Custom interest value (for custom label events)")
	cmd.Flags().StringVar(&headersJSON, "headers-json", "", "Optional headers JSON object to include on delivery")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON body file, or '-' for stdin (merged with flags)")
	return cmd
}

func newWebhooksUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)

	cmd := &cobra.Command{
		Use:   "update <webhook_id>",
		Short: "Update webhook (PATCH /webhooks/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "webhooks.update", fmt.Errorf("webhook_id is required"), nil)
			}

			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "webhooks.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "webhooks.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/webhooks/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "webhooks.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "webhooks.update", resp, metaFrom(meta, resp), body)
		},
	}

	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newWebhooksDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <webhook_id>",
		Short: "Delete webhook (DELETE /webhooks/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "webhooks.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "webhooks.delete", fmt.Errorf("webhook_id is required"), nil)
			}

			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/webhooks/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "webhooks.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhooks.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newWebhooksEventTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event-types",
		Short: "List available webhook event types (GET /webhooks/event-types)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.event_types", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhooks/event-types", nil)
			if err != nil {
				return printError(cmd, "webhooks.event_types", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhooks.event_types", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWebhooksTestCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "test <webhook_id>",
		Short: "Send a test payload to a webhook (POST /webhooks/{id}/test, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "webhooks.test", fmt.Errorf("refusing to send test payload without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.test", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "webhooks.test", fmt.Errorf("webhook_id is required"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/webhooks/"+url.PathEscape(id)+"/test", nil, nil)
			if err != nil {
				return printError(cmd, "webhooks.test", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhooks.test", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm sending a test delivery")
	return cmd
}

func newWebhooksResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <webhook_id>",
		Short: "Resume a disabled webhook (POST /webhooks/{id}/resume)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhooks.resume", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "webhooks.resume", fmt.Errorf("webhook_id is required"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/webhooks/"+url.PathEscape(id)+"/resume", nil, nil)
			if err != nil {
				return printError(cmd, "webhooks.resume", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhooks.resume", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWebhookEventsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "events",
		Short: "Inspect webhook delivery events",
	}

	cmd.AddCommand(newWebhookEventsListCmd())
	cmd.AddCommand(newWebhookEventsGetCmd())
	cmd.AddCommand(newWebhookEventsSummaryCmd())
	cmd.AddCommand(newWebhookEventsSummaryByDateCmd())

	return cmd
}

func newWebhookEventsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List webhook events (GET /webhook-events)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhook_events.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "webhook_events.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhook-events", q)
			if err != nil {
				return printError(cmd, "webhook_events.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhook_events.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newWebhookEventsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <event_id>",
		Short: "Get webhook event (GET /webhook-events/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhook_events.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "webhook_events.get", fmt.Errorf("event_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhook-events/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "webhook_events.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhook_events.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWebhookEventsSummaryCmd() *cobra.Command {
	var (
		queryPairs []string
	)
	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Summary of webhook events (GET /webhook-events/summary)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhook_events.summary", err, nil)
			}
			q := url.Values{}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "webhook_events.summary", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhook-events/summary", q)
			if err != nil {
				return printError(cmd, "webhook_events.summary", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhook_events.summary", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Query param (repeatable): key=value")
	return cmd
}

func newWebhookEventsSummaryByDateCmd() *cobra.Command {
	var (
		queryPairs []string
	)
	cmd := &cobra.Command{
		Use:   "summary-by-date",
		Short: "Summary of webhook events by date (GET /webhook-events/summary-by-date)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "webhook_events.summary_by_date", err, nil)
			}
			q := url.Values{}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "webhook_events.summary_by_date", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/webhook-events/summary-by-date", q)
			if err != nil {
				return printError(cmd, "webhook_events.summary_by_date", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "webhook_events.summary_by_date", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Query param (repeatable): key=value")
	return cmd
}
