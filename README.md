# Adapto CMS CLI

Command-line interface for the [Adapto CMS](https://adaptocms.com) management API.

## Install

### Quick install (macOS/Linux)

```bash
curl -sSL https://raw.githubusercontent.com/eggnita/adapto_cms_cli/main/scripts/install.sh | bash
```

### From source

```bash
go install github.com/eggnita/adapto_cms_cli@latest
```

### From releases

Download the binary for your platform from [GitHub Releases](https://github.com/eggnita/adapto_cms_cli/releases).

## Configuration

Set these environment variables:

| Variable | Description | Required |
|----------|-------------|----------|
| `ADAPTO_TOKEN` | Bearer token (from `adapto auth login`) | Yes (for management commands) |
| `ADAPTO_API_URL` | API base URL (default: `https://api.adaptocms.com`) | No |
| `ADAPTO_TENANT_ID` | Tenant ID for multi-tenant setups | No |

All variables can also be passed as flags: `--token`, `--api-url`, `--tenant-id`.

## Quick Start

```bash
# Login and get a token
adapto auth login --email you@example.com --password yourpassword

# Set the token
export ADAPTO_TOKEN=<access_token from login>

# List articles
adapto articles list

# Get JSON output
adapto articles list --json

# Create an article (interactive prompts for missing fields)
adapto articles create

# Create with flags
adapto articles create \
  --title "My Article" \
  --content "Hello world" \
  --slug "my-article" \
  --author "Jane" \
  --language "en"
```

## Commands

```
adapto
├── version                          Print CLI version
├── auth
│   ├── login                        Login with email/password
│   ├── register                     Register new account
│   ├── logout                       Revoke refresh token
│   ├── refresh                      Refresh access token
│   ├── me                           Current user info
│   ├── change-password              Change password
│   ├── request-password-reset       Request reset email
│   ├── reset-password               Reset with token
│   ├── activate                     Activate account
│   ├── resend-activation            Resend activation email
│   ├── login-github                 GitHub OAuth
│   ├── callback-github              GitHub OAuth callback
│   └── login-google                 Google OAuth
├── articles
│   ├── list                         List articles
│   ├── create                       Create article
│   ├── get <id>                     Get by ID
│   ├── get-by-slug <slug>           Get by slug
│   ├── update <id>                  Update article
│   ├── delete <id>                  Delete article
│   ├── publish <id>                 Publish
│   ├── archive <id>                 Archive
│   ├── translations <id>            List translations
│   ├── create-translation <id>      Create translation
│   └── categories <id>              List categories
├── categories
│   ├── list / create / get / update / delete
│   ├── get-by-slug / subcategories
│   ├── articles / add-article / remove-article
│   └── translations / create-translation
├── pages
│   ├── list / create / get / update / delete
│   ├── get-by-slug / publish / archive
│   └── translations / create-translation
├── collections
│   ├── list / create / get / update / delete
│   ├── get-by-slug
│   └── items
│       ├── list / create / get / update / delete
│       ├── create-batch / get-by-slug
│       ├── publish / archive
│       └── translations / create-translation
├── files
│   ├── list / get / update / delete
│   ├── create-metadata
│   ├── upload / upload-by-id
│   └── multipart-init / multipart-complete / multipart-abort
├── microcopy
│   ├── list / count / create / get / update / delete
│   ├── get-by-key / get-by-language
│   └── translations / create-translation
└── status
    └── version                      API version info
```

## Output

By default, commands output formatted tables. Use `--json` for JSON output:

```bash
adapto articles list --json
adapto articles get abc123 --json
```

## Interactive Mode

When running in a terminal (TTY), missing required parameters will be prompted interactively. When piped or in scripts, missing parameters produce an error with usage hints.

## Development

```bash
# Build
make build

# Run tests
make test

# Regenerate API client from OpenAPI spec
make generate
```

## License

MIT
