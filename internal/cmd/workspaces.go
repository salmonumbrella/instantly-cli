package cmd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

func newWorkspacesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "workspaces",
		Aliases:   []string{"workspace"},
		Short:     "Workspace management",
		ValidArgs: []string{"current"},
	}

	cmd.AddCommand(newWorkspacesCurrentCmd())
	cmd.AddCommand(newWorkspacesCreateCmd())
	cmd.AddCommand(newWorkspacesChangeOwnerCmd())
	cmd.AddCommand(newWorkspacesWhiteLabelDomainCmd())

	return cmd
}

func newWorkspacesCurrentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "Get or update current workspace",
	}
	cmd.AddCommand(newWorkspacesCurrentGetCmd())
	cmd.AddCommand(newWorkspacesCurrentUpdateCmd())
	return cmd
}

func newWorkspacesCurrentGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get current workspace (GET /workspaces/current)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.current.get", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspaces/current", nil)
			if err != nil {
				return printError(cmd, "workspaces.current.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspaces.current.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWorkspacesCurrentUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update current workspace (PATCH /workspaces/current)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.current.update", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspaces.current.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "workspaces.current.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/workspaces/current", nil, body)
			if err != nil {
				return printError(cmd, "workspaces.current.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspaces.current.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newWorkspacesCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a workspace (POST /workspaces/create, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "workspaces.create", fmt.Errorf("refusing to create workspace without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspaces.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "workspaces.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/workspaces/create", nil, body)
			if err != nil {
				return printError(cmd, "workspaces.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspaces.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm creating a workspace")
	return cmd
}

func newWorkspacesChangeOwnerCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
		confirm  bool
	)
	cmd := &cobra.Command{
		Use:   "change-owner",
		Short: "Change workspace owner (POST /workspaces/current/change-owner, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "workspaces.change_owner", fmt.Errorf("refusing to change owner without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.change_owner", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspaces.change_owner", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "workspaces.change_owner", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/workspaces/current/change-owner", nil, body)
			if err != nil {
				return printError(cmd, "workspaces.change_owner", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspaces.change_owner", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm changing owner")
	return cmd
}

func newWorkspacesWhiteLabelDomainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "whitelabel-domain",
		Aliases: []string{"white-label-domain", "wl-domain"},
		Short:   "Manage workspace whitelabel domain",
	}
	cmd.AddCommand(newWorkspacesWhiteLabelDomainGetCmd())
	cmd.AddCommand(newWorkspacesWhiteLabelDomainSetCmd())
	cmd.AddCommand(newWorkspacesWhiteLabelDomainDeleteCmd())
	return cmd
}

func newWorkspacesWhiteLabelDomainGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get whitelabel domain (GET /workspaces/current/whitelabel-domain)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.get", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspaces/current/whitelabel-domain", nil)
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspaces.whitelabel_domain.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWorkspacesWhiteLabelDomainSetCmd() *cobra.Command {
	var (
		domain   string
		confirm  bool
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set whitelabel domain (POST /workspaces/current/whitelabel-domain, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "workspaces.whitelabel_domain.set", fmt.Errorf("refusing to set whitelabel domain without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.set", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.set", err, nil)
			}
			if strings.TrimSpace(domain) != "" {
				body["domain"] = strings.TrimSpace(domain)
			}
			domainVal, ok := body["domain"].(string)
			if !ok || strings.TrimSpace(domainVal) == "" {
				return printError(cmd, "workspaces.whitelabel_domain.set", fmt.Errorf("--domain is required (or set in body)"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/workspaces/current/whitelabel-domain", nil, body)
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.set", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspaces.whitelabel_domain.set", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&domain, "domain", "", "Domain to set")
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body (merged with flags)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin (merged with flags)")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm updating whitelabel domain")
	return cmd
}

func newWorkspacesWhiteLabelDomainDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete whitelabel domain (DELETE /workspaces/current/whitelabel-domain, requires --confirm)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !confirm {
				return printError(cmd, "workspaces.whitelabel_domain.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.delete", err, nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/workspaces/current/whitelabel-domain", nil)
			if err != nil {
				return printError(cmd, "workspaces.whitelabel_domain.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspaces.whitelabel_domain.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newWorkspaceMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace-members",
		Aliases: []string{"members", "workspace-member"},
		Short:   "Manage workspace members",
	}

	cmd.AddCommand(newWorkspaceMembersListCmd())
	cmd.AddCommand(newWorkspaceMembersGetCmd())
	cmd.AddCommand(newWorkspaceMembersCreateCmd())
	cmd.AddCommand(newWorkspaceMembersUpdateCmd())
	cmd.AddCommand(newWorkspaceMembersDeleteCmd())

	return cmd
}

func newWorkspaceMembersListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspace members (GET /workspace-members)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_members.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "workspace_members.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-members", q)
			if err != nil {
				return printError(cmd, "workspace_members.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_members.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newWorkspaceMembersGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <member_id>",
		Short: "Get workspace member (GET /workspace-members/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_members.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "workspace_members.get", fmt.Errorf("member_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-members/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "workspace_members.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_members.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWorkspaceMembersCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create workspace member (POST /workspace-members)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_members.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspace_members.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "workspace_members.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/workspace-members", nil, body)
			if err != nil {
				return printError(cmd, "workspace_members.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspace_members.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	return cmd
}

func newWorkspaceMembersUpdateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "update <member_id>",
		Short: "Update workspace member (PATCH /workspace-members/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_members.update", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "workspace_members.update", fmt.Errorf("member_id is required"), nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspace_members.update", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "workspace_members.update", fmt.Errorf("provide --data-json or --data-file with at least one field"), nil)
			}
			resp, meta, err := client.PatchJSON(cmdContext(cmd), "/workspace-members/"+url.PathEscape(id), nil, body)
			if err != nil {
				return printError(cmd, "workspace_members.update", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspace_members.update", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON patch body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON patch body file, or '-' for stdin")
	return cmd
}

func newWorkspaceMembersDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <member_id>",
		Short: "Delete workspace member (DELETE /workspace-members/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "workspace_members.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_members.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "workspace_members.delete", fmt.Errorf("member_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/workspace-members/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "workspace_members.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_members.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newWorkspaceGroupMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace-group-members",
		Aliases: []string{"group-members", "workspace-group-member"},
		Short:   "Manage workspace group members",
	}

	cmd.AddCommand(newWorkspaceGroupMembersListCmd())
	cmd.AddCommand(newWorkspaceGroupMembersGetCmd())
	cmd.AddCommand(newWorkspaceGroupMembersCreateCmd())
	cmd.AddCommand(newWorkspaceGroupMembersDeleteCmd())
	cmd.AddCommand(newWorkspaceGroupMembersAdminCmd())

	return cmd
}

func newWorkspaceGroupMembersListCmd() *cobra.Command {
	var (
		limit         int
		startingAfter string
		queryPairs    []string
	)
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workspace group members (GET /workspace-group-members)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_group_members.list", err, nil)
			}
			q := url.Values{}
			if limit > 0 {
				q.Set("limit", fmt.Sprint(limit))
			}
			if strings.TrimSpace(startingAfter) != "" {
				q.Set("starting_after", strings.TrimSpace(startingAfter))
			}
			if err := applyQueryPairs(q, queryPairs); err != nil {
				return printError(cmd, "workspace_group_members.list", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-group-members", q)
			if err != nil {
				return printError(cmd, "workspace_group_members.list", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_group_members.list", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Max results (default 100)")
	cmd.Flags().StringVar(&startingAfter, "starting-after", "", "Cursor for pagination")
	cmd.Flags().StringArrayVar(&queryPairs, "query", nil, "Extra query param (repeatable): key=value")
	return cmd
}

func newWorkspaceGroupMembersGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <group_member_id>",
		Short: "Get workspace group member (GET /workspace-group-members/{id})",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_group_members.get", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "workspace_group_members.get", fmt.Errorf("group_member_id is required"), nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-group-members/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "workspace_group_members.get", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_group_members.get", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWorkspaceGroupMembersCreateCmd() *cobra.Command {
	var (
		dataJSON string
		dataFile string
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create workspace group member (POST /workspace-group-members)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_group_members.create", err, nil)
			}
			body, err := readJSONObjectInput(dataJSON, dataFile)
			if err != nil {
				return printError(cmd, "workspace_group_members.create", err, nil)
			}
			if len(body) == 0 {
				return printError(cmd, "workspace_group_members.create", fmt.Errorf("provide --data-json or --data-file"), nil)
			}
			resp, meta, err := client.PostJSON(cmdContext(cmd), "/workspace-group-members", nil, body)
			if err != nil {
				return printError(cmd, "workspace_group_members.create", err, metaFrom(meta, nil))
			}
			return printWriteResult(cmd, "workspace_group_members.create", resp, metaFrom(meta, resp), body)
		},
	}
	cmd.Flags().StringVar(&dataJSON, "data-json", "", "Raw JSON request body")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to JSON request body file, or '-' for stdin")
	return cmd
}

func newWorkspaceGroupMembersDeleteCmd() *cobra.Command {
	var confirm bool
	cmd := &cobra.Command{
		Use:   "delete <group_member_id>",
		Short: "Delete workspace group member (DELETE /workspace-group-members/{id}, requires --confirm)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return printError(cmd, "workspace_group_members.delete", fmt.Errorf("refusing to delete without --confirm"), nil)
			}
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_group_members.delete", err, nil)
			}
			id := strings.TrimSpace(args[0])
			if id == "" {
				return printError(cmd, "workspace_group_members.delete", fmt.Errorf("group_member_id is required"), nil)
			}
			resp, meta, err := client.DeleteJSON(cmdContext(cmd), "/workspace-group-members/"+url.PathEscape(id), nil)
			if err != nil {
				return printError(cmd, "workspace_group_members.delete", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_group_members.delete", resp, metaFrom(meta, resp))
		},
	}
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive action")
	return cmd
}

func newWorkspaceGroupMembersAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Get workspace group members admin info (GET /workspace-group-members/admin)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_group_members.admin", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-group-members/admin", nil)
			if err != nil {
				return printError(cmd, "workspace_group_members.admin", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_group_members.admin", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWorkspaceBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace-billing",
		Aliases: []string{"billing"},
		Short:   "Workspace billing details",
	}
	cmd.AddCommand(newWorkspaceBillingPlanDetailsCmd())
	cmd.AddCommand(newWorkspaceBillingSubscriptionDetailsCmd())
	return cmd
}

func newWorkspaceBillingPlanDetailsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan-details",
		Short: "Get plan details (GET /workspace-billing/plan-details)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_billing.plan_details", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-billing/plan-details", nil)
			if err != nil {
				return printError(cmd, "workspace_billing.plan_details", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_billing.plan_details", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}

func newWorkspaceBillingSubscriptionDetailsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subscription-details",
		Short: "Get subscription details (GET /workspace-billing/subscription-details)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := clientFromFlags()
			if err != nil {
				return printError(cmd, "workspace_billing.subscription_details", err, nil)
			}
			resp, meta, err := client.GetJSON(cmdContext(cmd), "/workspace-billing/subscription-details", nil)
			if err != nil {
				return printError(cmd, "workspace_billing.subscription_details", err, metaFrom(meta, nil))
			}
			return printResult(cmd, "workspace_billing.subscription_details", resp, metaFrom(meta, resp))
		},
	}
	return cmd
}
