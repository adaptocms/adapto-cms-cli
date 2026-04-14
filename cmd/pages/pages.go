package pages

import (
	"fmt"
	"strings"

	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/cmdutil"
	"github.com/eggnita/adapto_cms_cli/internal/output"
	"github.com/eggnita/adapto_cms_cli/internal/prompt"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pages",
	Short: "Manage pages",
}

func init() {
	Cmd.AddCommand(listCmd, createCmd, getCmd, getBySlugCmd, updateCmd, deleteCmd,
		publishCmd, archiveCmd, translationsCmd, createTranslationCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List pages",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		keyword, _ := cmd.Flags().GetString("keyword")
		language, _ := cmd.Flags().GetString("language")
		field, _ := cmd.Flags().GetString("field")
		tag, _ := cmd.Flags().GetString("tag")
		order, _ := cmd.Flags().GetString("order")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		resp, err := c.ListPagesManagePagesGetWithResponse(cmdutil.Ctx(), &client.ListPagesManagePagesGetParams{
			Status:   cmdutil.StringPtr(status),
			Keyword:  cmdutil.StringPtr(keyword),
			Language: cmdutil.StringPtr(language),
			Field:    cmdutil.StringPtr(field),
			Tag:      cmdutil.StringPtr(tag),
			Order:    cmdutil.StringPtr(order),
			Page:     cmdutil.IntPtr(page),
			Limit:    cmdutil.IntPtr(limit),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				data := resp.JSON200
				fmt.Printf("Total: %d (page %d/%d)\n\n", data.Total, data.Page, data.Pages)
				headers := []string{"ID", "Title", "Status", "Language", "Slug"}
				rows := make([][]string, len(data.Items))
				for i, p := range data.Items {
					rows[i] = []string{p.Id, output.Truncate(p.Title, 40), string(p.Status), p.Language, output.Truncate(p.Slug, 30)}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a page",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		slug, _ := cmd.Flags().GetString("slug")
		menuLabel, _ := cmd.Flags().GetString("menu-label")
		parentID, _ := cmd.Flags().GetString("parent-id")
		language, _ := cmd.Flags().GetString("language")
		status, _ := cmd.Flags().GetString("status")
		tags, _ := cmd.Flags().GetString("tags")

		var err error
		if title, err = prompt.RequireArg("title", title); err != nil {
			return err
		}
		if content, err = prompt.RequireArg("content", content); err != nil {
			return err
		}
		if slug, err = prompt.RequireArg("slug", slug); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.PageCreateModel{
			Title:     title,
			Content:   content,
			Slug:      slug,
			Language:  language,
			MenuLabel: cmdutil.StringPtr(menuLabel),
			ParentId:  cmdutil.StringPtr(parentID),
			Tags:      cmdutil.StringSlicePtr(tags),
		}
		if status != "" {
			s := client.PageStatus(status)
			body.Status = &s
		}

		resp, err := c.CreatePageManagePagesPostWithResponse(cmdutil.Ctx(), &client.CreatePageManagePagesPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printPage(resp.JSON201)
			})
		}
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a page by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetPageManagePagesPageIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetPageManagePagesPageIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printPage(resp.JSON200)
			})
		}
		return nil
	},
}

var getBySlugCmd = &cobra.Command{
	Use:   "get-by-slug <slug>",
	Short: "Get a page by slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetPageBySlugManagePagesBySlugSlugGetWithResponse(cmdutil.Ctx(), args[0], &client.GetPageBySlugManagePagesBySlugSlugGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printPage(resp.JSON200)
			})
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.PageUpdateModel{}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			body.Title = &v
		}
		if v, _ := cmd.Flags().GetString("content"); v != "" {
			body.Content = &v
		}
		if v, _ := cmd.Flags().GetString("slug"); v != "" {
			body.Slug = &v
		}
		if v, _ := cmd.Flags().GetString("menu-label"); v != "" {
			body.MenuLabel = &v
		}
		if v, _ := cmd.Flags().GetString("parent-id"); v != "" {
			body.ParentId = &v
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			body.Language = &v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			s := client.PageStatus(v)
			body.Status = &s
		}
		if v, _ := cmd.Flags().GetString("tags"); v != "" {
			body.Tags = cmdutil.StringSlicePtr(v)
		}

		resp, err := c.UpdatePageManagePagesPageIdPutWithResponse(cmdutil.Ctx(), args[0], &client.UpdatePageManagePagesPageIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printPage(resp.JSON200)
			})
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeletePageManagePagesPageIdDeleteWithResponse(cmdutil.Ctx(), args[0], &client.DeletePageManagePagesPageIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Page deleted.")
		return nil
	},
}

var publishCmd = &cobra.Command{
	Use:   "publish <id>",
	Short: "Publish a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.PublishPageManagePagesPageIdPublishPostWithResponse(cmdutil.Ctx(), args[0], &client.PublishPageManagePagesPageIdPublishPostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Page published.")
		return nil
	},
}

var archiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.ArchivePageManagePagesPageIdArchivePostWithResponse(cmdutil.Ctx(), args[0], &client.ArchivePageManagePagesPageIdArchivePostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Page archived.")
		return nil
	},
}

var translationsCmd = &cobra.Command{
	Use:   "translations <id>",
	Short: "List translations of a page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetPageTranslationsManagePagesPageIdTranslationsGetWithResponse(cmdutil.Ctx(), args[0], &client.GetPageTranslationsManagePagesPageIdTranslationsGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"ID", "Title", "Language", "Status", "Slug"}
				rows := make([][]string, len(items))
				for i, p := range items {
					rows[i] = []string{p.Id, output.Truncate(p.Title, 40), p.Language, string(p.Status), p.Slug}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var createTranslationCmd = &cobra.Command{
	Use:   "create-translation <source_id>",
	Short: "Create a page translation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		slug, _ := cmd.Flags().GetString("slug")
		language, _ := cmd.Flags().GetString("language")
		menuLabel, _ := cmd.Flags().GetString("menu-label")
		parentID, _ := cmd.Flags().GetString("parent-id")
		tags, _ := cmd.Flags().GetString("tags")

		var err error
		if title, err = prompt.RequireArg("title", title); err != nil {
			return err
		}
		if content, err = prompt.RequireArg("content", content); err != nil {
			return err
		}
		if slug, err = prompt.RequireArg("slug", slug); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.PageCreateModel{
			Title:     title,
			Content:   content,
			Slug:      slug,
			Language:  language,
			MenuLabel: cmdutil.StringPtr(menuLabel),
			ParentId:  cmdutil.StringPtr(parentID),
			Tags:      cmdutil.StringSlicePtr(tags),
		}

		resp, err := c.CreatePageTranslationManagePagesSourceIdTranslationsPostWithResponse(cmdutil.Ctx(), args[0], &client.CreatePageTranslationManagePagesSourceIdTranslationsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printPage(resp.JSON201)
			})
		}
		return nil
	},
}

func printPage(p *client.PageResponseModel) {
	pairs := [][2]string{
		{"ID", p.Id},
		{"Title", p.Title},
		{"Slug", p.Slug},
		{"Status", string(p.Status)},
		{"Language", p.Language},
		{"Tags", strings.Join(p.Tags, ", ")},
		{"Created", p.CreatedAt},
		{"Updated", p.UpdatedAt},
	}
	if p.MenuLabel != nil {
		pairs = append(pairs, [2]string{"Menu Label", *p.MenuLabel})
	}
	if p.ParentId != nil {
		pairs = append(pairs, [2]string{"Parent ID", *p.ParentId})
	}
	if p.PublishedAt != nil {
		pairs = append(pairs, [2]string{"Published", *p.PublishedAt})
	}
	output.KeyValue(pairs)
}

func init() {
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().String("tag", "", "Filter by tag")
	cmdutil.AddListFlags(listCmd)

	for _, c := range []*cobra.Command{createCmd, createTranslationCmd} {
		c.Flags().String("title", "", "Page title")
		c.Flags().String("content", "", "Page content")
		c.Flags().String("slug", "", "URL-friendly identifier")
		c.Flags().String("menu-label", "", "Menu label")
		c.Flags().String("parent-id", "", "Parent page ID")
		c.Flags().String("language", "", "Language code")
		c.Flags().String("status", "", "Status")
		c.Flags().String("tags", "", "Comma-separated tags")
	}

	updateCmd.Flags().String("title", "", "Page title")
	updateCmd.Flags().String("content", "", "Page content")
	updateCmd.Flags().String("slug", "", "URL-friendly identifier")
	updateCmd.Flags().String("menu-label", "", "Menu label")
	updateCmd.Flags().String("parent-id", "", "Parent page ID")
	updateCmd.Flags().String("language", "", "Language code")
	updateCmd.Flags().String("status", "", "Status")
	updateCmd.Flags().String("tags", "", "Comma-separated tags")
}
