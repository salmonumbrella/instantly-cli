package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "accounts",
		Aliases: []string{"account", "a"},
		Short:   "Manage sender accounts",
	}

	cmd.AddCommand(newAccountsListCmd())
	cmd.AddCommand(newAccountsGetCmd())
	cmd.AddCommand(newAccountsCreateCmd())
	cmd.AddCommand(newAccountsUpdateCmd())
	cmd.AddCommand(newAccountsWarmupEnableCmd())
	cmd.AddCommand(newAccountsWarmupDisableCmd())
	cmd.AddCommand(newAccountsTestVitalsCmd())
	cmd.AddCommand(newAccountsDeleteCmd())
	cmd.AddCommand(newAccountsAnalyticsDailyCmd())

	return cmd
}

func newAccountsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		status        int
		provider      int
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List accounts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts.list", err, nil)
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
			if cmd.Flags().Changed("status") {
				q.Set("status", fmt.Sprintf("%d", status))
			}
			if cmd.Flags().Changed("provider") {
				q.Set("provider_code", fmt.Sprintf("%d", provider))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "accounts.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/accounts", q)
			if err != nil {
				return printError(cmd, "accounts.list", err, metaFrom(meta, nil))
			}

			return printResult(cmd, "accounts.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination (use response next_starting_after)")
	cmd.Flags().StringVar(&search, "search", "", "Search accounts")
	cmd.Flags().IntVar(&status, "status", 0, "Filter by status (1 active, 2 paused, negative error codes)")
	cmd.Flags().Lookup("status").NoOptDefVal = "1"
	cmd.Flags().IntVar(&provider, "provider", 0, "Filter by provider_code (1 IMAP, 2 Google, 3 Microsoft, 4 AWS)")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")

	return cmd
}

func newAccountsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <email>",
		Aliases: []string{"show"},
		Short:   "Get account by email",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts.get", err, nil)
			}

			email := strings.TrimSpace(args[0])
			if email == "" {
				return printError(cmd, "accounts.get", fmt.Errorf("email is required"), nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/accounts/"+url.PathEscape(email), nil)
			if err != nil {
				return printError(cmd, "accounts.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "accounts.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newAccountsCreateCmd() *cobra.Command {
	var (
		email        string
		firstName    string
		lastName     string
		providerCode int

		imapHost     string
		imapPort     int
		imapUsername string
		imapPassword string

		smtpHost     string
		smtpPort     int
		smtpUsername string
		smtpPassword string

		dataJSON string
		dataFile string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an account (POST /accounts)",
		Long: strings.TrimSpace(`
Create an email account. This usually requires valid IMAP/SMTP credentials.

You can either:
- supply a full JSON body via --data-json/--data-file, and optionally override fields via flags, or
- use flags only for the common fields.
`),
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts.create", err, nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(dataJSON) != "" || strings.TrimSpace(dataFile) != "" {
				m, err := readJSONObjectInput(dataJSON, dataFile)
				if err != nil {
					return printError(cmd, "accounts.create", err, nil)
				}
				mergeMaps(body, m)
			}

			if strings.TrimSpace(email) != "" {
				body["email"] = strings.TrimSpace(email)
			}
			if cmd.Flags().Changed("first-name") {
				body["first_name"] = firstName
			}
			if cmd.Flags().Changed("last-name") {
				body["last_name"] = lastName
			}
			if cmd.Flags().Changed("provider-code") {
				body["provider_code"] = providerCode
			}

			if cmd.Flags().Changed("imap-host") {
				body["imap_host"] = imapHost
			}
			if cmd.Flags().Changed("imap-port") {
				body["imap_port"] = imapPort
			}
			if cmd.Flags().Changed("imap-username") {
				body["imap_username"] = imapUsername
			}
			if cmd.Flags().Changed("imap-password") {
				body["imap_password"] = imapPassword
			}

			if cmd.Flags().Changed("smtp-host") {
				body["smtp_host"] = smtpHost
			}
			if cmd.Flags().Changed("smtp-port") {
				body["smtp_port"] = smtpPort
			}
			if cmd.Flags().Changed("smtp-username") {
				body["smtp_username"] = smtpUsername
			}
			if cmd.Flags().Changed("smtp-password") {
				body["smtp_password"] = smtpPassword
			}

			emailVal, ok := body["email"].(string)
			if !ok || strings.TrimSpace(emailVal) == "" {
				return printError(cmd, "accounts.create", fmt.Errorf("--email is required (or set in --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/accounts", nil, body)
			if err != nil {
				return printError(cmd, "accounts.create", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "accounts.create", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Account email")
	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().IntVar(&providerCode, "provider-code", 0, "Provider code (1 IMAP, 2 Google, 3 Microsoft, 4 AWS)")

	cmd.Flags().StringVar(&imapHost, "imap-host", "", "IMAP host")
	cmd.Flags().IntVar(&imapPort, "imap-port", 0, "IMAP port")
	cmd.Flags().StringVar(&imapUsername, "imap-username", "", "IMAP username")
	cmd.Flags().StringVar(&imapPassword, "imap-password", "", "IMAP password")

	cmd.Flags().StringVar(&smtpHost, "smtp-host", "", "SMTP host")
	cmd.Flags().IntVar(&smtpPort, "smtp-port", 0, "SMTP port")
	cmd.Flags().StringVar(&smtpUsername, "smtp-username", "", "SMTP username")
	cmd.Flags().StringVar(&smtpPassword, "smtp-password", "", "SMTP password")

	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON body file, or '-' for stdin")

	return cmd
}

func newAccountsUpdateCmd() *cobra.Command {
	var (
		firstName          string
		lastName           string
		dailyLimit         int
		sendingGap         int
		trackingDomainName string

		dataJSON string
		dataFile string
	)

	cmd := &cobra.Command{
		Use:   "update <email>",
		Short: "Update account settings (PATCH /accounts/{email})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts.update", err, nil)
			}

			email := strings.TrimSpace(args[0])
			if email == "" {
				return printError(cmd, "accounts.update", fmt.Errorf("email is required"), nil)
			}

			body := map[string]any{}
			if strings.TrimSpace(dataJSON) != "" || strings.TrimSpace(dataFile) != "" {
				m, err := readJSONObjectInput(dataJSON, dataFile)
				if err != nil {
					return printError(cmd, "accounts.update", err, nil)
				}
				mergeMaps(body, m)
			}

			if cmd.Flags().Changed("first-name") {
				body["first_name"] = firstName
			}
			if cmd.Flags().Changed("last-name") {
				body["last_name"] = lastName
			}
			if cmd.Flags().Changed("daily-limit") {
				body["daily_limit"] = dailyLimit
			}
			if cmd.Flags().Changed("sending-gap") {
				body["sending_gap"] = sendingGap
			}
			if cmd.Flags().Changed("tracking-domain-name") {
				body["tracking_domain_name"] = trackingDomainName
			}

			if len(body) == 0 {
				return printError(cmd, "accounts.update", fmt.Errorf("no fields to update (provide flags or --data-json/--data-file)"), nil)
			}

			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/accounts/"+url.PathEscape(email), nil, body)
			if err != nil {
				return printError(cmd, "accounts.update", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "accounts.update", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&firstName, "first-name", "", "First name")
	cmd.Flags().StringVar(&lastName, "last-name", "", "Last name")
	cmd.Flags().IntVar(&dailyLimit, "daily-limit", 0, "Daily limit")
	cmd.Flags().IntVar(&sendingGap, "sending-gap", 0, "Minutes between emails")
	cmd.Flags().StringVar(&trackingDomainName, "tracking-domain-name", "", "Tracking domain name")

	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON body file, or '-' for stdin")

	return cmd
}

func newAccountsWarmupEnableCmd() *cobra.Command {
	return accountBulkEmailActionCmd("warmup-enable", "Enable warmup for an account", "/accounts/warmup/enable")
}

func newAccountsWarmupDisableCmd() *cobra.Command {
	return accountBulkEmailActionCmd("warmup-disable", "Disable warmup for an account", "/accounts/warmup/disable")
}

func newAccountsTestVitalsCmd() *cobra.Command {
	return accountBulkEmailActionCmd("test-vitals", "Test IMAP/SMTP vitals for an account", "/accounts/test/vitals")
}

func accountBulkEmailActionCmd(use, short, endpoint string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use + " <email>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts."+use, err, nil)
			}
			email := strings.TrimSpace(args[0])
			if email == "" {
				return printError(cmd, "accounts."+use, fmt.Errorf("email is required"), nil)
			}

			payload := map[string]any{"emails": []string{email}}
			resp, meta, err := client.PostJSON(cmdContext(cmd), endpoint, nil, payload)
			if err != nil {
				return printError(cmd, "accounts."+use, err, metaFrom(meta, nil))
			}
			return printResult(cmd, "accounts."+use, resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newAccountsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <email>",
		Short: "Delete an account (requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "accounts.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts.delete", err, nil)
			}
			email := strings.TrimSpace(args[0])
			if email == "" {
				return printError(cmd, "accounts.delete", fmt.Errorf("email is required"), nil)
			}

			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/accounts/"+url.PathEscape(email), nil)
			if err != nil {
				return printError(cmd, "accounts.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "accounts.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newAccountsAnalyticsDailyCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "analytics-daily",
		Short: "Daily account analytics (GET /accounts/analytics/daily)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "accounts.analytics_daily", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "accounts.analytics_daily", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/accounts/analytics/daily", q)
			if err != nil {
				return printError(cmd, "accounts.analytics_daily", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "accounts.analytics_daily", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}
