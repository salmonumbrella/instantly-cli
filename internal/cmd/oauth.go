package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newOAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oauth",
		Short: "OAuth flows (Google/Microsoft account connection)",
	}

	cmd.AddCommand(newOAuthGoogleInitCmd())
	cmd.AddCommand(newOAuthMicrosoftInitCmd())
	cmd.AddCommand(newOAuthSessionStatusCmd())

	return cmd
}

func newOAuthGoogleInitCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "google-init",
		Short: "Init Google OAuth (POST /oauth/google/init)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "oauth.google_init", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "oauth.google_init", err, nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/oauth/google/init", nil, body)
			if err != nil {
				return printError(cmd, "oauth.google_init", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "oauth.google_init", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	return cmd
}

func newOAuthMicrosoftInitCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "microsoft-init",
		Short: "Init Microsoft OAuth (POST /oauth/microsoft/init)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "oauth.microsoft_init", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "oauth.microsoft_init", err, nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/oauth/microsoft/init", nil, body)
			if err != nil {
				return printError(cmd, "oauth.microsoft_init", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "oauth.microsoft_init", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	return cmd
}

func newOAuthSessionStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session-status <session_id>",
		Short: "Get OAuth session status (GET /oauth/session/status/{sessionId})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "oauth.session_status", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "oauth.session_status", fmt.Errorf("session_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/oauth/session/status/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "oauth.session_status", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "oauth.session_status", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}
