package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/salmonumbrella/instantly-cli/internal/api"
	"github.com/salmonumbrella/instantly-cli/internal/outfmt"
)

type rootFlags struct {
	Output  string
	JSON    bool
	Quiet   bool
	Silent  bool
	Debug   bool
	Timeout time.Duration
	BaseURL string
	APIKey  string
	DryRun  bool

	JQ     string
	Fields string

	Max429Retries  int
	Max5xxRetries  int
	RetryDelay     time.Duration
	MaxRetryDelay  time.Duration
	IdempotencyKey string
}

var flags = rootFlags{
	Output:  defaultOutput(),
	Timeout: 60 * time.Second,
	BaseURL: api.DefaultBaseURL,
}

func defaultOutput() string {
	v := strings.TrimSpace(os.Getenv("INSTANTLY_OUTPUT"))
	if v != "" {
		return v
	}
	// This CLI is agent-only: default to "agent" output to provide stable envelopes + meta.
	return "agent"
}

func resetFlagsToDefaults() {
	flags.Output = defaultOutput()
	flags.JSON = false
	flags.Quiet = false
	flags.Silent = false
	flags.Debug = false
	flags.Timeout = 60 * time.Second
	flags.BaseURL = api.DefaultBaseURL
	flags.APIKey = strings.TrimSpace(os.Getenv("INSTANTLY_API_KEY"))
	flags.DryRun = false

	flags.JQ = ""
	flags.Fields = ""

	flags.Max429Retries = 0
	flags.Max5xxRetries = 0
	flags.RetryDelay = 1 * time.Second
	flags.MaxRetryDelay = 30 * time.Second
	flags.IdempotencyKey = ""
}

func newRootCmd() *cobra.Command {
	resetFlagsToDefaults()

	cmd := &cobra.Command{
		Use:           "instantly",
		Short:         "Agent-friendly CLI for Instantly.ai",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// --json is just a shorthand for --output json.
			if flags.JSON {
				if cmd.Flags().Changed("output") && flags.Output != "json" {
					return fmt.Errorf("--json conflicts with --output %s", flags.Output)
				}
				flags.Output = "json"
			}

			mode, err := outfmt.ParseMode(flags.Output)
			if err != nil {
				return err
			}

			ctx := cmd.Context()
			ctx = outfmt.WithMode(ctx, mode)
			cmd.SetContext(ctx)

			if strings.TrimSpace(flags.JQ) != "" || strings.TrimSpace(flags.Fields) != "" {
				// Filtering only applies to JSON-ish outputs.
				if mode != outfmt.JSON && mode != outfmt.JSONL && mode != outfmt.Agent {
					if cmd.Flags().Changed("output") {
						return fmt.Errorf("--jq/--fields require --output json, jsonl, or agent (or omit --output)")
					}
					flags.Output = "json"
					mode = outfmt.JSON
					ctx = outfmt.WithMode(cmd.Context(), mode)
					cmd.SetContext(ctx)
				}
			}

			if flags.Silent {
				cmd.SetOut(io.Discard)
				cmd.SetErr(io.Discard)
				return nil
			}

			if flags.Quiet {
				// Quiet: suppress stderr and text output, but allow JSON outputs.
				cmd.SetErr(io.Discard)
				if mode == outfmt.Text {
					cmd.SetOut(io.Discard)
				}
			}

			return nil
		},
	}

	initRootFlagsAndCommands(cmd)
	return cmd
}

// Execute runs the root command.
func Execute() error {
	cmd := newRootCmd()
	cmd.SetContext(context.Background())
	return cmd.Execute()
}

func cmdContext(cmd *cobra.Command) context.Context { return cmd.Context() }

func clientFromFlags() (*api.Client, error) {
	if strings.TrimSpace(flags.APIKey) == "" && !flags.DryRun {
		return nil, errors.New("missing API key: set INSTANTLY_API_KEY or pass --api-key")
	}
	c := api.NewClient(flags.BaseURL, flags.APIKey, flags.Timeout)
	c.DryRun = flags.DryRun
	c.Max429Retries = flags.Max429Retries
	c.Max5xxRetries = flags.Max5xxRetries
	c.RetryDelay = flags.RetryDelay
	c.MaxRetryDelay = flags.MaxRetryDelay
	c.IdempotencyKey = flags.IdempotencyKey
	return c, nil
}

func initRootFlagsAndCommands(rootCmd *cobra.Command) {
	rootCmd.PersistentFlags().StringVarP(&flags.Output, "output", "o", flags.Output, "output format (text, json, jsonl, agent)")
	rootCmd.PersistentFlags().BoolVar(&flags.JSON, "json", false, "shorthand for --output json")
	rootCmd.PersistentFlags().BoolVar(&flags.Quiet, "quiet", false, "suppress stderr and text output")
	rootCmd.PersistentFlags().BoolVar(&flags.Silent, "silent", false, "suppress all output (stdout and stderr)")
	rootCmd.PersistentFlags().BoolVar(&flags.Debug, "debug", false, "debug logging (stderr)")
	rootCmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", flags.Timeout, "http timeout (e.g. 30s, 2m)")
	rootCmd.PersistentFlags().StringVar(&flags.BaseURL, "base-url", flags.BaseURL, "Instantly API base URL")
	rootCmd.PersistentFlags().StringVar(&flags.APIKey, "api-key", flags.APIKey, "Instantly API key (or set INSTANTLY_API_KEY)")
	rootCmd.PersistentFlags().BoolVar(&flags.DryRun, "dry-run", false, "Do not make network calls; print the request that would be made")

	rootCmd.PersistentFlags().StringVar(&flags.JQ, "jq", "", "JQ expression to filter JSON/agent output")
	rootCmd.PersistentFlags().StringVar(&flags.Fields, "fields", "", "Comma-separated fields to select (shorthand for --jq)")

	rootCmd.PersistentFlags().IntVar(&flags.Max429Retries, "max-429-retries", 0, "Max retries for 429 responses (default 0; safe for GETs)")
	rootCmd.PersistentFlags().IntVar(&flags.Max5xxRetries, "max-5xx-retries", 0, "Max retries for transient 5xx responses (default 0; safe for GETs)")
	rootCmd.PersistentFlags().DurationVar(&flags.RetryDelay, "retry-delay", flags.RetryDelay, "Base delay between retries (e.g. 1s)")
	rootCmd.PersistentFlags().DurationVar(&flags.MaxRetryDelay, "max-retry-delay", flags.MaxRetryDelay, "Max delay between retries")
	rootCmd.PersistentFlags().StringVar(&flags.IdempotencyKey, "idempotency-key", "", "Idempotency key for write requests (enables safe retries for writes when supported)")

	rootCmd.AddCommand(newAccountsCmd())
	rootCmd.AddCommand(newAccountCampaignMappingsCmd())
	rootCmd.AddCommand(newCampaignsCmd())
	rootCmd.AddCommand(newLeadsCmd())
	rootCmd.AddCommand(newLeadListsCmd())
	rootCmd.AddCommand(newEmailsCmd())
	rootCmd.AddCommand(newAnalyticsCmd())
	rootCmd.AddCommand(newJobsCmd())
	rootCmd.AddCommand(newAPICmd())
	rootCmd.AddCommand(newWebhooksCmd())
	rootCmd.AddCommand(newTagsCmd())
	rootCmd.AddCommand(newBlockListEntriesCmd())
	rootCmd.AddCommand(newLeadLabelsCmd())
	rootCmd.AddCommand(newSubsequencesCmd())
	rootCmd.AddCommand(newInboxPlacementCmd())
	rootCmd.AddCommand(newOAuthCmd())
	rootCmd.AddCommand(newAPIKeysCmd())
	rootCmd.AddCommand(newAuditLogsCmd())
	rootCmd.AddCommand(newWorkspacesCmd())
	rootCmd.AddCommand(newWorkspaceMembersCmd())
	rootCmd.AddCommand(newWorkspaceGroupMembersCmd())
	rootCmd.AddCommand(newWorkspaceBillingCmd())
	rootCmd.AddCommand(newSupersearchEnrichmentCmd())
	rootCmd.AddCommand(newCRMActionsCmd())
	rootCmd.AddCommand(newDFYEmailAccountOrdersCmd())
	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newSchemaCmd())
}
