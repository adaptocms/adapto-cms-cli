package articles

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/cmdutil"
	"github.com/eggnita/adapto_cms_cli/internal/output"
	"github.com/eggnita/adapto_cms_cli/internal/prompt"
	"github.com/spf13/cobra"
)

// Cmd is the root articles command.
var Cmd = &cobra.Command{
	Use:   "articles",
	Short: "Manage articles",
}

func init() {
	Cmd.AddCommand(listCmd, createCmd, getCmd, getBySlugCmd, updateCmd, deleteCmd,
		publishCmd, archiveCmd, translationsCmd, createTranslationCmd, categoriesCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List articles",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		category, _ := cmd.Flags().GetString("category")
		tag, _ := cmd.Flags().GetString("tag")
		keyword, _ := cmd.Flags().GetString("keyword")
		language, _ := cmd.Flags().GetString("language")
		field, _ := cmd.Flags().GetString("field")
		order, _ := cmd.Flags().GetString("order")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		resp, err := c.ListArticlesManageArticlesGetWithResponse(cmdutil.Ctx(), &client.ListArticlesManageArticlesGetParams{
			Status:   cmdutil.StringPtr(status),
			Category: cmdutil.StringPtr(category),
			Tag:      cmdutil.StringPtr(tag),
			Keyword:  cmdutil.StringPtr(keyword),
			Language: cmdutil.StringPtr(language),
			Field:    cmdutil.StringPtr(field),
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
				headers := []string{"ID", "Title", "Status", "Language", "Slug", "Author"}
				rows := make([][]string, len(data.Items))
				for i, a := range data.Items {
					rows[i] = []string{a.Id, output.Truncate(a.Title, 40), string(a.Status), a.Language, output.Truncate(a.Slug, 30), a.Author}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an article",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		slug, _ := cmd.Flags().GetString("slug")
		author, _ := cmd.Flags().GetString("author")
		summary, _ := cmd.Flags().GetString("summary")
		language, _ := cmd.Flags().GetString("language")
		status, _ := cmd.Flags().GetString("status")
		tags, _ := cmd.Flags().GetString("tags")
		source, _ := cmd.Flags().GetString("source")

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
		if author, err = prompt.RequireArg("author", author); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.ArticleCreateModel{
			Title:    title,
			Content:  content,
			Slug:     slug,
			Author:   author,
			Summary:  summary,
			Language: language,
			Tags:     cmdutil.StringSlicePtr(tags),
		}
		if status != "" {
			s := client.ArticleStatus(status)
			body.Status = &s
		}
		if source != "" {
			var src client.ArticleSourceModel
			if err := json.Unmarshal([]byte(source), &src); err != nil {
				return fmt.Errorf("invalid --source JSON: %w", err)
			}
			body.Source = src
		} else {
			body.Source = client.ArticleSourceModel{Type: "internal", Name: "CLI"}
		}

		resp, err := c.CreateArticleManageArticlesPostWithResponse(cmdutil.Ctx(), &client.CreateArticleManageArticlesPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printArticle(resp.JSON201)
			})
		}
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get an article by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetArticleManageArticlesArticleIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetArticleManageArticlesArticleIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printArticle(resp.JSON200)
			})
		}
		return nil
	},
}

var getBySlugCmd = &cobra.Command{
	Use:   "get-by-slug <slug>",
	Short: "Get an article by slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetArticleBySlugManageArticlesBySlugSlugGetWithResponse(cmdutil.Ctx(), args[0], &client.GetArticleBySlugManageArticlesBySlugSlugGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printArticle(resp.JSON200)
			})
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.ArticleUpdateModel{}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			body.Title = &v
		}
		if v, _ := cmd.Flags().GetString("content"); v != "" {
			body.Content = &v
		}
		if v, _ := cmd.Flags().GetString("slug"); v != "" {
			body.Slug = &v
		}
		if v, _ := cmd.Flags().GetString("author"); v != "" {
			body.Author = &v
		}
		if v, _ := cmd.Flags().GetString("summary"); v != "" {
			body.Summary = &v
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			body.Language = &v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			s := client.ArticleStatus(v)
			body.Status = &s
		}
		if v, _ := cmd.Flags().GetString("tags"); v != "" {
			body.Tags = cmdutil.StringSlicePtr(v)
		}
		if v, _ := cmd.Flags().GetString("source"); v != "" {
			var src client.ArticleSourceModel
			if err := json.Unmarshal([]byte(v), &src); err != nil {
				return fmt.Errorf("invalid --source JSON: %w", err)
			}
			body.Source = &src
		}

		resp, err := c.UpdateArticleManageArticlesArticleIdPutWithResponse(cmdutil.Ctx(), args[0], &client.UpdateArticleManageArticlesArticleIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printArticle(resp.JSON200)
			})
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteArticleManageArticlesArticleIdDeleteWithResponse(cmdutil.Ctx(), args[0], &client.DeleteArticleManageArticlesArticleIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Article deleted.")
		return nil
	},
}

var publishCmd = &cobra.Command{
	Use:   "publish <id>",
	Short: "Publish an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.PublishArticleManageArticlesArticleIdPublishPostWithResponse(cmdutil.Ctx(), args[0], &client.PublishArticleManageArticlesArticleIdPublishPostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Article published.")
		return nil
	},
}

var archiveCmd = &cobra.Command{
	Use:   "archive <id>",
	Short: "Archive an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.ArchiveArticleManageArticlesArticleIdArchivePostWithResponse(cmdutil.Ctx(), args[0], &client.ArchiveArticleManageArticlesArticleIdArchivePostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Article archived.")
		return nil
	},
}

var translationsCmd = &cobra.Command{
	Use:   "translations <id>",
	Short: "List translations of an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetTranslationsManageArticlesArticleIdTranslationsGetWithResponse(cmdutil.Ctx(), args[0], &client.GetTranslationsManageArticlesArticleIdTranslationsGetParams{})
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
				for i, a := range items {
					rows[i] = []string{a.Id, output.Truncate(a.Title, 40), a.Language, string(a.Status), a.Slug}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var createTranslationCmd = &cobra.Command{
	Use:   "create-translation <source_id>",
	Short: "Create an article translation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		content, _ := cmd.Flags().GetString("content")
		slug, _ := cmd.Flags().GetString("slug")
		author, _ := cmd.Flags().GetString("author")
		summary, _ := cmd.Flags().GetString("summary")
		language, _ := cmd.Flags().GetString("language")
		tags, _ := cmd.Flags().GetString("tags")
		source, _ := cmd.Flags().GetString("source")

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
		if author, err = prompt.RequireArg("author", author); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.ArticleCreateModel{
			Title:    title,
			Content:  content,
			Slug:     slug,
			Author:   author,
			Summary:  summary,
			Language: language,
			Tags:     cmdutil.StringSlicePtr(tags),
		}
		if source != "" {
			var src client.ArticleSourceModel
			if err := json.Unmarshal([]byte(source), &src); err != nil {
				return fmt.Errorf("invalid --source JSON: %w", err)
			}
			body.Source = src
		} else {
			body.Source = client.ArticleSourceModel{Type: "internal", Name: "CLI"}
		}

		resp, err := c.CreateTranslationManageArticlesSourceIdTranslationsPostWithResponse(cmdutil.Ctx(), args[0], &client.CreateTranslationManageArticlesSourceIdTranslationsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printArticle(resp.JSON201)
			})
		}
		return nil
	},
}

var categoriesCmd = &cobra.Command{
	Use:   "categories <id>",
	Short: "List categories of an article",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetArticleCategoriesManageArticlesArticleIdCategoriesGetWithResponse(cmdutil.Ctx(), args[0], &client.GetArticleCategoriesManageArticlesArticleIdCategoriesGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"Category ID"}
				rows := make([][]string, len(items))
				for i, id := range items {
					rows[i] = []string{id}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

func printArticle(a *client.ArticleResponseModel) {
	pairs := [][2]string{
		{"ID", a.Id},
		{"Title", a.Title},
		{"Slug", a.Slug},
		{"Status", string(a.Status)},
		{"Language", a.Language},
		{"Author", a.Author},
		{"Summary", a.Summary},
		{"Tags", strings.Join(a.Tags, ", ")},
	}
	if a.CreatedAt != nil {
		pairs = append(pairs, [2]string{"Created", *a.CreatedAt})
	}
	if a.UpdatedAt != nil {
		pairs = append(pairs, [2]string{"Updated", *a.UpdatedAt})
	}
	if a.PublishedAt != nil {
		pairs = append(pairs, [2]string{"Published", *a.PublishedAt})
	}
	output.KeyValue(pairs)
}

func init() {
	listCmd.Flags().String("status", "", "Filter by status")
	listCmd.Flags().String("category", "", "Filter by category ID")
	listCmd.Flags().String("tag", "", "Filter by tag")
	cmdutil.AddListFlags(listCmd)

	for _, c := range []*cobra.Command{createCmd, createTranslationCmd} {
		c.Flags().String("title", "", "Article title")
		c.Flags().String("content", "", "Article content")
		c.Flags().String("slug", "", "URL-friendly identifier")
		c.Flags().String("author", "", "Author name")
		c.Flags().String("summary", "", "Article summary")
		c.Flags().String("language", "", "Language code")
		c.Flags().String("status", "", "Status (draft/published)")
		c.Flags().String("tags", "", "Comma-separated tags")
		c.Flags().String("source", "", "Source JSON")
	}

	updateCmd.Flags().String("title", "", "Article title")
	updateCmd.Flags().String("content", "", "Article content")
	updateCmd.Flags().String("slug", "", "URL-friendly identifier")
	updateCmd.Flags().String("author", "", "Author name")
	updateCmd.Flags().String("summary", "", "Article summary")
	updateCmd.Flags().String("language", "", "Language code")
	updateCmd.Flags().String("status", "", "Status")
	updateCmd.Flags().String("tags", "", "Comma-separated tags")
	updateCmd.Flags().String("source", "", "Source JSON")
}
