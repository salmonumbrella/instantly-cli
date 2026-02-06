package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newCRMActionsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "crm-actions",
		Aliases: []string{"crm"},
		Short:   "CRM-related operations",
	}

	cmd.AddCommand(newCRMPhoneNumbersCmd())
	return cmd
}

func newCRMPhoneNumbersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "phone-numbers",
		Aliases: []string{"phones"},
		Short:   "Manage CRM phone numbers",
	}
	cmd.AddCommand(newCRMPhoneNumbersListCmd())
	cmd.AddCommand(newCRMPhoneNumbersDeleteCmd())
	return cmd
}

func newCRMPhoneNumbersListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List phone numbers (GET /crm-actions/phone-numbers)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "crm_actions.phone_numbers.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "crm_actions.phone_numbers.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/crm-actions/phone-numbers", q)
			if err != nil {
				return printError(cmd, "crm_actions.phone_numbers.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "crm_actions.phone_numbers.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newCRMPhoneNumbersDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <phone_number_id>",
		Short: "Delete phone number (DELETE /crm-actions/phone-numbers/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "crm_actions.phone_numbers.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "crm_actions.phone_numbers.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "crm_actions.phone_numbers.delete", fmt.Errorf("phone_number_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/crm-actions/phone-numbers/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "crm_actions.phone_numbers.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "crm_actions.phone_numbers.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newDFYEmailAccountOrdersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dfy-email-account-orders",
		Aliases: []string{"dfy-orders", "orders"},
		Short:   "DFY (done-for-you) email account orders",
	}

	cmd.AddCommand(newDFYOrdersListCmd())
	cmd.AddCommand(newDFYOrdersCreateCmd())
	cmd.AddCommand(newDFYOrdersAccountsCmd())
	cmd.AddCommand(newDFYOrdersDomainsCmd())

	return cmd
}

func newDFYOrdersListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DFY orders (GET /dfy-email-account-orders)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "dfy_orders.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "dfy_orders.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/dfy-email-account-orders", q)
			if err != nil {
				return printError(cmd, "dfy_orders.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "dfy_orders.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newDFYOrdersCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a DFY order (POST /dfy-email-account-orders, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "dfy_orders.create", fmt.Errorf("refusing to create order without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "dfy_orders.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "dfy_orders.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "dfy_orders.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/dfy-email-account-orders", nil, body)
			if err != nil {
				return printError(cmd, "dfy_orders.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "dfy_orders.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm creating an order")
	return cmd
}

func newDFYOrdersAccountsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "accounts",
		Aliases: []string{"email-accounts"},
		Short:   "DFY email accounts",
	}
	cmd.AddCommand(newDFYOrdersAccountsListCmd())
	cmd.AddCommand(newDFYOrdersAccountsCancelCmd())
	return cmd
}

func newDFYOrdersAccountsListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DFY email accounts (GET /dfy-email-account-orders/accounts)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "dfy_orders.accounts.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "dfy_orders.accounts.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/dfy-email-account-orders/accounts", q)
			if err != nil {
				return printError(cmd, "dfy_orders.accounts.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "dfy_orders.accounts.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newDFYOrdersAccountsCancelCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel DFY email accounts (POST /dfy-email-account-orders/accounts/cancel, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "dfy_orders.accounts.cancel", fmt.Errorf("refusing to cancel without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "dfy_orders.accounts.cancel", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "dfy_orders.accounts.cancel", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "dfy_orders.accounts.cancel", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/dfy-email-account-orders/accounts/cancel", nil, body)
			if err != nil {
				return printError(cmd, "dfy_orders.accounts.cancel", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "dfy_orders.accounts.cancel", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm canceling accounts")
	return cmd
}

func newDFYOrdersDomainsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domains",
		Aliases: []string{"domain"},
		Short:   "DFY domain utilities",
	}
	cmd.AddCommand(newDFYOrdersDomainsCheckCmd())
	cmd.AddCommand(newDFYOrdersDomainsSimilarCmd())
	cmd.AddCommand(newDFYOrdersDomainsPreWarmedUpListCmd())
	return cmd
}

func newDFYOrdersDomainsCheckCmd() *cobra.Command {
	return dfyOrdersDomainsPostCmd(
		"check",
		"dfy_orders.domains.check",
		"Check domains (POST /dfy-email-account-orders/domains/check, requires --confirm)",
		"/dfy-email-account-orders/domains/check",
	)
}

func newDFYOrdersDomainsSimilarCmd() *cobra.Command {
	return dfyOrdersDomainsPostCmd(
		"similar",
		"dfy_orders.domains.similar",
		"Find similar domains (POST /dfy-email-account-orders/domains/similar, requires --confirm)",
		"/dfy-email-account-orders/domains/similar",
	)
}

func newDFYOrdersDomainsPreWarmedUpListCmd() *cobra.Command {
	return dfyOrdersDomainsPostCmd(
		"pre-warmed-up-list",
		"dfy_orders.domains.pre_warmed_up_list",
		"Get pre-warmed-up domain list (POST /dfy-email-account-orders/domains/pre-warmed-up-list, requires --confirm)",
		"/dfy-email-account-orders/domains/pre-warmed-up-list",
	)
}

func dfyOrdersDomainsPostCmd(use, op, short, endpoint string) *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, op, fmt.Errorf("refusing to run without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, op, err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, op, err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, op, fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), endpoint, nil, body)
			if err != nil {
				return printError(cmd, op, err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, op, resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm action")
	return cmd
}
