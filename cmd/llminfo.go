package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(llmInfoCmd)
}

var llmInfoCmd = &cobra.Command{
	Use:   "llm-info",
	Short: "Print full CLI reference for LLM consumption",
	Long:  "Output a comprehensive markdown description of every command, flag, and workflow so an LLM agent can understand and use the Adapto CMS CLI.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(llmInfoText)
	},
}

const llmInfoText = `# Adapto CMS CLI — Complete Reference

## Overview

Adapto CMS CLI (` + "`adapto`" + `) is a command-line interface for the Adapto CMS management API. It manages articles, pages, categories, custom collections, files, and micro copy entries across a multi-tenant, multi-language headless CMS.

## Authentication

### Credential storage
After ` + "`adapto auth login`" + `, tokens are saved to ` + "`~/.adapto/credentials.json`" + `. Subsequent commands read from this file automatically.

### Environment variables
| Variable | Purpose |
|----------|---------|
| ` + "`ADAPTO_API_URL`" + ` | API base URL (default: ` + "`https://api.adaptocms.com`" + `) |
| ` + "`ADAPTO_TOKEN`" + ` | Bearer token (overrides stored credential) |
| ` + "`ADAPTO_TENANT_ID`" + ` | Tenant ID (overrides stored credential) |

### Multi-tenancy
A user can belong to multiple organizations, each with one or more tenants. After login, the CLI prompts you to select a tenant (or auto-selects if only one exists). Use ` + "`adapto auth switch-tenant`" + ` to change the active tenant.

## Global Flags

Every command accepts these flags:

| Flag | Type | Description |
|------|------|-------------|
| ` + "`--api-url`" + ` | string | Adapto API base URL (env: ADAPTO_API_URL) |
| ` + "`--token`" + ` | string | Bearer token (env: ADAPTO_TOKEN) |
| ` + "`--tenant-id`" + ` | string | Tenant ID (env: ADAPTO_TENANT_ID) |
| ` + "`--json`" + ` | bool | Output as JSON instead of table |
| ` + "`--verbose`" + ` | bool | Show HTTP request/response details |

---

## Commands

### adapto version

Print the CLI version.

` + "```" + `
adapto version
` + "```" + `

---

### adapto auth

Authentication commands.

#### adapto auth login
Login with email and password. Saves tokens to credentials file and prompts for tenant selection.

| Flag | Description |
|------|-------------|
| ` + "`--email`" + ` | Email address |
| ` + "`--password`" + ` | Password |

` + "```" + `
adapto auth login --email user@example.com --password secret
` + "```" + `

#### adapto auth register
Register a new account.

| Flag | Description |
|------|-------------|
| ` + "`--email`" + ` | Email address (required) |
| ` + "`--password`" + ` | Password (required) |
| ` + "`--first-name`" + ` | First name |
| ` + "`--last-name`" + ` | Last name |

#### adapto auth logout
Logout and revoke refresh token. Clears stored credentials.

| Flag | Description |
|------|-------------|
| ` + "`--refresh-token`" + ` | Refresh token to revoke (defaults to stored token) |

#### adapto auth refresh
Refresh the access token. Updates stored credentials.

| Flag | Description |
|------|-------------|
| ` + "`--refresh-token`" + ` | Refresh token (defaults to stored token) |

#### adapto auth me
Get current user info (ID, email, status, name).

#### adapto auth orgs
List your organizations and their tenants, showing which tenant is active.

#### adapto auth switch-tenant
Switch the active tenant/organization.

| Flag | Description |
|------|-------------|
| ` + "`--tenant-id`" + ` | Tenant/organization ID to switch to (interactive picker if omitted) |

#### adapto auth change-password
Change your password.

| Flag | Description |
|------|-------------|
| ` + "`--current-password`" + ` | Current password |
| ` + "`--new-password`" + ` | New password |

#### adapto auth request-password-reset
Request a password reset email.

| Flag | Description |
|------|-------------|
| ` + "`--email`" + ` | Email address |

#### adapto auth reset-password
Reset password with a token received via email.

| Flag | Description |
|------|-------------|
| ` + "`--token`" + ` | Password reset token |
| ` + "`--new-password`" + ` | New password |

#### adapto auth activate
Activate account with a token received via email.

| Flag | Description |
|------|-------------|
| ` + "`--token`" + ` | Activation token |

#### adapto auth resend-activation
Resend activation email.

| Flag | Description |
|------|-------------|
| ` + "`--email`" + ` | Email address |

#### adapto auth login-github
Login via GitHub OAuth. Returns the OAuth URL to visit.

| Flag | Description |
|------|-------------|
| ` + "`--redirect-uri`" + ` | OAuth redirect URI |

#### adapto auth callback-github
Complete GitHub OAuth callback.

| Flag | Description |
|------|-------------|
| ` + "`--code`" + ` | OAuth authorization code |
| ` + "`--redirect-uri`" + ` | OAuth redirect URI |

#### adapto auth login-google
Login via Google credential.

| Flag | Description |
|------|-------------|
| ` + "`--credential`" + ` | Google ID token |

---

### adapto articles

Manage articles. All subcommands require authentication.

#### adapto articles list
List articles with pagination and filters.

| Flag | Description |
|------|-------------|
| ` + "`--status`" + ` | Filter by status (draft/published/archived) |
| ` + "`--category`" + ` | Filter by category ID |
| ` + "`--tag`" + ` | Filter by tag |
| ` + "`--keyword`" + ` | Search keyword |
| ` + "`--language`" + ` | Filter by language code |
| ` + "`--field`" + ` | Sort field |
| ` + "`--order`" + ` | Sort order (asc/desc) |
| ` + "`--page`" + ` | Page number |
| ` + "`--limit`" + ` | Items per page |

` + "```" + `
adapto articles list --status published --language en --limit 10
` + "```" + `

#### adapto articles create
Create an article.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Article title (required) |
| ` + "`--content`" + ` | Article content (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--author`" + ` | Author name (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--summary`" + ` | Article summary |
| ` + "`--status`" + ` | Status (draft/published) |
| ` + "`--tags`" + ` | Comma-separated tags |
| ` + "`--source`" + ` | Source JSON (default: {"type":"internal","name":"CLI"}) |

` + "```" + `
adapto articles create --title "Hello World" --content "<p>Body</p>" --slug hello-world --author "Jane" --language en
` + "```" + `

#### adapto articles get <id>
Get an article by ID.

#### adapto articles get-by-slug <slug>
Get an article by slug.

#### adapto articles update <id>
Update an article. Only provided flags are changed.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Article title |
| ` + "`--content`" + ` | Article content |
| ` + "`--slug`" + ` | URL-friendly identifier |
| ` + "`--author`" + ` | Author name |
| ` + "`--summary`" + ` | Article summary |
| ` + "`--language`" + ` | Language code |
| ` + "`--status`" + ` | Status |
| ` + "`--tags`" + ` | Comma-separated tags |
| ` + "`--source`" + ` | Source JSON |

#### adapto articles delete <id>
Delete an article.

#### adapto articles publish <id>
Publish an article (set status to published).

#### adapto articles archive <id>
Archive an article.

#### adapto articles translations <id>
List all translations of an article.

#### adapto articles create-translation <source_id>
Create a translation of an existing article.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Title (required) |
| ` + "`--content`" + ` | Content (required) |
| ` + "`--slug`" + ` | Slug (required) |
| ` + "`--author`" + ` | Author (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--summary`" + ` | Summary |
| ` + "`--tags`" + ` | Comma-separated tags |
| ` + "`--source`" + ` | Source JSON |

#### adapto articles categories <id>
List category IDs associated with an article.

---

### adapto categories

Manage categories. All subcommands require authentication.

#### adapto categories list
List categories with pagination and filters.

| Flag | Description |
|------|-------------|
| ` + "`--parent-id`" + ` | Filter by parent category |
| ` + "`--keyword`" + ` | Search keyword |
| ` + "`--language`" + ` | Filter by language code |
| ` + "`--field`" + ` | Sort field |
| ` + "`--order`" + ` | Sort order (asc/desc) |
| ` + "`--page`" + ` | Page number |
| ` + "`--limit`" + ` | Items per page |

#### adapto categories create
Create a category.

| Flag | Description |
|------|-------------|
| ` + "`--name`" + ` | Category name (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--description`" + ` | Category description |
| ` + "`--parent-id`" + ` | Parent category ID |

#### adapto categories get <id>
Get a category by ID.

#### adapto categories get-by-slug <slug>
Get a category by slug.

#### adapto categories update <id>
Update a category. Only provided flags are changed.

| Flag | Description |
|------|-------------|
| ` + "`--name`" + ` | Category name |
| ` + "`--slug`" + ` | URL-friendly identifier |
| ` + "`--description`" + ` | Category description |
| ` + "`--parent-id`" + ` | Parent category ID |
| ` + "`--language`" + ` | Language code |

#### adapto categories delete <id>
Delete a category.

#### adapto categories subcategories <id>
List subcategories of a category.

#### adapto categories articles <category_id>
List article IDs in a category.

#### adapto categories add-article <category_id> <article_id>
Add an article to a category.

#### adapto categories remove-article <category_id> <article_id>
Remove an article from a category.

#### adapto categories translations <id>
List translations of a category.

#### adapto categories create-translation <source_id>
Create a category translation.

| Flag | Description |
|------|-------------|
| ` + "`--name`" + ` | Category name (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--description`" + ` | Category description |
| ` + "`--parent-id`" + ` | Parent category ID |

---

### adapto collections

Manage custom collections. All subcommands require authentication.

#### adapto collections list
List collections with pagination and filters.

| Flag | Description |
|------|-------------|
| ` + "`--keyword`" + ` | Search keyword |
| ` + "`--language`" + ` | Filter by language code |
| ` + "`--field`" + ` | Sort field |
| ` + "`--order`" + ` | Sort order (asc/desc) |
| ` + "`--page`" + ` | Page number |
| ` + "`--limit`" + ` | Items per page |

#### adapto collections create
Create a collection.

| Flag | Description |
|------|-------------|
| ` + "`--name`" + ` | Collection name (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--description`" + ` | Collection description (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--fields-json`" + ` | Field definitions JSON |
| ` + "`--status`" + ` | Status |

#### adapto collections get <id>
Get a collection by ID.

#### adapto collections get-by-slug <slug>
Get a collection by slug.

#### adapto collections update <id>
Update a collection. Only provided flags are changed.

| Flag | Description |
|------|-------------|
| ` + "`--name`" + ` | Collection name |
| ` + "`--slug`" + ` | URL-friendly identifier |
| ` + "`--description`" + ` | Collection description |
| ` + "`--language`" + ` | Language code |
| ` + "`--fields-json`" + ` | Field definitions JSON |
| ` + "`--status`" + ` | Status |

#### adapto collections delete <id>
Delete a collection.

#### adapto collections items list <collection_id>
List items in a collection.

| Flag | Description |
|------|-------------|
| ` + "`--status`" + ` | Filter by status |
| ` + "`--keyword`" + ` | Search keyword |
| ` + "`--language`" + ` | Filter by language code |
| ` + "`--field`" + ` | Sort field |
| ` + "`--order`" + ` | Sort order (asc/desc) |
| ` + "`--page`" + ` | Page number |
| ` + "`--limit`" + ` | Items per page |

#### adapto collections items create <collection_id>
Create a collection item.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Item title (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--data-json`" + ` | Item data JSON (required) |
| ` + "`--status`" + ` | Status |

` + "```" + `
adapto collections items create abc123 --title "My Item" --slug my-item --language en --data-json '{"field1":"value1"}'
` + "```" + `

#### adapto collections items create-batch <collection_id>
Create multiple items in batch.

| Flag | Description |
|------|-------------|
| ` + "`--items-json`" + ` | Batch items JSON (required) |

#### adapto collections items get <collection_id> <item_id>
Get a collection item by ID.

#### adapto collections items get-by-slug <collection_id> <slug>
Get a collection item by slug.

#### adapto collections items update <collection_id> <item_id>
Update a collection item. Only provided flags are changed.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Item title |
| ` + "`--slug`" + ` | URL-friendly identifier |
| ` + "`--language`" + ` | Language code |
| ` + "`--data-json`" + ` | Item data JSON |
| ` + "`--status`" + ` | Status |

#### adapto collections items delete <collection_id> <item_id>
Delete a collection item.

#### adapto collections items publish <collection_id> <item_id>
Publish a collection item.

#### adapto collections items archive <collection_id> <item_id>
Archive a collection item.

#### adapto collections items translations <collection_id> <item_id>
List translations of a collection item.

#### adapto collections items create-translation <collection_id> <source_id>
Create a translation of a collection item.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Item title (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--data-json`" + ` | Item data JSON (required) |
| ` + "`--status`" + ` | Status |

---

### adapto pages

Manage pages. All subcommands require authentication.

#### adapto pages list
List pages with pagination and filters.

| Flag | Description |
|------|-------------|
| ` + "`--status`" + ` | Filter by status |
| ` + "`--tag`" + ` | Filter by tag |
| ` + "`--keyword`" + ` | Search keyword |
| ` + "`--language`" + ` | Filter by language code |
| ` + "`--field`" + ` | Sort field |
| ` + "`--order`" + ` | Sort order (asc/desc) |
| ` + "`--page`" + ` | Page number |
| ` + "`--limit`" + ` | Items per page |

#### adapto pages create
Create a page.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Page title (required) |
| ` + "`--content`" + ` | Page content (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--menu-label`" + ` | Menu label |
| ` + "`--parent-id`" + ` | Parent page ID |
| ` + "`--status`" + ` | Status |
| ` + "`--tags`" + ` | Comma-separated tags |

#### adapto pages get <id>
Get a page by ID.

#### adapto pages get-by-slug <slug>
Get a page by slug.

#### adapto pages update <id>
Update a page. Only provided flags are changed.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Page title |
| ` + "`--content`" + ` | Page content |
| ` + "`--slug`" + ` | URL-friendly identifier |
| ` + "`--menu-label`" + ` | Menu label |
| ` + "`--parent-id`" + ` | Parent page ID |
| ` + "`--language`" + ` | Language code |
| ` + "`--status`" + ` | Status |
| ` + "`--tags`" + ` | Comma-separated tags |

#### adapto pages delete <id>
Delete a page.

#### adapto pages publish <id>
Publish a page.

#### adapto pages archive <id>
Archive a page.

#### adapto pages translations <id>
List translations of a page.

#### adapto pages create-translation <source_id>
Create a page translation.

| Flag | Description |
|------|-------------|
| ` + "`--title`" + ` | Page title (required) |
| ` + "`--content`" + ` | Page content (required) |
| ` + "`--slug`" + ` | URL-friendly identifier (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--menu-label`" + ` | Menu label |
| ` + "`--parent-id`" + ` | Parent page ID |
| ` + "`--tags`" + ` | Comma-separated tags |

---

### adapto files

Manage files. All subcommands require authentication.

#### adapto files list
List files with pagination and filters.

| Flag | Description |
|------|-------------|
| ` + "`--type`" + ` | Filter by file type |
| ` + "`--filename`" + ` | Filter by filename |
| ` + "`--content-type`" + ` | Filter by MIME type |
| ` + "`--tag`" + ` | Filter by tag |
| ` + "`--field`" + ` | Sort field |
| ` + "`--order`" + ` | Sort order (asc/desc) |
| ` + "`--page`" + ` | Page number |
| ` + "`--limit`" + ` | Items per page |

#### adapto files create-metadata
Create file metadata (before uploading content).

| Flag | Description |
|------|-------------|
| ` + "`--filename`" + ` | Original filename (required) |
| ` + "`--content-type`" + ` | MIME type (required) |
| ` + "`--tags`" + ` | Comma-separated tags |

#### adapto files upload <filepath>
Upload a file directly. Currently delegates to create-metadata + upload-by-id workflow.

#### adapto files upload-by-id <file_id> <filepath>
Upload file content for an existing file record. Outputs a curl command for the actual upload.

#### adapto files get <id>
Get file info by ID.

#### adapto files update <id>
Update file metadata.

| Flag | Description |
|------|-------------|
| ` + "`--filename`" + ` | New filename |
| ` + "`--tags`" + ` | Comma-separated tags |

#### adapto files delete <id>
Delete a file.

#### adapto files multipart-init <file_id>
Initialize a multipart upload. Returns a file ID and upload ID.

#### adapto files multipart-upload <file_id> <upload_id> <part_number> <filepath>
Upload a part. Outputs a curl command for the actual upload.

#### adapto files multipart-complete <file_id> <upload_id>
Complete a multipart upload.

| Flag | Description |
|------|-------------|
| ` + "`--parts`" + ` | Parts JSON array (required) |

#### adapto files multipart-abort <file_id> <upload_id>
Abort a multipart upload.

---

### adapto microcopy

Manage micro copy entries (short translatable text snippets). All subcommands require authentication.

#### adapto microcopy list
List micro copy entries.

| Flag | Description |
|------|-------------|
| ` + "`--language`" + ` | Filter by language |
| ` + "`--tags`" + ` | Filter by tags |

#### adapto microcopy count
Count micro copy entries.

| Flag | Description |
|------|-------------|
| ` + "`--language`" + ` | Filter by language |
| ` + "`--tags`" + ` | Filter by tags |

#### adapto microcopy create
Create a micro copy entry.

| Flag | Description |
|------|-------------|
| ` + "`--key`" + ` | Micro copy key (required) |
| ` + "`--value`" + ` | Text value (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--tags`" + ` | Comma-separated tags |
| ` + "`--translation-of`" + ` | Source micro copy ID (links as translation) |

#### adapto microcopy get <id>
Get micro copy by ID.

#### adapto microcopy get-by-key <key>
Get micro copy by key.

| Flag | Description |
|------|-------------|
| ` + "`--language`" + ` | Filter by language |

#### adapto microcopy get-by-language <language>
Get all micro copy entries for a language.

#### adapto microcopy update <id>
Update a micro copy entry.

| Flag | Description |
|------|-------------|
| ` + "`--key`" + ` | Micro copy key |
| ` + "`--value`" + ` | Text value |
| ` + "`--language`" + ` | Language code |
| ` + "`--tags`" + ` | Tags |

#### adapto microcopy delete <id>
Delete a micro copy entry.

#### adapto microcopy translations <id>
List translations of a micro copy entry.

#### adapto microcopy create-translation <source_id>
Create a micro copy translation.

| Flag | Description |
|------|-------------|
| ` + "`--key`" + ` | Micro copy key (required) |
| ` + "`--value`" + ` | Text value (required) |
| ` + "`--language`" + ` | Language code (required) |
| ` + "`--tags`" + ` | Comma-separated tags |

---

### adapto status

Check API status. Running ` + "`adapto status`" + ` directly returns the API health status.

#### adapto status version
Get API version info.

---

## Common Workflows

### 1. Login and select a tenant

` + "```bash" + `
adapto auth login --email user@example.com --password secret
# Interactive tenant picker appears if you have multiple tenants
# Or specify directly:
adapto auth switch-tenant --tenant-id TENANT_ID
` + "```" + `

### 2. List and create articles

` + "```bash" + `
# List published articles in English
adapto articles list --status published --language en --limit 20

# Create a new draft article
adapto articles create \
  --title "Getting Started" \
  --content "<p>Welcome to our platform.</p>" \
  --slug getting-started \
  --author "Editorial Team" \
  --language en \
  --status draft

# Publish the article
adapto articles publish ARTICLE_ID
` + "```" + `

### 3. Create a translation

` + "```bash" + `
# List translations of an article
adapto articles translations ARTICLE_ID

# Create a French translation
adapto articles create-translation ARTICLE_ID \
  --title "Premiers pas" \
  --content "<p>Bienvenue sur notre plateforme.</p>" \
  --slug premiers-pas \
  --author "Équipe éditoriale" \
  --language fr
` + "```" + `

### 4. Manage categories and link articles

` + "```bash" + `
# Create a category
adapto categories create --name "Tutorials" --slug tutorials --language en

# Add an article to the category
adapto categories add-article CATEGORY_ID ARTICLE_ID

# List articles in the category
adapto categories articles CATEGORY_ID
` + "```" + `

### 5. Work with custom collections

` + "```bash" + `
# Create a collection
adapto collections create \
  --name "Team Members" \
  --slug team-members \
  --description "Our team" \
  --language en

# Add an item
adapto collections items create COLLECTION_ID \
  --title "Jane Doe" \
  --slug jane-doe \
  --language en \
  --data-json '{"role":"Engineer","bio":"Loves code"}'

# Publish the item
adapto collections items publish COLLECTION_ID ITEM_ID
` + "```" + `

### 6. Manage micro copy

` + "```bash" + `
# Create a micro copy entry
adapto microcopy create --key "nav.home" --value "Home" --language en

# Create a translation
adapto microcopy create-translation SOURCE_ID --key "nav.home" --value "Accueil" --language fr

# Get all micro copy for a language
adapto microcopy get-by-language fr
` + "```" + `

### 7. File management

` + "```bash" + `
# List files
adapto files list --type image --limit 10

# Create file metadata, then upload via curl
adapto files create-metadata --filename photo.jpg --content-type image/jpeg
# Use the returned file ID with curl to upload the actual file content
` + "```" + `

### 8. JSON output for scripting

` + "```bash" + `
# Get article as JSON for piping
adapto articles get ARTICLE_ID --json

# List all articles as JSON
adapto articles list --json --limit 100
` + "```" + `

## Notes

- All IDs are strings (UUIDs).
- Language codes follow ISO 639-1 (e.g., "en", "fr", "de").
- Pagination: use ` + "`--page`" + ` and ` + "`--limit`" + ` flags. Responses include total, page, and pages fields.
- Flags marked "(required)" will be prompted interactively if omitted in a TTY. In non-interactive mode, they must be provided.
- Use ` + "`--json`" + ` on any command to get machine-readable JSON output.
`
