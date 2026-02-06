package cmd

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func newCampaignsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "campaigns",
		Aliases: []string{"campaign", "c"},
		Short:   "Manage campaigns",
	}

	cmd.AddCommand(newCampaignsListCmd())
	cmd.AddCommand(newCampaignsGetCmd())
	cmd.AddCommand(newCampaignsCreateCmd())
	cmd.AddCommand(newCampaignsUpdateCmd())
	cmd.AddCommand(newCampaignsActivateCmd())
	cmd.AddCommand(newCampaignsPauseCmd())
	cmd.AddCommand(newCampaignsDeleteCmd())
	cmd.AddCommand(newCampaignsSearchByContactCmd())
	cmd.AddCommand(newCampaignsAnalyticsOverviewCmd())
	cmd.AddCommand(newCampaignsAnalyticsStepsCmd())

	return cmd
}

func newCampaignsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List campaigns",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.list", err, nil)
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
				return printError(cmd, "campaigns.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns", q)
			if err != nil {
				return printError(cmd, "campaigns.list", err, metaFrom(meta, nil))
			}

			return printResult(cmd, "campaigns.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination (use response next_starting_after)")
	cmd.Flags().StringVar(&search, "search", "", "Search by campaign name")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newCampaignsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <campaign_id>",
		Aliases: []string{"show"},
		Short:   "Get campaign by ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "campaigns.get", fmt.Errorf("campaign_id is required"), nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "campaigns.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newCampaignsCreateCmd() *cobra.Command {
	var (
		name       string
		subject    string
		body       string
		senders    string
		sendersMax int
		dailyLimit int
		emailGap   int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a campaign (agent-friendly defaults)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.create", err, nil)
			}
			if strings.TrimSpace(name) == "" {
				return printError(cmd, "campaigns.create", fmt.Errorf("--name is required"), nil)
			}
			if strings.TrimSpace(subject) == "" {
				return printError(cmd, "campaigns.create", fmt.Errorf("--subject is required"), nil)
			}
			if strings.TrimSpace(body) == "" {
				return printError(cmd, "campaigns.create", fmt.Errorf("--body is required"), nil)
			}

			var emailList []string
			switch strings.TrimSpace(strings.ToLower(senders)) {
			case "", "auto":
				if client.DryRun {
					// In real mode, we would discover eligible senders via GET /accounts.
					// For --dry-run, keep the shape correct without performing additional calls.
					emailList = []string{"<auto>"}
					break
				}
				// Auto-pick eligible senders by listing accounts and filtering like the MCP tool.
				accountsResp, _, err := client.GetJSON(cmdContext(cmd), "/accounts", url.Values{"limit": []string{"100"}})
				if err != nil {
					return printError(cmd, "campaigns.create", fmt.Errorf("auto sender discovery failed: %w", err), nil)
				}
				m, ok := accountsResp.(map[string]any)
				if !ok {
					return printError(cmd, "campaigns.create", fmt.Errorf("unexpected /accounts response shape"), nil)
				}
				items, _ := m["items"].([]any)
				for _, it := range items {
					acc, _ := it.(map[string]any)
					if acc == nil {
						continue
					}
					// status == 1, setup_pending == false, warmup_status == 1
					status, _ := acc["status"].(float64)
					setupPending, _ := acc["setup_pending"].(bool)
					warmupStatus, _ := acc["warmup_status"].(float64)
					email, _ := acc["email"].(string)
					if int(status) == 1 && !setupPending && int(warmupStatus) == 1 && email != "" {
						emailList = append(emailList, email)
					}
					if sendersMax > 0 && len(emailList) >= sendersMax {
						break
					}
				}
				if len(emailList) == 0 {
					return printError(cmd, "campaigns.create", fmt.Errorf("no eligible sender accounts found (need active + setup complete + warmup complete)"), nil)
				}
			default:
				parts := strings.Split(senders, ",")
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						emailList = append(emailList, p)
					}
				}
				if len(emailList) == 0 {
					return printError(cmd, "campaigns.create", fmt.Errorf("--senders must be 'auto' or a comma-separated list of emails"), nil)
				}
			}

			payload := buildCreateCampaignPayload(name, subject, body, emailList, dailyLimit, emailGap)

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/campaigns", nil, payload)
			if err != nil {
				return printError(cmd, "campaigns.create", err, metaFrom(meta, nil))
			}
			// Always include the request payload used so agents can replay / debug.
			outMeta := map[string]any{
				"payload_used": payload,
			}
			if m := metaFrom(meta, resp); m != nil {
				for k, v := range m {
					outMeta[k] = v
				}
			}
			return printResult(cmd, "campaigns.create", resp, outMeta)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Campaign name")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&body, "body", "", "Email body (plain text; converted to HTML)")
	cmd.Flags().StringVar(&senders, "senders", "auto", "Sender accounts: 'auto' or comma-separated emails")
	cmd.Flags().IntVar(&sendersMax, "senders-max", 1, "When --senders=auto, pick up to N eligible senders")
	cmd.Flags().IntVar(&dailyLimit, "daily-limit", 30, "Daily send limit")
	cmd.Flags().IntVar(&emailGap, "email-gap", 10, "Gap between emails (minutes)")

	return cmd
}

func buildCreateCampaignPayload(name, subject, body string, emailList []string, dailyLimit, emailGap int) map[string]any {
	subject = regexp.MustCompile(`[\r\n]+`).ReplaceAllString(subject, " ")
	subject = strings.TrimSpace(subject)

	htmlBody := convertLineBreaksToHTML(body)

	return map[string]any{
		"name": name,
		"sequences": []any{
			map[string]any{
				"steps": []any{
					map[string]any{
						"type":  "email",
						"delay": 0,
						"variants": []any{
							map[string]any{
								"subject": subject,
								"body":    htmlBody,
							},
						},
					},
				},
			},
		},
		"email_list":         emailList,
		"open_tracking":      false,
		"link_tracking":      false,
		"daily_limit":        dailyLimit,
		"email_gap":          emailGap,
		"stop_on_reply":      true,
		"stop_on_auto_reply": true,
		"campaign_schedule": map[string]any{
			"schedules": []any{
				map[string]any{
					"name":     "Default Schedule",
					"timezone": "America/New_York",
					"timing": map[string]any{
						"from": "09:00",
						"to":   "17:00",
					},
					"days": map[string]any{
						"0": false,
						"1": true,
						"2": true,
						"3": true,
						"4": true,
						"5": true,
						"6": false,
					},
				},
			},
		},
	}
}

func convertLineBreaksToHTML(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	parts := strings.Split(text, "\n\n")
	var out strings.Builder
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		p = strings.ReplaceAll(p, "\n", "<br />")
		out.WriteString("<p>")
		out.WriteString(p)
		out.WriteString("</p>")
	}
	return out.String()
}

func newCampaignsUpdateCmd() *cobra.Command {
	var (
		name            string
		dailyLimit      int
		dailyLimitSet   bool
		emailGap        int
		emailGapSet     bool
		openTracking    bool
		openTrackingSet bool
		linkTracking    bool
		linkTrackingSet bool
	)

	cmd := &cobra.Command{
		Use:   "update <campaign_id>",
		Short: "Update campaign settings (partial)",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("daily-limit") {
				dailyLimitSet = true
			}
			if cmd.Flags().Changed("email-gap") {
				emailGapSet = true
			}
			if cmd.Flags().Changed("open-tracking") {
				openTrackingSet = true
			}
			if cmd.Flags().Changed("link-tracking") {
				linkTrackingSet = true
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "campaigns.update", fmt.Errorf("campaign_id is required"), nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(name) != "" {
				body["name"] = name
			}
			if dailyLimitSet {
				body["daily_limit"] = dailyLimit
			}
			if emailGapSet {
				body["email_gap"] = emailGap
			}
			if openTrackingSet {
				body["open_tracking"] = openTracking
			}
			if linkTrackingSet {
				body["link_tracking"] = linkTracking
			}

			if len(body) == 0 {
				return printError(cmd, "campaigns.update", fmt.Errorf("no fields to update"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/campaigns/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "campaigns.update", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns.update", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Campaign name")
	cmd.Flags().IntVar(&dailyLimit, "daily-limit", 0, "Daily send limit")
	cmd.Flags().IntVar(&emailGap, "email-gap", 0, "Gap between emails (minutes)")
	cmd.Flags().BoolVar(&openTracking, "open-tracking", false, "Enable open tracking")
	cmd.Flags().BoolVar(&linkTracking, "link-tracking", false, "Enable link tracking")

	return cmd
}

func newCampaignsActivateCmd() *cobra.Command {
	return campaignActionCmd("activate", "Activate campaign", "/campaigns/%s/activate")
}

func newCampaignsPauseCmd() *cobra.Command {
	return campaignActionCmd("pause", "Pause campaign", "/campaigns/%s/pause")
}

func campaignActionCmd(use, short, endpointFmt string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use + " <campaign_id>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns."+use, err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "campaigns."+use, fmt.Errorf("campaign_id is required"), nil)
			}
			path := fmt.Sprintf(endpointFmt, url.PathEscape(id))
			resp, meta, err := client.PostJSON(cmdContext(cmd), path, nil, nil)
			if err != nil {
				return printError(cmd, "campaigns."+use, err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns."+use, resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newCampaignsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <campaign_id>",
		Short: "Delete a campaign (requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "campaigns.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "campaigns.delete", fmt.Errorf("campaign_id is required"), nil)
			}

			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/campaigns/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "campaigns.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newCampaignsSearchByContactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "search-by-contact <contact_email>",
		Short: "Find campaigns a contact is enrolled in",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.search_by_contact", err, nil)
			}
			email := strings.TrimSpace(args[0])
			if email == "" {
				return printError(cmd, "campaigns.search_by_contact", fmt.Errorf("contact_email is required"), nil)
			}
			q := url.Values{"contact_email": []string{email}}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns/search-by-contact", q)
			if err != nil {
				return printError(cmd, "campaigns.search_by_contact", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns.search_by_contact", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newCampaignsAnalyticsOverviewCmd() *cobra.Command {
	var queryPairs []string
	cmd := &cobra.Command{
		Use:   "analytics-overview",
		Short: "Campaign analytics overview (GET /campaigns/analytics/overview)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.analytics_overview", err, nil)
			}
			q := url.Values{}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "campaigns.analytics_overview", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns/analytics/overview", q)
			if err != nil {
				return printError(cmd, "campaigns.analytics_overview", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns.analytics_overview", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Query param (repeatable): key=value")
	return cmd
}

func newCampaignsAnalyticsStepsCmd() *cobra.Command {
	var queryPairs []string
	cmd := &cobra.Command{
		Use:   "analytics-steps",
		Short: "Campaign analytics steps (GET /campaigns/analytics/steps)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "campaigns.analytics_steps", err, nil)
			}
			q := url.Values{}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "campaigns.analytics_steps", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns/analytics/steps", q)
			if err != nil {
				return printError(cmd, "campaigns.analytics_steps", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "campaigns.analytics_steps", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Query param (repeatable): key=value")
	return cmd
}
