package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newAnalyticsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "analytics",
		Aliases: []string{"stats"},
		Short:   "Analytics and reporting",
	}

	cmd.AddCommand(newAnalyticsCampaignCmd())
	cmd.AddCommand(newAnalyticsDailyCmd())
	cmd.AddCommand(newAnalyticsWarmupCmd())

	return cmd
}

func newAnalyticsCampaignCmd() *cobra.Command {
	var (
		campaignID        string
		startDate         string
		endDate           string
		excludeTotalLeads bool
		excludeSet        bool
	)

	cmd := &cobra.Command{
		Use:   "campaign",
		Short: "Get campaign analytics",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("exclude-total-leads-count") {
				excludeSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "analytics.campaign", err, nil)
			}

			q := url.Values{}
			if campaignID != "" {
				q.Set("id", campaignID)
			}
			if startDate != "" {
				q.Set("start_date", startDate)
			}
			if endDate != "" {
				q.Set("end_date", endDate)
			}
			if excludeSet {
				q.Set("exclude_total_leads_count", fmt.Sprintf("%v", excludeTotalLeads))
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns/analytics", q)
			if err != nil {
				return printError(cmd, "analytics.campaign", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "analytics.campaign", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&campaignID, "campaign-id", "", "Campaign ID (omit for all campaigns)")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().BoolVar(&excludeTotalLeads, "exclude-total-leads-count", false, "Exclude total leads count for faster response")

	return cmd
}

func newAnalyticsDailyCmd() *cobra.Command {
	var (
		campaignID     string
		startDate      string
		endDate        string
		campaignStatus int
		statusSet      bool
	)

	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Get daily campaign analytics",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("campaign-status") {
				statusSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "analytics.daily", err, nil)
			}

			q := url.Values{}
			if campaignID != "" {
				q.Set("campaign_id", campaignID)
			}
			if startDate != "" {
				q.Set("start_date", startDate)
			}
			if endDate != "" {
				q.Set("end_date", endDate)
			}
			if statusSet {
				q.Set("campaign_status", fmt.Sprintf("%d", campaignStatus))
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/campaigns/analytics/daily", q)
			if err != nil {
				return printError(cmd, "analytics.daily", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "analytics.daily", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&campaignID, "campaign-id", "", "Campaign ID (omit for all campaigns)")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")
	cmd.Flags().IntVar(&campaignStatus, "campaign-status", 0, "Campaign status filter (-99 for all)")

	return cmd
}

func newAnalyticsWarmupCmd() *cobra.Command {
	var (
		email     string
		emails    string
		startDate string
		endDate   string
	)

	cmd := &cobra.Command{
		Use:   "warmup",
		Short: "Get warmup analytics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "analytics.warmup", err, nil)
			}

			var emailList []string
			if strings.TrimSpace(emails) != "" {
				for _, p := range strings.Split(emails, ",") {
					p = strings.TrimSpace(p)
					if p != "" {
						emailList = append(emailList, p)
					}
				}
			} else if strings.TrimSpace(email) != "" {
				emailList = []string{strings.TrimSpace(email)}
			}

			body := map[string]any{}
			if len(emailList) > 0 {
				body["emails"] = emailList
			}
			if startDate != "" {
				body["start_date"] = startDate
			}
			if endDate != "" {
				body["end_date"] = endDate
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/accounts/warmup-analytics", nil, body)
			if err != nil {
				return printError(cmd, "analytics.warmup", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "analytics.warmup", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Single account email")
	cmd.Flags().StringVar(&emails, "emails", "", "Comma-separated list of account emails")
	cmd.Flags().StringVar(&startDate, "start-date", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end-date", "", "End date (YYYY-MM-DD)")

	return cmd
}
