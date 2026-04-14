package categories

import (
	"encoding/json"
	"fmt"

	"github.com/eggnita/adapto_cms_cli/internal/client"
	"github.com/eggnita/adapto_cms_cli/internal/cmdutil"
	"github.com/eggnita/adapto_cms_cli/internal/output"
	"github.com/eggnita/adapto_cms_cli/internal/prompt"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "categories",
	Short: "Manage categories",
}

func init() {
	Cmd.AddCommand(listCmd, createCmd, getCmd, getBySlugCmd, updateCmd, deleteCmd,
		subcategoriesCmd, articlesCmd, addArticleCmd, removeArticleCmd,
		translationsCmd, createTranslationCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		parentID, _ := cmd.Flags().GetString("parent-id")
		language, _ := cmd.Flags().GetString("language")
		keyword, _ := cmd.Flags().GetString("keyword")
		field, _ := cmd.Flags().GetString("field")
		order, _ := cmd.Flags().GetString("order")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		resp, err := c.ListCategoriesManageCategoriesGetWithResponse(cmdutil.Ctx(), &client.ListCategoriesManageCategoriesGetParams{
			ParentId: cmdutil.StringPtr(parentID),
			Language: cmdutil.StringPtr(language),
			Keyword:  cmdutil.StringPtr(keyword),
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

		// resp.JSON200 is *PaginatedResponse with Items []interface{}, so unmarshal Body manually.
		var paginated struct {
			Items []client.CategoryResponseModel `json:"items"`
			Total int                            `json:"total"`
			Page  int                            `json:"page"`
			Pages int                            `json:"pages"`
		}
		if err := json.Unmarshal(resp.Body, &paginated); err != nil {
			return fmt.Errorf("failed to parse categories list: %w", err)
		}
		output.Print(paginated, func(d interface{}) {
			fmt.Printf("Total: %d (page %d/%d)\n\n", paginated.Total, paginated.Page, paginated.Pages)
			headers := []string{"ID", "Name", "Language", "Slug", "Parent ID"}
			rows := make([][]string, len(paginated.Items))
			for i, cat := range paginated.Items {
				pid := ""
				if cat.ParentId != nil {
					pid = *cat.ParentId
				}
				rows[i] = []string{cat.Id, cat.Name, cat.Language, cat.Slug, pid}
			}
			output.Table(headers, rows)
		})
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a category",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		description, _ := cmd.Flags().GetString("description")
		parentID, _ := cmd.Flags().GetString("parent-id")
		language, _ := cmd.Flags().GetString("language")

		var err error
		if name, err = prompt.RequireArg("name", name); err != nil {
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

		body := client.CategoryCreateModel{
			Name:        name,
			Slug:        slug,
			Language:    language,
			Description: cmdutil.StringPtr(description),
			ParentId:    cmdutil.StringPtr(parentID),
		}

		resp, err := c.CreateCategoryManageCategoriesPostWithResponse(cmdutil.Ctx(), &client.CreateCategoryManageCategoriesPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printCategory(resp.JSON201)
			})
		}
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a category by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetCategoryManageCategoriesCategoryIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetCategoryManageCategoriesCategoryIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printCategory(resp.JSON200)
			})
		}
		return nil
	},
}

var getBySlugCmd = &cobra.Command{
	Use:   "get-by-slug <slug>",
	Short: "Get a category by slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetCategoryBySlugManageCategoriesBySlugSlugGetWithResponse(cmdutil.Ctx(), args[0], &client.GetCategoryBySlugManageCategoriesBySlugSlugGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printCategory(resp.JSON200)
			})
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.CategoryUpdateModel{}
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			body.Name = &v
		}
		if v, _ := cmd.Flags().GetString("slug"); v != "" {
			body.Slug = &v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			body.Description = &v
		}
		if v, _ := cmd.Flags().GetString("parent-id"); v != "" {
			body.ParentId = &v
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			body.Language = &v
		}

		resp, err := c.UpdateCategoryManageCategoriesCategoryIdPutWithResponse(cmdutil.Ctx(), args[0], &client.UpdateCategoryManageCategoriesCategoryIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printCategory(resp.JSON200)
			})
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteCategoryManageCategoriesCategoryIdDeleteWithResponse(cmdutil.Ctx(), args[0], &client.DeleteCategoryManageCategoriesCategoryIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Category deleted.")
		return nil
	},
}

var subcategoriesCmd = &cobra.Command{
	Use:   "subcategories <id>",
	Short: "List subcategories",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetSubcategoriesManageCategoriesCategoryIdSubcategoriesGetWithResponse(cmdutil.Ctx(), args[0], &client.GetSubcategoriesManageCategoriesCategoryIdSubcategoriesGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"ID", "Name", "Language", "Slug"}
				rows := make([][]string, len(items))
				for i, cat := range items {
					rows[i] = []string{cat.Id, cat.Name, cat.Language, cat.Slug}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var articlesCmd = &cobra.Command{
	Use:   "articles <category_id>",
	Short: "List articles in a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetArticlesForCategoryManageCategoriesCategoryIdArticlesGetWithResponse(cmdutil.Ctx(), args[0], &client.GetArticlesForCategoryManageCategoriesCategoryIdArticlesGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"Article ID"}
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

var addArticleCmd = &cobra.Command{
	Use:   "add-article <category_id> <article_id>",
	Short: "Add an article to a category",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.AddArticleToCategoryManageCategoriesCategoryIdArticlesArticleIdPostWithResponse(cmdutil.Ctx(), args[0], args[1], &client.AddArticleToCategoryManageCategoriesCategoryIdArticlesArticleIdPostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Article added to category.")
		return nil
	},
}

var removeArticleCmd = &cobra.Command{
	Use:   "remove-article <category_id> <article_id>",
	Short: "Remove an article from a category",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.RemoveArticleFromCategoryManageCategoriesCategoryIdArticlesArticleIdDeleteWithResponse(cmdutil.Ctx(), args[0], args[1], &client.RemoveArticleFromCategoryManageCategoriesCategoryIdArticlesArticleIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Article removed from category.")
		return nil
	},
}

var translationsCmd = &cobra.Command{
	Use:   "translations <id>",
	Short: "List translations of a category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetTranslationsManageCategoriesCategoryIdTranslationsGetWithResponse(cmdutil.Ctx(), args[0], &client.GetTranslationsManageCategoriesCategoryIdTranslationsGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				items := *resp.JSON200
				headers := []string{"ID", "Name", "Language", "Slug"}
				rows := make([][]string, len(items))
				for i, cat := range items {
					rows[i] = []string{cat.Id, cat.Name, cat.Language, cat.Slug}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var createTranslationCmd = &cobra.Command{
	Use:   "create-translation <source_id>",
	Short: "Create a category translation",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		description, _ := cmd.Flags().GetString("description")
		parentID, _ := cmd.Flags().GetString("parent-id")
		language, _ := cmd.Flags().GetString("language")

		var err error
		if name, err = prompt.RequireArg("name", name); err != nil {
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

		body := client.CategoryCreateModel{
			Name:        name,
			Slug:        slug,
			Language:    language,
			Description: cmdutil.StringPtr(description),
			ParentId:    cmdutil.StringPtr(parentID),
		}

		resp, err := c.CreateTranslationManageCategoriesSourceIdTranslationsPostWithResponse(cmdutil.Ctx(), args[0], &client.CreateTranslationManageCategoriesSourceIdTranslationsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printCategory(resp.JSON201)
			})
		}
		return nil
	},
}

func printCategory(cat *client.CategoryResponseModel) {
	desc := ""
	if cat.Description != nil {
		desc = *cat.Description
	}
	parentID := ""
	if cat.ParentId != nil {
		parentID = *cat.ParentId
	}
	pairs := [][2]string{
		{"ID", cat.Id},
		{"Name", cat.Name},
		{"Slug", cat.Slug},
		{"Language", cat.Language},
		{"Description", desc},
		{"Parent ID", parentID},
		{"Created", cat.CreatedAt},
		{"Updated", cat.UpdatedAt},
	}
	output.KeyValue(pairs)
}

func init() {
	listCmd.Flags().String("parent-id", "", "Filter by parent category")
	cmdutil.AddListFlags(listCmd)

	for _, c := range []*cobra.Command{createCmd, createTranslationCmd} {
		c.Flags().String("name", "", "Category name")
		c.Flags().String("slug", "", "URL-friendly identifier")
		c.Flags().String("description", "", "Category description")
		c.Flags().String("parent-id", "", "Parent category ID")
		c.Flags().String("language", "", "Language code")
	}

	updateCmd.Flags().String("name", "", "Category name")
	updateCmd.Flags().String("slug", "", "URL-friendly identifier")
	updateCmd.Flags().String("description", "", "Category description")
	updateCmd.Flags().String("parent-id", "", "Parent category ID")
	updateCmd.Flags().String("language", "", "Language code")

}
