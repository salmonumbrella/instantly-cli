package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newAccountCampaignMappingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "account-campaign-mappings",
		Aliases: []string{"acm", "account-campaign-mapping"},
		Short:   "Resolve account campaign mappings by email",
	}
	cmd.AddCommand(newAccountCampaignMappingsGetCmd())
	return cmd
}

func newAccountCampaignMappingsGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <email>",
		Short: "Get account campaign mappings (GET /account-campaign-mappings/{email})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "account_campaign_mappings.get", err, nil)
			}
			email := strings.TrimSpace(args[0])
			if email == "" {
				return printError(cmd, "account_campaign_mappings.get", fmt.Errorf("email is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/account-campaign-mappings/"+url.PathEscape(email), nil)
			if err != nil {
				return printError(cmd, "account_campaign_mappings.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "account_campaign_mappings.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}
