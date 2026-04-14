package microcopy

import (
	"fmt"

	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/cmdutil"
	"github.com/eggnita/adapto_cms_cli/internal/output"
	"github.com/eggnita/adapto_cms_cli/internal/prompt"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "microcopy",
	Short: "Manage micro copy entries",
}

func init() {
	Cmd.AddCommand(listCmd, countCmd, createCmd, getCmd, getByKeyCmd, getByLanguageCmd,
		updateCmd, deleteCmd, translationsCmd, createTranslationCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List micro copy entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		language, _ := cmd.Flags().GetString("language")
		tags, _ := cmd.Flags().GetString("tags")

		resp, err := c.ListMicroCopiesManageMicroCopyGetWithResponse(cmdutil.Ctx(), &client.ListMicroCopiesManageMicroCopyGetParams{
			Language: cmdutil.StringPtr(language),
			Tags:     cmdutil.StringPtr(tags),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"ID", "Key", "Value", "Language", "Tags"}
				rows := make([][]string, len(items))
				for i, mc := range items {
					rows[i] = []string{mc.Id, mc.Key, output.Truncate(mc.Value, 50), mc.Language, mc.Tags}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Count micro copy entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		language, _ := cmd.Flags().GetString("language")
		tags, _ := cmd.Flags().GetString("tags")

		resp, err := c.CountMicroCopiesManageMicroCopyCountGetWithResponse(cmdutil.Ctx(), &client.CountMicroCopiesManageMicroCopyCountGetParams{
			Language: cmdutil.StringPtr(language),
			Tags:     cmdutil.StringPtr(tags),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				m := *resp.JSON200
				if count, ok := m["count"]; ok {
					fmt.Printf("Count: %v\n", count)
				} else {
					output.JSON(m)
				}
			})
		}
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a micro copy entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		language, _ := cmd.Flags().GetString("language")
		translationOf, _ := cmd.Flags().GetString("translation-of")
		tags, _ := cmd.Flags().GetString("tags")

		var err error
		if key, err = prompt.RequireArg("key", key); err != nil {
			return err
		}
		if value, err = prompt.RequireArg("value", value); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.MicroCopyCreateModel{
			Key:           key,
			Value:         value,
			Language:      language,
			TranslationOf: cmdutil.StringPtr(translationOf),
			Tags:          cmdutil.StringPtr(tags),
		}

		resp, err := c.CreateMicroCopyManageMicroCopyPostWithResponse(cmdutil.Ctx(), &client.CreateMicroCopyManageMicroCopyPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printMicroCopy(resp.JSON201)
			})
		}
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get micro copy by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetMicroCopyManageMicroCopyMicroCopyIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetMicroCopyManageMicroCopyMicroCopyIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printMicroCopy(resp.JSON200)
			})
		}
		return nil
	},
}

var getByKeyCmd = &cobra.Command{
	Use:   "get-by-key <key>",
	Short: "Get micro copy by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		language, _ := cmd.Flags().GetString("language")

		resp, err := c.GetMicroCopyByKeyManageMicroCopyByKeyKeyGetWithResponse(cmdutil.Ctx(), args[0], &client.GetMicroCopyByKeyManageMicroCopyByKeyKeyGetParams{
			Language: cmdutil.StringPtr(language),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printMicroCopy(resp.JSON200)
			})
		}
		return nil
	},
}

var getByLanguageCmd = &cobra.Command{
	Use:   "get-by-language <language>",
	Short: "Get all micro copy for a language",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetMicroCopiesByLanguageManageMicroCopyLanguageLanguageGetWithResponse(cmdutil.Ctx(), args[0], &client.GetMicroCopiesByLanguageManageMicroCopyLanguageLanguageGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"ID", "Key", "Value", "Language", "Tags"}
				rows := make([][]string, len(items))
				for i, mc := range items {
					rows[i] = []string{mc.Id, mc.Key, output.Truncate(mc.Value, 50), mc.Language, mc.Tags}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a micro copy entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.MicroCopyUpdateModel{}
		if v, _ := cmd.Flags().GetString("key"); v != "" {
			body.Key = &v
		}
		if v, _ := cmd.Flags().GetString("value"); v != "" {
			body.Value = &v
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			body.Language = &v
		}
		if v, _ := cmd.Flags().GetString("tags"); v != "" {
			body.Tags = &v
		}

		resp, err := c.UpdateMicroCopyManageMicroCopyMicroCopyIdPutWithResponse(cmdutil.Ctx(), args[0], &client.UpdateMicroCopyManageMicroCopyMicroCopyIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printMicroCopy(resp.JSON200)
			})
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a micro copy entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteMicroCopyManageMicroCopyMicroCopyIdDeleteWithResponse(cmdutil.Ctx(), args[0], &client.DeleteMicroCopyManageMicroCopyMicroCopyIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Micro copy entry deleted.")
		return nil
	},
}

var translationsCmd = &cobra.Command{
	Use:   "translations <id>",
	Short: "List translations of a micro copy entry",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetMicroCopyTranslationsManageMicroCopyTranslationsMicroCopyIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetMicroCopyTranslationsManageMicroCopyTranslationsMicroCopyIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"ID", "Key", "Value", "Language"}
				rows := make([][]string, len(items))
				for i, mc := range items {
					rows[i] = []string{mc.Id, mc.Key, output.Truncate(mc.Value, 50), mc.Language}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var createTranslationCmd = &cobra.Command{
	Use:   "create-translation <source_id>",
	Short: "Create a micro copy translation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")
		language, _ := cmd.Flags().GetString("language")
		tags, _ := cmd.Flags().GetString("tags")

		var err error
		if key, err = prompt.RequireArg("key", key); err != nil {
			return err
		}
		if value, err = prompt.RequireArg("value", value); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.MicroCopyCreateModel{
			Key:      key,
			Value:    value,
			Language: language,
			Tags:     cmdutil.StringPtr(tags),
		}

		resp, err := c.CreateMicroCopyTranslationManageMicroCopySourceIdTranslationsPostWithResponse(cmdutil.Ctx(), args[0], &client.CreateMicroCopyTranslationManageMicroCopySourceIdTranslationsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printMicroCopy(resp.JSON201)
			})
		}
		return nil
	},
}

func printMicroCopy(mc *client.MicroCopyResponseModel) {
	translationOf := ""
	if mc.TranslationOf != nil {
		translationOf = *mc.TranslationOf
	}
	pairs := [][2]string{
		{"ID", mc.Id},
		{"Key", mc.Key},
		{"Value", mc.Value},
		{"Language", mc.Language},
		{"Tags", mc.Tags},
		{"Translation Of", translationOf},
		{"Created", mc.CreatedAt},
		{"Updated", mc.UpdatedAt},
	}
	output.KeyValue(pairs)
}

func init() {
	listCmd.Flags().String("language", "", "Filter by language")
	listCmd.Flags().String("tags", "", "Filter by tags")

	countCmd.Flags().String("language", "", "Filter by language")
	countCmd.Flags().String("tags", "", "Filter by tags")

	for _, c := range []*cobra.Command{createCmd, createTranslationCmd} {
		c.Flags().String("key", "", "Micro copy key")
		c.Flags().String("value", "", "Text value")
		c.Flags().String("language", "", "Language code")
		c.Flags().String("tags", "", "Comma-separated tags")
	}
	createCmd.Flags().String("translation-of", "", "Source micro copy ID")

	updateCmd.Flags().String("key", "", "Micro copy key")
	updateCmd.Flags().String("value", "", "Text value")
	updateCmd.Flags().String("language", "", "Language code")
	updateCmd.Flags().String("tags", "", "Tags")

	getByKeyCmd.Flags().String("language", "", "Filter by language")
}
