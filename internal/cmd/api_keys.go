package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newAPIKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "api-keys",
		Aliases: []string{"apikeys", "api-key"},
		Short:   "Manage Instantly API keys",
	}

	cmd.AddCommand(newAPIKeysListCmd())
	cmd.AddCommand(newAPIKeysCreateCmd())
	cmd.AddCommand(newAPIKeysDeleteCmd())

	return cmd
}

func newAPIKeysListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List API keys (GET /api-keys)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "api_keys.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "api_keys.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/api-keys", q)
			if err != nil {
				return printError(cmd, "api_keys.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "api_keys.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newAPIKeysCreateCmd() *cobra.Command {
	var (
		name     string
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create API key (POST /api-keys, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "api_keys.create", fmt.Errorf("refusing to create api key without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "api_keys.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "api_keys.create", err, nil)
			}
			if strings.TrimSpace(name) != "" {
				body["name"] = strings.TrimSpace(name)
			}
			nameVal, ok := body["name"].(string)
			if !ok || strings.TrimSpace(nameVal) == "" {
				return printError(cmd, "api_keys.create", fmt.Errorf("--name is required (or set in --data-json/--data-file)"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/api-keys", nil, body)
			if err != nil {
				return printError(cmd, "api_keys.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "api_keys.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "API key name")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged with flags)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm creating a new API key")
	return cmd
}

func newAPIKeysDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <api_key_id>",
		Short: "Delete API key (DELETE /api-keys/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "api_keys.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "api_keys.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "api_keys.delete", fmt.Errorf("api_key_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/api-keys/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "api_keys.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "api_keys.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}
