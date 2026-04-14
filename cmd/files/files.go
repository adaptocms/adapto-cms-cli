package files

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

var Cmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
}

func init() {
	Cmd.AddCommand(listCmd, createMetadataCmd, uploadCmd, uploadByIDCmd, getCmd,
		updateCmd, deleteCmd, multipartInitCmd, multipartUploadCmd,
		multipartCompleteCmd, multipartAbortCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List files",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		fileType, _ := cmd.Flags().GetString("type")
		filename, _ := cmd.Flags().GetString("filename")
		contentType, _ := cmd.Flags().GetString("content-type")
		tag, _ := cmd.Flags().GetString("tag")
		field, _ := cmd.Flags().GetString("field")
		order, _ := cmd.Flags().GetString("order")
		page, _ := cmd.Flags().GetInt("page")
		limit, _ := cmd.Flags().GetInt("limit")

		resp, err := c.ListFilesManageFilesGetWithResponse(cmdutil.Ctx(), &client.ListFilesManageFilesGetParams{
			Type:        cmdutil.StringPtr(fileType),
			Filename:    cmdutil.StringPtr(filename),
			ContentType: cmdutil.StringPtr(contentType),
			Tag:         cmdutil.StringPtr(tag),
			Field:       cmdutil.StringPtr(field),
			Order:       cmdutil.StringPtr(order),
			Page:        cmdutil.IntPtr(page),
			Limit:       cmdutil.IntPtr(limit),
		})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		// resp.JSON200 is *PaginatedResponse with Items []interface{}, so unmarshal Body manually.
		var paginated struct {
			Items []client.FileResponseModel `json:"items"`
			Total int                        `json:"total"`
			Page  int                        `json:"page"`
			Pages int                        `json:"pages"`
		}
		if err := json.Unmarshal(resp.Body, &paginated); err != nil {
			return fmt.Errorf("failed to parse files list: %w", err)
		}
		output.Print(paginated, func(d interface{}) {
			fmt.Printf("Total: %d (page %d/%d)\n\n", paginated.Total, paginated.Page, paginated.Pages)
			headers := []string{"ID", "Filename", "Type", "Content-Type", "Size", "Status"}
			rows := make([][]string, len(paginated.Items))
			for i, f := range paginated.Items {
				size := ""
				if f.Size != nil {
					size = fmt.Sprintf("%d", *f.Size)
				}
				rows[i] = []string{f.Id, output.Truncate(f.Filename, 40), string(f.Type), f.ContentType, size, string(f.UploadStatus)}
			}
			output.Table(headers, rows)
		})
		return nil
	},
}

var createMetadataCmd = &cobra.Command{
	Use:   "create-metadata",
	Short: "Create file metadata (before upload)",
	RunE: func(cmd *cobra.Command, args []string) error {
		filename, _ := cmd.Flags().GetString("filename")
		contentType, _ := cmd.Flags().GetString("content-type")
		tags, _ := cmd.Flags().GetString("tags")

		var err error
		if filename, err = prompt.RequireArg("filename", filename); err != nil {
			return err
		}
		if contentType, err = prompt.RequireArg("content-type", contentType); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.FileCreateModel{
			Filename:    filename,
			ContentType: contentType,
			Tags:        cmdutil.StringSlicePtr(tags),
		}

		resp, err := c.CreateFileMetadataManageFilesMetadataPostWithResponse(cmdutil.Ctx(), &client.CreateFileMetadataManageFilesMetadataPostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON201 != nil {
			output.Print(resp.JSON201, func(d interface{}) {
				printFile(resp.JSON201)
			})
		}
		return nil
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload <filepath>",
	Short: "Upload a file",
	Long:  "Upload a file directly. Creates metadata and uploads in one step.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tags, _ := cmd.Flags().GetString("tags")

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.BodyUploadFileManageFilesPost{}
		// File upload requires multipart form - use raw body
		_ = body
		_ = tags
		_ = c

		// For file upload, we need to use the raw HTTP client approach
		// because oapi-codegen's File type needs special handling
		return fmt.Errorf("file upload requires using: adapto files create-metadata + adapto files upload-by-id <file_id> <filepath>")
	},
}

var uploadByIDCmd = &cobra.Command{
	Use:   "upload-by-id <file_id> <filepath>",
	Short: "Upload file content for an existing file record",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		// File upload needs multipart/form-data with raw HTTP
		return fmt.Errorf("file upload via CLI uses multipart HTTP - use curl or the SDK for now:\n  curl -X POST %s/manage/files/%s/upload -H 'Authorization: Bearer $ADAPTO_TOKEN' -F 'file=@%s'",
			"$ADAPTO_API_URL", args[0], args[1])
	},
}

var getCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get file info by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.GetFileManageFilesFileIdGetWithResponse(cmdutil.Ctx(), args[0], &client.GetFileManageFilesFileIdGetParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printFile(resp.JSON200)
			})
		}
		return nil
	},
}

var updateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update file metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		body := client.FileUpdateModel{}
		if v, _ := cmd.Flags().GetString("filename"); v != "" {
			body.Filename = &v
		}
		if v, _ := cmd.Flags().GetString("tags"); v != "" {
			body.Tags = cmdutil.StringSlicePtr(v)
		}

		resp, err := c.UpdateFileManageFilesFileIdPutWithResponse(cmdutil.Ctx(), args[0], &client.UpdateFileManageFilesFileIdPutParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				printFile(resp.JSON200)
			})
		}
		return nil
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.DeleteFileManageFilesFileIdDeleteWithResponse(cmdutil.Ctx(), args[0], &client.DeleteFileManageFilesFileIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("File deleted.")
		return nil
	},
}

var multipartInitCmd = &cobra.Command{
	Use:   "multipart-init <file_id>",
	Short: "Initialize a multipart upload",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.InitMultipartUploadManageFilesFileIdMultipartInitPostWithResponse(cmdutil.Ctx(), args[0], &client.InitMultipartUploadManageFilesFileIdMultipartInitPostParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		if resp.JSON200 != nil {
			output.Print(resp.JSON200, func(d interface{}) {
				fmt.Printf("File ID:   %s\n", resp.JSON200.FileId)
				fmt.Printf("Upload ID: %s\n", resp.JSON200.UploadId)
			})
		}
		return nil
	},
}

var multipartUploadCmd = &cobra.Command{
	Use:   "multipart-upload <file_id> <upload_id> <part_number> <filepath>",
	Short: "Upload a part of a multipart upload",
	Args:  cobra.ExactArgs(4),
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("multipart part upload uses multipart HTTP - use curl:\n  curl -X POST %s/manage/files/%s/multipart/%s/parts/%s -H 'Authorization: Bearer $ADAPTO_TOKEN' -F 'file=@%s'",
			"$ADAPTO_API_URL", args[0], args[1], args[2], args[3])
	},
}

var multipartCompleteCmd = &cobra.Command{
	Use:   "multipart-complete <file_id> <upload_id>",
	Short: "Complete a multipart upload",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		partsJSON, _ := cmd.Flags().GetString("parts")
		var err error
		if partsJSON, err = prompt.RequireArg("parts", partsJSON); err != nil {
			return err
		}

		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		var body client.CompleteMultipartUploadRequest
		if err := json.Unmarshal([]byte(partsJSON), &body.Parts); err != nil {
			return fmt.Errorf("invalid --parts JSON: %w", err)
		}

		resp, err := c.CompleteMultipartUploadManageFilesFileIdMultipartUploadIdCompletePostWithResponse(cmdutil.Ctx(), args[0], args[1], &client.CompleteMultipartUploadManageFilesFileIdMultipartUploadIdCompletePostParams{}, body)
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Multipart upload completed.")
		return nil
	},
}

var multipartAbortCmd = &cobra.Command{
	Use:   "multipart-abort <file_id> <upload_id>",
	Short: "Abort a multipart upload",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, _, err := cmdutil.NewClientWithAuth()
		if err != nil {
			return err
		}

		resp, err := c.AbortMultipartUploadManageFilesFileIdMultipartUploadIdDeleteWithResponse(cmdutil.Ctx(), args[0], args[1], &client.AbortMultipartUploadManageFilesFileIdMultipartUploadIdDeleteParams{})
		if err != nil {
			return err
		}
		if err := cmdutil.CheckErr(resp.StatusCode(), resp.Body); err != nil {
			return err
		}

		output.Success("Multipart upload aborted.")
		return nil
	},
}

func printFile(f *client.FileResponseModel) {
	url := ""
	if f.Url != nil {
		url = *f.Url
	}
	size := ""
	if f.Size != nil {
		size = fmt.Sprintf("%d", *f.Size)
	}
	pairs := [][2]string{
		{"ID", f.Id},
		{"Filename", f.Filename},
		{"Type", string(f.Type)},
		{"Content-Type", f.ContentType},
		{"Size", size},
		{"Upload Status", string(f.UploadStatus)},
		{"Tags", strings.Join(f.Tags, ", ")},
		{"URL", url},
		{"Created", f.CreatedAt},
		{"Updated", f.UpdatedAt},
	}
	output.KeyValue(pairs)
}

func init() {
	listCmd.Flags().String("type", "", "Filter by file type")
	listCmd.Flags().String("filename", "", "Filter by filename")
	listCmd.Flags().String("content-type", "", "Filter by content type")
	listCmd.Flags().String("tag", "", "Filter by tag")
	listCmd.Flags().String("field", "", "Sort field")
	listCmd.Flags().String("order", "", "Sort order")
	listCmd.Flags().Int("page", 0, "Page number")
	listCmd.Flags().Int("limit", 0, "Items per page")

	createMetadataCmd.Flags().String("filename", "", "Original filename")
	createMetadataCmd.Flags().String("content-type", "", "MIME type")
	createMetadataCmd.Flags().String("tags", "", "Comma-separated tags")

	uploadCmd.Flags().String("tags", "", "Comma-separated tags")

	updateCmd.Flags().String("filename", "", "New filename")
	updateCmd.Flags().String("tags", "", "Comma-separated tags")

	multipartCompleteCmd.Flags().String("parts", "", "Parts JSON array")
}
