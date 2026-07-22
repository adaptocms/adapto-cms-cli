package project

import (
	"fmt"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/adaptocms/adapto-cms-cli/internal/credentials"
	"github.com/adaptocms/adapto-cms-cli/internal/output"
	"github.com/adaptocms/adapto-cms-cli/internal/prompt"
	"github.com/charmbracelet/lipgloss"
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
	Use:   "use [project-id]",
	Short: "Set the active project",
	Long: `Set the active project used by content and API-key commands.

Run without arguments to pick from a selector listing every project across all
your organizations. Pass a project id (positionally or with --project-id) to
set it directly; the id is validated before it is saved.`,
	Example: "  adapto project use\n  adapto project use <project-id>",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project-id")
		if len(args) > 0 {
			projectID = args[0]
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		if projectID == "" {
			if !prompt.IsTTY() {
				return fmt.Errorf("no project id given: pass a project id (or --project-id) in a non-interactive shell")
			}
			if projectID, err = cmdutil.SelectProjectAllOrgs(c); err != nil {
				return err
			}
		} else {
			tResp, err := c.GetTenantTenantsTenantIdGetWithResponse(cmdutil.Ctx(), projectID)
			if err != nil {
				return err
			}
			if err := cmdutil.CheckErr(tResp.HTTPResponse, tResp.Body); err != nil {
				return fmt.Errorf("project %s not found or not accessible: %w", projectID, err)
			}
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

var updateCmd = &cobra.Command{
	Use:   "update [project-id]",
	Short: "Update a project",
	Long: `Update a project's name, description, or languages. Only provided flags are changed.

Run without a project id to pick one from a selector.`,
	Example: "  adapto project update --name \"New Name\"\n" +
		"  adapto project update <project-id> --description \"Marketing content\" --languages en-US,fr-FR",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		body := client.UpdateTenantRequest{}
		changed := false
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			body.Name = &v
			changed = true
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			body.Description = &v
			changed = true
		}
		if v, _ := cmd.Flags().GetStringSlice("languages"); len(v) > 0 {
			body.EnabledLanguages = &v
			changed = true
		}
		if !changed {
			return fmt.Errorf("nothing to update: pass --name, --description, or --languages")
		}

		projectID, _ := cmd.Flags().GetString("project-id")
		if len(args) > 0 {
			projectID = args[0]
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		if projectID == "" {
			if !prompt.IsTTY() {
				return fmt.Errorf("no project id given: pass a project id (or --project-id) in a non-interactive shell")
			}
			if projectID, err = cmdutil.SelectProjectAllOrgs(c); err != nil {
				return err
			}
		}

		resp, err := c.UpdateTenantTenantsTenantIdPatchWithResponse(cmdutil.Ctx(), projectID, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			tenant := *resp.JSON200
			output.Print(tenant, func(d interface{}) {
				fmt.Printf("Project updated: %s (%s)\n", tenant.Name, tenant.Id)
			})
		} else {
			output.Successf("Project %s updated.", projectID)
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [project-id]",
	Short: "Delete a project and all its content",
	Long: `Delete a project, permanently destroying all content scoped to it.

Run without arguments to pick a project from an interactive selector; you are
then asked to type the project name to confirm before anything is deleted.

Pass a project id (positionally or with --project-id) to delete it immediately,
with no confirmation prompt — the explicit id is treated as the confirmation.
This is the form for scripts and AI agents. In a non-interactive shell you must
pass an id, since the selector cannot open.`,
	Example: "  # Pick from a selector, then type the name to confirm\n" +
		"  adapto project delete\n\n" +
		"  # Delete immediately by id, no prompt\n" +
		"  adapto project delete <project-id>\n" +
		"  adapto project delete --project-id <project-id>",
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, _ := cmd.Flags().GetString("project-id")
		if len(args) > 0 {
			projectID = args[0]
		}

		c, cfg, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		if projectID != "" {
			return deleteProject(cmd, c, projectID)
		}

		if !prompt.IsTTY() {
			return fmt.Errorf("no project id given: pass a project id (or --project-id) in a non-interactive shell")
		}

		tenant, err := cmdutil.SelectProjectInActiveOrg(c, cfg.TenantID)
		if err != nil {
			return err
		}

		printDeleteWarning(cmd, tenant)
		for {
			typed, err := prompt.AskString("Type the project name to confirm (leave empty to cancel):")
			if err != nil {
				return err
			}
			if typed == tenant.Name {
				break
			}
			if typed == "" {
				output.Success("Cancelled; project not deleted.")
				return nil
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "That does not match %q. Try again, or leave empty to cancel.\n", tenant.Name)
		}

		return deleteProject(cmd, c, tenant.Id)
	},
}

func printDeleteWarning(cmd *cobra.Command, t *client.Tenant) {
	red := lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	w := cmd.ErrOrStderr()
	fmt.Fprintln(w)
	fmt.Fprintln(w, red.Render(fmt.Sprintf("⚠  Deleting project %q permanently removes it and ALL of its content —", t.Name)))
	fmt.Fprintln(w, red.Render("   articles, pages, collections, files, micro copy, and API keys. This cannot be undone."))
	fmt.Fprintln(w)
}

func deleteProject(cmd *cobra.Command, c *client.ClientWithResponses, id string) error {
	resp, err := c.DeleteTenantTenantsTenantIdDeleteWithResponse(cmdutil.Ctx(), id)
	if err != nil {
		return err
	}
	if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
		return err
	}

	output.Successf("Project %s deleted.", id)

	if creds, err := credentials.Load(); err == nil && creds.TenantID == id {
		creds.TenantID = ""
		if err := credentials.Save(creds); err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not clear active project: %v\n", err)
		} else if refreshErr := cmdutil.RefreshSession(); refreshErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not refresh session: %v\n", refreshErr)
		}
		output.Success("This was your active project. Run 'adapto project use' to pick another.")
	}
	return nil
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)

	createCmd.Flags().String("name", "", "Project name (required)")
	createCmd.Flags().String("description", "", "Project description")
	createCmd.Flags().String("org-id", "", "Organization ID (defaults to your only organization)")
	createCmd.Flags().String("default-language", "", "Default language code (default en-US)")
	createCmd.Flags().StringSlice("languages", nil, "Additional language codes (comma-separated)")

	useCmd.Flags().String("project-id", "", "Project ID to set active (opens a selector if omitted)")

	updateCmd.Flags().String("name", "", "New project name")
	updateCmd.Flags().String("description", "", "New project description")
	updateCmd.Flags().StringSlice("languages", nil, "Replacement language codes (comma-separated)")
	updateCmd.Flags().String("project-id", "", "Project ID to update (opens a selector if omitted)")

	deleteCmd.Flags().String("project-id", "", "Project ID to delete immediately, without confirmation (opens a selector if omitted)")
}
