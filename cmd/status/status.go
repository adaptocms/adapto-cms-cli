package status

import (
	"encoding/json"
	"fmt"

	"github.com/adaptocms/adapto-cms-cli/internal/cmdutil"
	"github.com/adaptocms/adapto-cms-cli/internal/output"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "status",
	Short:   "API status commands",
	Long:    "Check API status. Running 'adapto status' directly returns the API health status.",
	Example: "adapto status --json",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetStatusManageStatusGetWithResponse(cmdutil.Ctx())
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)
		output.Print(data, func(d interface{}) {
			m := d.(map[string]interface{})
			for k, v := range m {
				fmt.Printf("%s: %v\n", k, v)
			}
		})
		return nil
	},
}

func init() {
	Cmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Get API version info",
	Example: "adapto status version --json",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetVersionsManageStatusVersionGetWithResponse(cmdutil.Ctx())
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.HTTPResponse, resp.Body); err != nil {
			return err
		}

		var data map[string]interface{}
		_ = json.Unmarshal(resp.Body, &data)
		output.Print(data, func(d interface{}) {
			m := d.(map[string]interface{})
			for k, v := range m {
				fmt.Printf("%s: %v\n", k, v)
			}
		})
		return nil
	},
}
