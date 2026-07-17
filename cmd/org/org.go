package org

import (
	"fmt"

	"github.com/adaptocms/adapto-cms-cli/internal/client"
	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/adaptocms/adapto-cms-cli/internal/output"
	"github.com/adaptocms/adapto-cms-cli/internal/prompt"
	"github.com/spf13/cobra"
)

// Cmd is the root org command.
var Cmd = &cobra.Command{
	Use:   "org",
	Short: "Manage organizations",
}

var createCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create an organization",
	Example: "adapto org create --name \"My Org\" --description \"Marketing content\"",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		description, _ := cmd.Flags().GetString("description")

		var err error
		if name, err = prompt.RequireArg("name", name); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.CreateOrgOrgsPostWithResponse(cmdutil.Ctx(), client.CreateOrgRequest{
			Name:        name,
			Description: cmdutil.StringPtr(description),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}
		if resp.JSON200 == nil {
			return fmt.Errorf("organization creation returned no data")
		}

		org := *resp.JSON200

		if refreshErr := cmdutil.RefreshSession(); refreshErr != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not refresh session: %v\n", refreshErr)
		}

		output.Print(org, func(d interface{}) {
			fmt.Printf("Organization created: %s (%s)\n", org.Name, org.Id)
		})
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List your organizations",
	Example: "adapto org list --json",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.ListMyOrgsOrgsGetWithResponse(cmdutil.Ctx())
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}

		orgs := []client.Organization{}
		if resp.JSON200 != nil {
			orgs = *resp.JSON200
		}
		output.Print(orgs, func(d interface{}) {
			if len(orgs) == 0 {
				fmt.Println("No organizations found.")
				return
			}
			for _, o := range orgs {
				fmt.Printf("%s  %s\n", o.Id, o.Name)
			}
		})
		return nil
	},
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(listCmd)

	createCmd.Flags().String("name", "", "Organization name (required)")
	createCmd.Flags().String("description", "", "Organization description")
}
