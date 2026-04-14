package collections

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
	Use:   "collections",
	Short: "Manage custom collections",
}

var itemsCmd = &cobra.Command{
	Use:   "items",
	Short: "Manage collection items",
}

func init() {
	Cmd.AddCommand(listCmd, createCmd, getCmd, getBySlugCmd, updateCmd, deleteCmd, itemsCmd)
	itemsCmd.AddCommand(itemsListCmd, itemsCreateCmd, itemsCreateBatchCmd,
		itemsGetCmd, itemsGetBySlugCmd, itemsUpdateCmd, itemsDeleteCmd,
		itemsPublishCmd, itemsArchiveCmd, itemsTranslationsCmd, itemsCreateTranslationCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List collections",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		keyword, _ := cmd.Flags().GetString("keyword")
		language, _ := cmd.Flags().GetString("language")
		field, _ := cmd.Flags().GetString("field")
		order, _ := cmd.Flags().GetString("order")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		resp, err := c.ListCollectionsManageCustomCollectionsGetWithResponse(cmdutil.Ctx(), &client.ListCollectionsManageCustomCollectionsGetParams{
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

		// resp.JSON200 is *PaginatedResponse with Items []interface{}, so unmarshal Body manually.
		var paginated struct {
			Items []client.CustomCollectionResponseModel `json:"items"`
			Total int                                    `json:"total"`
			Page  int                                    `json:"page"`
			Pages int                                    `json:"pages"`
		}
		if err := json.Unmarshal(resp.Body, &paginated); err != nil {
			return fmt.Errorf("failed to parse collections list: %w", err)
		}
		output.Print(paginated, func(d interface{}) {
			fmt.Printf("Total: %d (page %d/%d)\n\n", paginated.Total, paginated.Page, paginated.Pages)
			headers := []string{"ID", "Name", "Status", "Language", "Slug"}
			rows := make([][]string, len(paginated.Items))
			for i, col := range paginated.Items {
				rows[i] = []string{col.Id, col.Name, string(col.Status), col.Language, col.Slug}
			}
			output.Table(headers, rows)
		})
		return nil
	},
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a collection",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		description, _ := cmd.Flags().GetString("description")
		language, _ := cmd.Flags().GetString("language")
		fieldsJSON, _ := cmd.Flags().GetString("fields-json")
		status, _ := cmd.Flags().GetString("status")

		var err error
		if name, err = prompt.RequireArg("name", name); err != nil {
			return err
		}
		if slug, err = prompt.RequireArg("slug", slug); err != nil {
			return err
		}
		if description, err = prompt.RequireArg("description", description); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.CustomCollectionCreateModel{
			Name:        name,
			Slug:        slug,
			Description: description,
			Language:    language,
		}
		if fieldsJSON != "" {
			if err := json.Unmarshal([]byte(fieldsJSON), &body.Fields); err != nil {
				return fmt.Errorf("invalid --fields-json: %w", err)
			}
		}
		if status != "" {
			s := client.CustomCollectionStatus(status)
			body.Status = &s
		}

		resp, err := c.CreateCollectionManageCustomCollectionsPostWithResponse(cmdutil.Ctx(), &client.CreateCollectionManageCustomCollectionsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printCollection(resp.JSON201)
			})
		}
		return nil
	},
}

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get a collection by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetCollectionManageCustomCollectionsCollectionIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetCollectionManageCustomCollectionsCollectionIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printCollection(resp.JSON200)
			})
		}
		return nil
	},
}

var getBySlugCmd = &cobra.Command{
	Use:   "get-by-slug <slug>",
	Short: "Get a collection by slug",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetCollectionBySlugManageCustomCollectionsBySlugSlugGetWithResponse(cmdutil.Ctx(), args[0], &client.GetCollectionBySlugManageCustomCollectionsBySlugSlugGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printCollection(resp.JSON200)
			})
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.CustomCollectionUpdateModel{}
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			body.Name = &v
		}
		if v, _ := cmd.Flags().GetString("slug"); v != "" {
			body.Slug = &v
		}
		if v, _ := cmd.Flags().GetString("description"); v != "" {
			body.Description = &v
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			body.Language = &v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			s := client.CustomCollectionStatus(v)
			body.Status = &s
		}
		if v, _ := cmd.Flags().GetString("fields-json"); v != "" {
			var fields []client.FieldDefinitionModel
			if err := json.Unmarshal([]byte(v), &fields); err != nil {
				return fmt.Errorf("invalid --fields-json: %w", err)
			}
			body.Fields = &fields
		}

		resp, err := c.UpdateCollectionManageCustomCollectionsCollectionIdPutWithResponse(cmdutil.Ctx(), args[0], &client.UpdateCollectionManageCustomCollectionsCollectionIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printCollection(resp.JSON200)
			})
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteCollectionManageCustomCollectionsCollectionIdDeleteWithResponse(cmdutil.Ctx(), args[0], &client.DeleteCollectionManageCustomCollectionsCollectionIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Collection deleted.")
		return nil
	},
}

// --- Items subcommands ---

var itemsListCmd = &cobra.Command{
	Use:   "list <collection_id>",
	Short: "List items in a collection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		status, _ := cmd.Flags().GetString("status")
		keyword, _ := cmd.Flags().GetString("keyword")
		language, _ := cmd.Flags().GetString("language")
		field, _ := cmd.Flags().GetString("field")
		order, _ := cmd.Flags().GetString("order")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		resp, err := c.ListItemsManageCustomCollectionsCollectionIdItemsGetWithResponse(cmdutil.Ctx(), args[0], &client.ListItemsManageCustomCollectionsCollectionIdItemsGetParams{
			Status:   cmdutil.StringPtr(status),
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

		// resp.JSON200 is *PaginatedResponse with Items []interface{}, so unmarshal Body manually.
		var paginated struct {
			Items []client.CustomCollectionItemResponseModel `json:"items"`
			Total int                                        `json:"total"`
			Page  int                                        `json:"page"`
			Pages int                                        `json:"pages"`
		}
		if err := json.Unmarshal(resp.Body, &paginated); err != nil {
			return fmt.Errorf("failed to parse items list: %w", err)
		}
		output.Print(paginated, func(d interface{}) {
			fmt.Printf("Total: %d (page %d/%d)\n\n", paginated.Total, paginated.Page, paginated.Pages)
			headers := []string{"ID", "Title", "Status", "Language", "Slug"}
			rows := make([][]string, len(paginated.Items))
			for i, item := range paginated.Items {
				rows[i] = []string{item.Id, output.Truncate(item.Title, 40), string(item.Status), item.Language, output.Truncate(item.Slug, 30)}
			}
			output.Table(headers, rows)
		})
		return nil
	},
}

var itemsCreateCmd = &cobra.Command{
	Use:   "create <collection_id>",
	Short: "Create a collection item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		slug, _ := cmd.Flags().GetString("slug")
		dataJSON, _ := cmd.Flags().GetString("data-json")
		language, _ := cmd.Flags().GetString("language")
		status, _ := cmd.Flags().GetString("status")

		var err error
		if title, err = prompt.RequireArg("title", title); err != nil {
			return err
		}
		if slug, err = prompt.RequireArg("slug", slug); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}
		if dataJSON, err = prompt.RequireArg("data-json", dataJSON); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.CustomCollectionItemCreateModel{
			Title:    title,
			Slug:     slug,
			Language: language,
		}
		if err := json.Unmarshal([]byte(dataJSON), &body.Data); err != nil {
			return fmt.Errorf("invalid --data-json: %w", err)
		}
		if status != "" {
			s := client.CustomCollectionItemStatus(status)
			body.Status = &s
		}

		resp, err := c.CreateItemManageCustomCollectionsCollectionIdItemsPostWithResponse(cmdutil.Ctx(), args[0], &client.CreateItemManageCustomCollectionsCollectionIdItemsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printItem(resp.JSON201)
			})
		}
		return nil
	},
}

var itemsCreateBatchCmd = &cobra.Command{
	Use:   "create-batch <collection_id>",
	Short: "Create multiple items in batch",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		itemsJSON, _ := cmd.Flags().GetString("items-json")
		var err error
		if itemsJSON, err = prompt.RequireArg("items-json", itemsJSON); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		var body client.CustomCollectionBatchItemCreateModel
		if err := json.Unmarshal([]byte(itemsJSON), &body); err != nil {
			return fmt.Errorf("invalid --items-json: %w", err)
		}

		resp, err := c.CreateItemsBatchManageCustomCollectionsCollectionIdItemsBatchPostWithResponse(cmdutil.Ctx(), args[0], &client.CreateItemsBatchManageCustomCollectionsCollectionIdItemsBatchPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Successf("Batch created successfully.")
		return nil
	},
}

var itemsGetCmd = &cobra.Command{
	Use:   "get <collection_id> <item_id>",
	Short: "Get a collection item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetItemManageCustomCollectionsCollectionIdItemsItemIdGetWithResponse(cmdutil.Ctx(), args[0], args[1], &client.GetItemManageCustomCollectionsCollectionIdItemsItemIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printItem(resp.JSON200)
			})
		}
		return nil
	},
}

var itemsGetBySlugCmd = &cobra.Command{
	Use:   "get-by-slug <collection_id> <slug>",
	Short: "Get a collection item by slug",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetItemBySlugManageCustomCollectionsCollectionIdItemsBySlugSlugGetWithResponse(cmdutil.Ctx(), args[0], args[1], &client.GetItemBySlugManageCustomCollectionsCollectionIdItemsBySlugSlugGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printItem(resp.JSON200)
			})
		}
		return nil
	},
}

var itemsUpdateCmd = &cobra.Command{
	Use:   "update <collection_id> <item_id>",
	Short: "Update a collection item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.CustomCollectionItemUpdateModel{}
		if v, _ := cmd.Flags().GetString("title"); v != "" {
			body.Title = &v
		}
		if v, _ := cmd.Flags().GetString("slug"); v != "" {
			body.Slug = &v
		}
		if v, _ := cmd.Flags().GetString("language"); v != "" {
			body.Language = &v
		}
		if v, _ := cmd.Flags().GetString("status"); v != "" {
			s := client.CustomCollectionItemStatus(v)
			body.Status = &s
		}
		if v, _ := cmd.Flags().GetString("data-json"); v != "" {
			var data map[string]interface{}
			if err := json.Unmarshal([]byte(v), &data); err != nil {
				return fmt.Errorf("invalid --data-json: %w", err)
			}
			body.Data = &data
		}

		resp, err := c.UpdateItemManageCustomCollectionsCollectionIdItemsItemIdPutWithResponse(cmdutil.Ctx(), args[0], args[1], &client.UpdateItemManageCustomCollectionsCollectionIdItemsItemIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printItem(resp.JSON200)
			})
		}
		return nil
	},
}

var itemsDeleteCmd = &cobra.Command{
	Use:   "delete <collection_id> <item_id>",
	Short: "Delete a collection item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteItemManageCustomCollectionsCollectionIdItemsItemIdDeleteWithResponse(cmdutil.Ctx(), args[0], args[1], &client.DeleteItemManageCustomCollectionsCollectionIdItemsItemIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Item deleted.")
		return nil
	},
}

var itemsPublishCmd = &cobra.Command{
	Use:   "publish <collection_id> <item_id>",
	Short: "Publish a collection item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.PublishItemManageCustomCollectionsCollectionIdItemsItemIdPublishPostWithResponse(cmdutil.Ctx(), args[0], args[1], &client.PublishItemManageCustomCollectionsCollectionIdItemsItemIdPublishPostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Item published.")
		return nil
	},
}

var itemsArchiveCmd = &cobra.Command{
	Use:   "archive <collection_id> <item_id>",
	Short: "Archive a collection item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.ArchiveItemManageCustomCollectionsCollectionIdItemsItemIdArchivePostWithResponse(cmdutil.Ctx(), args[0], args[1], &client.ArchiveItemManageCustomCollectionsCollectionIdItemsItemIdArchivePostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Item archived.")
		return nil
	},
}

var itemsTranslationsCmd = &cobra.Command{
	Use:   "translations <collection_id> <item_id>",
	Short: "List translations of an item",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetItemTranslationsManageCustomCollectionsCollectionIdItemsItemIdTranslationsGetWithResponse(cmdutil.Ctx(), args[0], args[1], &client.GetItemTranslationsManageCustomCollectionsCollectionIdItemsItemIdTranslationsGetParams{})
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
				for i, item := range items {
					rows[i] = []string{item.Id, output.Truncate(item.Title, 40), item.Language, string(item.Status), item.Slug}
				}
				output.Table(headers, rows)
			})
		}
		return nil
	},
}

var itemsCreateTranslationCmd = &cobra.Command{
	Use:   "create-translation <collection_id> <source_id>",
	Short: "Create an item translation",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		slug, _ := cmd.Flags().GetString("slug")
		dataJSON, _ := cmd.Flags().GetString("data-json")
		language, _ := cmd.Flags().GetString("language")
		status, _ := cmd.Flags().GetString("status")

		var err error
		if title, err = prompt.RequireArg("title", title); err != nil {
			return err
		}
		if slug, err = prompt.RequireArg("slug", slug); err != nil {
			return err
		}
		if language, err = prompt.RequireArg("language", language); err != nil {
			return err
		}
		if dataJSON, err = prompt.RequireArg("data-json", dataJSON); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.CustomCollectionItemCreateModel{
			Title:    title,
			Slug:     slug,
			Language: language,
		}
		if err := json.Unmarshal([]byte(dataJSON), &body.Data); err != nil {
			return fmt.Errorf("invalid --data-json: %w", err)
		}
		if status != "" {
			s := client.CustomCollectionItemStatus(status)
			body.Status = &s
		}

		resp, err := c.CreateItemTranslationManageCustomCollectionsCollectionIdItemsSourceIdTranslationsPostWithResponse(cmdutil.Ctx(), args[0], args[1], &client.CreateItemTranslationManageCustomCollectionsCollectionIdItemsSourceIdTranslationsPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printItem(resp.JSON201)
			})
		}
		return nil
	},
}

func printCollection(col *client.CustomCollectionResponseModel) {
	pairs := [][2]string{
		{"ID", col.Id},
		{"Name", col.Name},
		{"Slug", col.Slug},
		{"Status", string(col.Status)},
		{"Language", col.Language},
		{"Description", col.Description},
		{"Created", col.CreatedAt},
		{"Updated", col.UpdatedAt},
	}
	output.KeyValue(pairs)
}

func printItem(item *client.CustomCollectionItemResponseModel) {
	dataBytes, _ := json.Marshal(item.Data)
	pairs := [][2]string{
		{"ID", item.Id},
		{"Title", item.Title},
		{"Slug", item.Slug},
		{"Status", string(item.Status)},
		{"Language", item.Language},
		{"Collection", item.CollectionId},
		{"Data", string(dataBytes)},
		{"Created", item.CreatedAt},
		{"Updated", item.UpdatedAt},
	}
	if item.PublishedAt != nil {
		pairs = append(pairs, [2]string{"Published", *item.PublishedAt})
	}
	output.KeyValue(pairs)
}

func init() {
	cmdutil.AddListFlags(listCmd)

	createCmd.Flags().String("name", "", "Collection name")
	createCmd.Flags().String("slug", "", "URL-friendly identifier")
	createCmd.Flags().String("description", "", "Collection description")
	createCmd.Flags().String("language", "", "Language code")
	createCmd.Flags().String("fields-json", "", "Field definitions JSON")
	createCmd.Flags().String("status", "", "Status")

	updateCmd.Flags().String("name", "", "Collection name")
	updateCmd.Flags().String("slug", "", "URL-friendly identifier")
	updateCmd.Flags().String("description", "", "Collection description")
	updateCmd.Flags().String("language", "", "Language code")
	updateCmd.Flags().String("fields-json", "", "Field definitions JSON")
	updateCmd.Flags().String("status", "", "Status")

	// Items list flags
	itemsListCmd.Flags().String("status", "", "Filter by status")
	cmdutil.AddListFlags(itemsListCmd)

	// Items create/create-translation flags
	for _, c := range []*cobra.Command{itemsCreateCmd, itemsCreateTranslationCmd} {
		c.Flags().String("title", "", "Item title")
		c.Flags().String("slug", "", "URL-friendly identifier")
		c.Flags().String("data-json", "", "Item data JSON")
		c.Flags().String("language", "", "Language code")
		c.Flags().String("status", "", "Status")
	}

	// Items update flags
	itemsUpdateCmd.Flags().String("title", "", "Item title")
	itemsUpdateCmd.Flags().String("slug", "", "URL-friendly identifier")
	itemsUpdateCmd.Flags().String("data-json", "", "Item data JSON")
	itemsUpdateCmd.Flags().String("language", "", "Language code")
	itemsUpdateCmd.Flags().String("status", "", "Status")

	// Items batch create
	itemsCreateBatchCmd.Flags().String("items-json", "", "Batch items JSON")
}
