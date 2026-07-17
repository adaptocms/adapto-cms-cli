## Custom Fields

Articles, categories, pages, and micro copy support arbitrary typed metadata via the `--custom-fields-json` flag on their `create`, `update`, and `create-translation` commands. The flag takes a JSON object mapping each field name to a field definition:

```bash
adapto articles update ARTICLE_ID --custom-fields-json '{
  "seo_title": {"type": "text", "value": "Welcome"},
  "read_time": {"type": "number", "value": 5},
  "related_posts": {"type": "reference", "multiple": true, "related_collection": "posts", "value": ["id1", "id2"]}
}'
```

Each field definition has:

- `type` (string, required) — one of: `text`, `textarea`, `number`, `date`, `date_range`, `boolean`, `reference`, `image`, `file`, `url`, `email`, `color`, `rich_text`
- `value` — the field value; its JSON type should match `type` (string, number, boolean, or an array when `multiple` is true)
- `multiple` (bool, optional, default false) — whether `value` is a list
- `related_collection` (string, optional) — for `reference` fields, the related collection ID
- `media_objects_placements` (array, optional) — media placements for `rich_text` fields (same shape as `--media-json`)

Notes:

- On `update` the supplied object **replaces the entire custom-fields map** — include every field you want to keep.
- Unknown keys inside a field definition are rejected so the payload matches the API contract exactly.
- For `file`/`image` fields, set `value` to a file ID; responses include a `file_urls` map resolving those IDs to URLs.

---

## Common Workflows

### 1. Get set up

Existing account:

```bash
adapto auth login --email user@example.com --password secret
# multiple tenants? select one:
adapto auth switch-tenant --tenant-id TENANT_ID
```

New account (the only manual step is pasting the activation token from the email):

```bash
adapto auth register --email user@example.com --password secret
adapto auth activate --token <token-from-email>                      # activates and logs in
adapto onboard --project-name "My Project" --default-language en-US  # creates project + API key, sets it active
```

### 2. List and create articles

```bash
# List published articles in English
adapto articles list --status published --language en-US --limit 20

# Create a new draft article
adapto articles create \
  --title "Getting Started" \
  --content "<p>Welcome to our platform.</p>" \
  --slug getting-started \
  --author "Editorial Team" \
  --language en-US \
  --status draft

# Publish the article
adapto articles publish ARTICLE_ID
```

### 3. Create a translation

```bash
# List translations of an article
adapto articles translations ARTICLE_ID

# Create a French translation
adapto articles create-translation ARTICLE_ID \
  --title "Premiers pas" \
  --content "<p>Bienvenue sur notre plateforme.</p>" \
  --slug premiers-pas \
  --author "Équipe éditoriale" \
  --language fr-FR
```

### 4. Manage categories and link articles

```bash
# Create a category
adapto categories create --name "Tutorials" --slug tutorials --language en-US

# Add an article to the category
adapto categories add-article CATEGORY_ID ARTICLE_ID

# List articles in the category
adapto categories articles CATEGORY_ID
```

### 5. Work with custom collections

```bash
# Create a collection
adapto collections create \
  --name "Team Members" \
  --slug team-members \
  --description "Our team" \
  --language en-US

# Add an item
adapto collections items create COLLECTION_ID \
  --title "Jane Doe" \
  --slug jane-doe \
  --language en-US \
  --data-json '{"role":"Engineer","bio":"Loves code"}'

# Publish the item
adapto collections items publish COLLECTION_ID ITEM_ID
```

### 6. Manage micro copy

```bash
# Create a micro copy entry
adapto microcopy create --key "nav.home" --value "Home" --language en-US

# Create a translation
adapto microcopy create-translation SOURCE_ID --key "nav.home" --value "Accueil" --language fr-FR

# Get all micro copy for a language
adapto microcopy get-by-language fr-FR
```

### 7. File upload

```bash
# Upload a file in one step (creates metadata + uploads content)
adapto files upload ./photo.jpg

# Or two-step: create metadata first, then upload by ID
adapto files create-metadata --filename photo.jpg --content-type image/jpeg
adapto files upload-by-id FILE_ID ./photo.jpg

# List files
adapto files list --type image --limit 10
```

### 8. Media objects placements

Articles, pages, and collection items support `--media-json` to attach media (images, videos, embeds). The flag accepts a JSON array of placement objects:

```bash
adapto articles create \
  --title "My Post" --content "<p>Hello</p>" --slug my-post --author "Jane" --language en-US \
  --media-json '[{"placement_key":"hero_image","media_object":{"id":"m1","file_id":"FILE_ID","url":"https://cdn.example.com/photo.jpg","type":"image"},"alt_text":"Hero image"}]'
```

Each placement object has:

- `placement_key` (string) — where the media goes (e.g. "hero_image", "body_image_1")
- `media_object` (object) — id, file_id, url, type (image/video/audio/document/other/youtube/vimeo/tiktok/instagram_reel/instagram_post), title, description
- `caption` (string|null), `alt_text` (string|null), `meta_data` (string|null)

### 9. JSON output for scripting

```bash
# Get article as JSON for piping
adapto articles get ARTICLE_ID --json

# List all articles as JSON
adapto articles list --json --limit 100
```

## Notes

- All IDs are strings (UUIDs).
- Language codes are locale codes of the form language-REGION (e.g., "en-US", "ro-RO", "de-DE").
- Pagination: use `--page` and `--limit` flags. Responses include total, page, and pages fields.
- Flags marked "(required)" will be prompted interactively if omitted in a TTY. In non-interactive mode, they must be provided.
- Use `--json` on any command to get machine-readable JSON output.
- `adapto collections items create-batch` is atomic: either every item in the batch is created or none is. It accepts at most 100 items per request and prints the created items with their ids. A conflict error (409) means the items' slug + language combinations already exist in the collection — a previous attempt may have succeeded, so verify with `items list` instead of retrying; a validation error (400/422) means nothing was persisted.
- HTTP errors include the request method and resolved URL — if the URL is not `https://api.adaptocms.com/...`, the CLI is misconfigured (see "Two APIs" above).
