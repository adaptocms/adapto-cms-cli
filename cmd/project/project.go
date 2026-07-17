package project

import (
	"fmt"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	"github.com/adaptocms/adapto-cms-cli/internal/output"
	"github.com/adaptocms/adapto-cms-cli/internal/prompt"
	"github.com/spf13/cobra"
)

// Cmd is the root project command.
var Cmd = &cobra.Command{
	Use:   "project",
	Short: "Manage projects",
}

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create a project",
	Example: "adapto project create --name \"My Project\" --default-language en-US --languages es-ES,fr-FR",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")
		orgID, _ := cmd.Flags().GetString("org-id")
		defaultLanguage, _ := cmd.Flags().GetString("default-language")
		secondary, _ := cmd.Flags().GetStringSlice("languages")

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		if orgID, err = cmdutil.ResolveOrgID(c, orgID); err != nil {
			return err
		}
		if name, err = prompt.RequireArg("name", name); err != nil {
			return err
		}
		enabled, err := cmdutil.PromptLanguages(defaultLanguage, secondary)
		if err != nil {
			return err
		}

		resp, err := c.CreateTenantTenantsPostWithResponse(cmdutil.Ctx(), &client.CreateTenantTenantsPostParams{OrganizationId: orgID}, client.CreateTenantRequest{
			Name:             name,
			Description:      cmdutil.StringPtr(description),
			EnabledLanguages: &enabled,
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}
		if resp.JSON200 == nil {
			return fmt.Errorf("project creation returned no data")
		}

		tenant := *resp.JSON200

		if refreshErr := cmdutil.RefreshSession(); refreshErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not refresh session: %v\n", refreshErr)
		}

		output.Print(tenant, func(d interface{}) {
			fmt.Printf("Project created: %s (%s)\n", tenant.Name, tenant.Id)
			fmt.Printf("Set it active with 'adapto project use %s'\n", tenant.Id)
		})
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List your projects",
	Example: "adapto project list --json",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		orgsResp, err := c.ListMyOrgsOrgsGetWithResponse(cmdutil.Ctx())
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(orgsResp.HTTPResponse, orgsResp.Body); err != nil {
			return err
		}

		activeTenant := ""
		if creds, err := credentials.Load(); err == nil {
			activeTenant = creds.TenantID
		}

		type projectEntry struct {
			ID        string   `json:"id"`
			Name      string   `json:"name"`
			OrgName   string   `json:"org_name"`
			Languages []string `json:"languages"`
			Active    bool     `json:"active"`
		}

		var entries []projectEntry
		if orgsResp.JSON200 != nil {
			for _, o := range *orgsResp.JSON200 {
				tResp, err := c.ListOrgTenantsTenantsByOrgOrgIdGetWithResponse(cmdutil.Ctx(), o.Id)
				if err != nil || tResp.JSON200 == nil {
					continue
				}
				for _, t := range *tResp.JSON200 {
					entries = append(entries, projectEntry{
						ID:        t.Id,
						Name:      t.Name,
						OrgName:   o.Name,
						Languages: t.EnabledLanguages,
						Active:    t.Id == activeTenant,
					})
				}
			}
		}

		output.Print(entries, func(d interface{}) {
			if len(entries) == 0 {
				fmt.Println("No projects found. Create one with 'adapto project create' or 'adapto onboard'.")
				return
			}
			for _, e := range entries {
				marker := ""
				if e.Active {
					marker = " (active)"
				}
				fmt.Printf("%s  %s — %s%s  %v\n", e.ID, e.Name, e.OrgName, marker, e.Languages)
			}
		})
		return nil
	},
}

var useCmd = &cobra.Command{
	Use:     "use [project-id]",
	Short:   "Set the active project",
	Example: "adapto project use <project-id>",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project-id")
		if len(args) > 0 {
			projectID = args[0]
		}

		var err error
		if projectID, err = prompt.RequireArg("project-id", projectID); err != nil {
			return err
		}

		creds, err := credentials.Load()
		if err != nil || creds.AccessToken == "" {
			return fmt.Errorf("not logged in: run 'adapto auth login' first")
		}
		creds.TenantID = projectID
		if err := credentials.Save(creds); err != nil {
			return fmt.Errorf("could not save credentials: %w", err)
		}

		output.Successf("Active project set to %s", projectID)
		return nil
	},
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(useCmd)

	createCmd.Flags().String("name", "", "Project name (required)")
	createCmd.Flags().String("description", "", "Project description")
	createCmd.Flags().String("org-id", "", "Organization ID (defaults to your only organization)")
	createCmd.Flags().String("default-language", "", "Default language code (default en-US)")
	createCmd.Flags().StringSlice("languages", nil, "Additional language codes (comma-separated)")

	useCmd.Flags().String("project-id", "", "Project ID to set active")
}
