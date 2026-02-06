package cmd

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newEmailsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "emails",
		Aliases: []string{"email", "e"},
		Short:   "Manage emails and inbox",
	}

	cmd.AddCommand(newEmailsListCmd())
	cmd.AddCommand(newEmailsGetCmd())
	cmd.AddCommand(newEmailsUnreadCountCmd())
	cmd.AddCommand(newEmailsReplyCmd())
	cmd.AddCommand(newEmailsForwardCmd())
	cmd.AddCommand(newEmailsUpdateCmd())
	cmd.AddCommand(newEmailsDeleteCmd())
	cmd.AddCommand(newEmailsVerifyCmd())
	cmd.AddCommand(newEmailsMarkThreadReadCmd())

	return cmd
}

func newEmailsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		search        string
		campaignID    string
		eaccount      string
		isUnread      bool
		isUnreadSet   bool
		queryPairs    []string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List emails",
		PreRun: func(cmd *cobra.Command, _ []string) {
			if cmd.Flags().Changed("unread") {
				isUnreadSet = true
			}
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.list", err, nil)
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
			if campaignID != "" {
				q.Set("campaign_id", campaignID)
			}
			if eaccount != "" {
				q.Set("eaccount", eaccount)
			}
			if isUnreadSet {
				q.Set("is_unread", fmt.Sprintf("%v", isUnread))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "emails.list", err, nil)
			}

			resp, meta, err := client.GetJSON(cmdContext(cmd), "/emails", q)
			if err != nil {
				return printError(cmd, "emails.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "emails.list", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination (use response next_starting_after)")
	cmd.Flags().StringVar(&search, "search", "", "Search emails")
	cmd.Flags().StringVar(&campaignID, "campaign-id", "", "Filter by campaign ID")
	cmd.Flags().StringVar(&eaccount, "eaccount", "", "Filter by sender account")
	cmd.Flags().BoolVar(&isUnread, "unread", false, "Filter unread emails (set explicitly to enable)")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")

	return cmd
}

func newEmailsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get <email_id>",
		Aliases: []string{"show"},
		Short:   "Get an email by ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "emails.get", fmt.Errorf("email_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/emails/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "emails.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "emails.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newEmailsUnreadCountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unread-count",
		Short: "Count unread emails",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.unread_count", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/emails/unread/count", nil)
			if err != nil {
				return printError(cmd, "emails.unread_count", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "emails.unread_count", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newEmailsReplyCmd() *cobra.Command {
	var (
		replyTo  string
		eaccount string
		subject  string
		body     string
		html     string
		text     string
		confirm  bool
	)

	cmd := &cobra.Command{
		Use:   "reply",
		Short: "Reply to an email thread (requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "emails.reply", fmt.Errorf("refusing to send email without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.reply", err, nil)
			}
			if strings.TrimSpace(replyTo) == "" {
				return printError(cmd, "emails.reply", fmt.Errorf("--reply-to is required"), nil)
			}
			if strings.TrimSpace(eaccount) == "" {
				return printError(cmd, "emails.reply", fmt.Errorf("--eaccount is required"), nil)
			}
			if strings.TrimSpace(subject) == "" {
				return printError(cmd, "emails.reply", fmt.Errorf("--subject is required"), nil)
			}

			// Payload shape (matches the MCP server): body is an object with optional html/text.
			// For backwards compatibility, --body maps to body.text unless --text/--html are set.
			bodyObj := map[string]any{}
			if strings.TrimSpace(html) != "" {
				bodyObj["html"] = html
			}
			if strings.TrimSpace(text) != "" {
				bodyObj["text"] = text
			}
			if strings.TrimSpace(body) != "" && len(bodyObj) == 0 {
				bodyObj["text"] = body
			}
			if len(bodyObj) == 0 {
				return printError(cmd, "emails.reply", fmt.Errorf("provide --text or --html (or legacy --body)"), nil)
			}

			payload := map[string]any{
				"reply_to_uuid": replyTo,
				"eaccount":      eaccount,
				"subject":       subject,
				"body":          bodyObj,
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/emails/reply", nil, payload)
			if err != nil {
				return printError(cmd, "emails.reply", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "emails.reply", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&replyTo, "reply-to", "", "Email UUID to reply to")
	cmd.Flags().StringVar(&eaccount, "eaccount", "", "Sender account email")
	cmd.Flags().StringVar(&subject, "subject", "", "Email subject")
	cmd.Flags().StringVar(&text, "text", "", "Email body text")
	cmd.Flags().StringVar(&html, "html", "", "Email body HTML")
	cmd.Flags().StringVar(&body, "body", "", "(legacy) Email body text")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm sending a real email")

	return cmd
}

func newEmailsVerifyCmd() *cobra.Command {
	var (
		email        string
		maxWait      time.Duration
		pollInterval time.Duration
		skipPolling  bool
	)

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify an email (polls until final status by default)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.verify", err, nil)
			}
			if strings.TrimSpace(email) == "" {
				return printError(cmd, "emails.verify", fmt.Errorf("--email is required"), nil)
			}
			if maxWait <= 0 {
				maxWait = 45 * time.Second
			}
			if pollInterval <= 0 {
				pollInterval = 2 * time.Second
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/email-verification", nil, map[string]any{"email": email})
			if err != nil {
				return printError(cmd, "emails.verify", err, metaFrom(meta, nil))
			}

			// Poll if pending.
			status := ""
			if m, ok := resp.(map[string]any); ok {
				if s, ok := m["verification_status"].(string); ok {
					status = s
				}
			}

			if status == "pending" && !skipPolling && maxWait > 0 {
				start := time.Now()
				polls := 0
				for time.Since(start) < maxWait {
					time.Sleep(pollInterval)
					polls++
					pollResp, _, pollErr := client.GetJSON(cmdContext(cmd), "/email-verification/"+url.PathEscape(email), nil)
					if pollErr != nil {
						continue
					}
					if m, ok := pollResp.(map[string]any); ok {
						if s, ok := m["verification_status"].(string); ok && (s == "verified" || s == "invalid") {
							m["_polling_info"] = map[string]any{
								"polls_made":         polls,
								"total_time_seconds": time.Since(start).Seconds(),
								"final_status":       s,
							}
							resp = m
							break
						}
					}
				}

				// Timeout reached while still pending: add context for agents.
				if m, ok := resp.(map[string]any); ok {
					if s, _ := m["verification_status"].(string); s == "pending" {
						m["_polling_info"] = map[string]any{
							"polls_made":         polls,
							"total_time_seconds": time.Since(start).Seconds(),
							"timeout_reached":    true,
							"note":               "Verification still pending after max wait. Check later with GET /email-verification/{email}.",
						}
						resp = m
					}
				}
			}

			return printResult(cmd, "emails.verify", resp, metaFrom(meta, resp))
		},
	}

	cmd.Flags().StringVar(&email, "email", "", "Email address to verify")
	cmd.Flags().DurationVar(&maxWait, "max-wait", 45*time.Second, "Max time to poll for a final result")
	cmd.Flags().DurationVar(&pollInterval, "poll-interval", 2*time.Second, "Time between polling attempts")
	cmd.Flags().BoolVar(&skipPolling, "skip-polling", false, "Return immediately even if status is pending")

	return cmd
}

func newEmailsMarkThreadReadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mark-thread-read <thread_id>",
		Short: "Mark an email thread as read",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.mark_thread_read", err, nil)
			}
			threadID := strings.TrimSpace(args[0])
			if threadID == "" {
				return printError(cmd, "emails.mark_thread_read", fmt.Errorf("thread_id is required"), nil)
			}

			resp, meta, err := client.PostJSON(cmdContext(cmd), "/emails/threads/"+url.PathEscape(threadID)+"/mark-as-read", nil, nil)
			if err != nil {
				return printError(cmd, "emails.mark_thread_read", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "emails.mark_thread_read", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newEmailsForwardCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "forward",
		Short: "Forward an email (POST /emails/forward, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "emails.forward", fmt.Errorf("refusing to forward email without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.forward", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "emails.forward", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "emails.forward", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/emails/forward", nil, body)
			if err != nil {
				return printError(cmd, "emails.forward", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "emails.forward", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm sending a real email forward")
	return cmd
}

func newEmailsUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update <email_id>",
		Short: "Update email (PATCH /emails/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "emails.update", fmt.Errorf("email_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "emails.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "emails.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/emails/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "emails.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "emails.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newEmailsDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <email_id>",
		Short: "Delete email (DELETE /emails/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "emails.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "emails.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "emails.delete", fmt.Errorf("email_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/emails/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "emails.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "emails.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}
