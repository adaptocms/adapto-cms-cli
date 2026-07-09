# Adapto CMS CLI — Complete Reference

## Overview

Adapto CMS CLI (`adapto`) is a command-line interface for the Adapto CMS Management API. It manages articles, pages, categories, custom collections, files, and micro copy entries across a multi-tenant, multi-language headless CMS.

## Two APIs: Management vs Public

The Adapto platform exposes two different APIs. Confusing them is the most common misconfiguration:

- **Management API** (`https://api.adaptocms.com`) — the authenticated admin API used to create and manage content. **This CLI talks only to the Management API.** Authentication is a user login (JWT), tenant-scoped.
- **Public API** (`https://public-api.adaptocms.com`) — the read-only, API-key-authenticated content delivery API consumed by client sites (Next.js/SvelteKit/Astro templates and the Client SDK).

Client-site repositories configure the Public API through `ADAPTO_API_URL` and `ADAPTO_API_KEY` in their `.env` files. **Those variables are unrelated to this CLI and are ignored by it.** The CLI's own environment variables are prefixed `ADAPTO_CLI_`.

Never point `ADAPTO_CLI_API_URL` at the Public API. You normally never need to set it at all — the default is correct. If a command fails with an error like `Not found (POST https://public-api.adaptocms.com/v1/auth/login)`, the URL in the error is telling you the CLI was pointed at the wrong API.

## Authentication

### Credential storage

After `adapto auth login`, tokens are saved to `~/.config/adapto/credentials.json`. Subsequent commands read from this file automatically.

### Automatic token refresh

The access token is short-lived. When a request fails with 401 and the active token came from the credentials file, the CLI automatically refreshes it using the stored refresh token, persists the rotated pair, and retries the request once — so you rarely need `adapto auth refresh` manually. If the refresh token itself has expired, the command fails with "session expired — run 'adapto auth login'". Tokens supplied via `--token`/`ADAPTO_CLI_TOKEN` are never auto-refreshed.

### Environment variables

| Variable               | Purpose                                                        |
| ---------------------- | -------------------------------------------------------------- |
| `ADAPTO_CLI_API_URL`   | Management API base URL (default: `https://api.adaptocms.com`) |
| `ADAPTO_CLI_TOKEN`     | Bearer token (overrides stored credential)                     |
| `ADAPTO_CLI_TENANT_ID` | Tenant ID (overrides stored credential)                        |

The pre-rename names (`ADAPTO_API_URL`, `ADAPTO_TOKEN`, `ADAPTO_TENANT_ID`) are ignored; the CLI prints a warning when one is set without its `ADAPTO_CLI_` counterpart.

### Multi-tenancy

A user can belong to multiple organizations, each with one or more tenants. After login, the CLI prompts you to select a tenant (or auto-selects if only one exists). Use `adapto auth switch-tenant` to change the active tenant.
