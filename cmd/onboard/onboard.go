package onboard

import (
	"fmt"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	"github.com/adaptocms/adapto-cms-cli/internal/output"
	"github.com/adaptocms/adapto-cms-cli/internal/prompt"
	"github.com/spf13/cobra"
)

// Cmd is the onboard command.
var Cmd = &cobra.Command{
	Use:     "onboard",
	Short:   "Set up your first project and API key",
	Long:    "Create your first project, issue an API key, and set it as the active project. Run this once after activating your account.",
	Example: "adapto onboard --project-name \"My Project\" --default-language en-US --languages es-ES,fr-FR",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("project-name")
		description, _ := cmd.Flags().GetString("description")
		defaultLanguage, _ := cmd.Flags().GetString("default-language")
		secondary, _ := cmd.Flags().GetStringSlice("languages")

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		orgID, err := cmdutil.ResolveOrgID(c, "")
		if err != nil {
			return err
		}

		if name, err = prompt.RequireArg("project-name", name); err != nil {
			return err
		}
		enabled, err := cmdutil.PromptLanguages(defaultLanguage, secondary)
		if err != nil {
			return err
		}

		tResp, err := c.CreateTenantTenantsPostWithResponse(cmdutil.Ctx(), &client.CreateTenantTenantsPostParams{OrganizationId: orgID}, client.CreateTenantRequest{
			Name:             name,
			Description:      cmdutil.StringPtr(description),
			EnabledLanguages: &enabled,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(tResp.HTTPResponse, tResp.Body); err != nil {
			return err
		}
		if tResp.JSON200 == nil {
			return fmt.Errorf("project creation returned no data")
		}
		tenant := *tResp.JSON200

		if refreshErr := cmdutil.RefreshSession(); refreshErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not refresh session: %v\n", refreshErr)
		}
		if c, _, err = cmdutil.NewClientWithAuth(); err != nil {
			return err
		}

		var apiKey *client.PublicApiKeyCreatedResponse
		kResp, err := c.IssueApiKeyApiKeysPostWithResponse(cmdutil.Ctx(), client.IssueApiKeyRequest{
			OrganizationId: orgID,
			TenantId:       tenant.Id,
		})
		if err == nil && kResp.JSON200 != nil {
			apiKey = kResp.JSON200
		}

		if creds, loadErr := credentials.Load(); loadErr == nil {
			creds.TenantID = tenant.Id
			if err := credentials.Save(creds); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not set active project: %v\n", err)
			}
		}

		statusResp, statusErr := c.SetMyOnboardingStatusAuthMeOnboardingPatchWithResponse(cmdutil.Ctx(), client.UpdateOnboardingStatusRequest{
			Status: client.OnboardingStatusCompleted,
		})
		if statusErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not mark onboarding complete: %v\n", statusErr)
		} else if err := cmdutil.CheckErr(statusResp.HTTPResponse, statusResp.Body); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not mark onboarding complete: %v\n", err)
		}

		result := map[string]interface{}{
			"project": map[string]interface{}{
				"id":                tenant.Id,
				"name":              tenant.Name,
				"enabled_languages": tenant.EnabledLanguages,
			},
		}
		if apiKey != nil {
			result["api_key"] = map[string]interface{}{
				"id":    apiKey.Id,
				"value": apiKey.Value,
			}
		}

		output.Print(result, func(d interface{}) {
			fmt.Printf("Project created: %s (%s)\n", tenant.Name, tenant.Id)
			fmt.Printf("Languages: %v\n", tenant.EnabledLanguages)
			if apiKey != nil {
				fmt.Printf("API key: %s\n", apiKey.Value)
				fmt.Println("Store this in your client app's environment (e.g. ADAPTO_API_KEY). View it anytime in Developer Tools -> API Keys.")
			} else {
				fmt.Println("Could not issue an API key. Create one later with 'adapto api-key issue'.")
			}
			fmt.Printf("\nActive project set to %s. Run 'adapto pages list' to confirm access.\n", tenant.Name)
		})
		return nil
	},
}

func init() {
	Cmd.Flags().String("project-name", "", "Project name (required)")
	Cmd.Flags().String("description", "", "Project description")
	Cmd.Flags().String("default-language", "", "Default language code (default en-US)")
	Cmd.Flags().StringSlice("languages", nil, "Additional language codes (comma-separated)")
}
