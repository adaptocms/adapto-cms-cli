package apikey

import (
	"fmt"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/adaptocms/adapto-cms-cli/internal/output"
	"github.com/adaptocms/adapto-cms-cli/internal/prompt"
	"github.com/spf13/cobra"
)

// Cmd is the root api-key command.
var Cmd = &cobra.Command{
	Use:   "api-key",
	Short: "Manage Public API keys",
	Long:  "Issue, list, and revoke Public API keys. A key authenticates an Adapto client app against the Public API; keep it confidential.",
}

var issueCmd = &cobra.Command{
	Use:     "issue",
	Short:   "Issue a Public API key for the active project",
	Example: "adapto api-key issue --expires-in 90d --json",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project-id")
		expiresIn, _ := cmd.Flags().GetString("expires-in")

		c, cfg, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}
		if projectID == "" {
			projectID = cfg.TenantID
		}
		if projectID == "" {
			return fmt.Errorf("no active project: run 'adapto project use <id>' or pass --project-id")
		}

		tResp, err := c.GetTenantTenantsTenantIdGetWithResponse(cmdutil.Ctx(), projectID)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(tResp.HTTPResponse, tResp.Body); err != nil {
			return err
		}
		if tResp.JSON200 == nil {
			return fmt.Errorf("project %s not found", projectID)
		}

		expiresAt, err := cmdutil.PromptExpiration(expiresIn)
		if err != nil {
			return err
		}

		resp, err := c.IssueApiKeyApiKeysPostWithResponse(cmdutil.Ctx(), client.IssueApiKeyRequest{
			OrganizationId: tResp.JSON200.OrganizationId,
			TenantId:       projectID,
			ExpiresAt:      expiresAt,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}
		if resp.JSON200 == nil {
			return fmt.Errorf("api key issuance returned no data")
		}

		key := *resp.JSON200
		output.Print(key, func(d interface{}) {
			fmt.Printf("API key issued: %s\n", key.Value)
			fmt.Println("Store this in your client app's environment (e.g. ADAPTO_API_KEY). Keep it confidential.")
		})
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List the active project's API keys",
	Example: "adapto api-key list --json",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project-id")

		c, cfg, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}
		if projectID == "" {
			projectID = cfg.TenantID
		}
		if projectID == "" {
			return fmt.Errorf("no active project: run 'adapto project use <id>' or pass --project-id")
		}

		resp, err := c.ListTenantApiKeysApiKeysByTenantTenantIdGetWithResponse(cmdutil.Ctx(), projectID)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}

		keys := []client.PublicApiKeyResponse{}
		if resp.JSON200 != nil {
			keys = *resp.JSON200
		}
		output.Print(keys, func(d interface{}) {
			if len(keys) == 0 {
				fmt.Println("No API keys found for this project.")
				return
			}
			for _, k := range keys {
				fmt.Printf("%s  %s  %s\n", k.Id, k.Status, k.Value)
			}
		})
		return nil
	},
}

var revokeCmd = &cobra.Command{
	Use:     "revoke [api-key-id]",
	Short:   "Revoke an API key",
	Example: "adapto api-key revoke <api-key-id>",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKeyID, _ := cmd.Flags().GetString("api-key-id")
		if len(args) > 0 {
			apiKeyID = args[0]
		}

		var err error
		if apiKeyID, err = prompt.RequireArg("api-key-id", apiKeyID); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteApiKeyApiKeysApiKeyIdDeleteWithResponse(cmdutil.Ctx(), apiKeyID)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}

		output.Successf("API key %s revoked.", apiKeyID)
		return nil
	},
}

func init() {
	Cmd.AddCommand(issueCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(revokeCmd)

	issueCmd.Flags().String("project-id", "", "Project ID to issue a key for (defaults to the active project)")
	issueCmd.Flags().String("expires-in", "", "Key expiration: never, 30d, 90d, or 1y (interactive selector if omitted)")
	listCmd.Flags().String("project-id", "", "Project ID to list keys for (defaults to the active project)")
	revokeCmd.Flags().String("api-key-id", "", "API key ID to revoke")
}
