# context-share

Share conversation context between Claude Code agents across sessions and machines.

One agent runs `/context-save my-key`, another runs `/context-load my-key` — on any machine — and picks up right where the first left off.

## How It Works

Two components:

1. **Server** — A small Go HTTP server that stores and retrieves context bundles, backed by SQLite.
2. **Plugin** — Three Claude Code slash commands (`/context-save`, `/context-load`, and `/context-delete`) that compile, transfer, and manage context.

The server is a self-contained binary with no external dependencies. SQLite is embedded — no database server to install or configure. The database file and table are created automatically on first run.

## Server Setup

### Requirements

- Go 1.22+ (to build from source), or just grab the binary

### Build

```bash
cd server
go build -o context-share .
```

### Run

```bash
export CONTEXT_SHARE_API_KEY="your-secret-api-key"  # required
export PORT="8787"                                    # optional, default: 8787
export DB_PATH="context-share.db"                     # optional, default: context-share.db

./context-share
```

The server will create `context-share.db` in the current directory on first run. No migrations or initialization needed.

### Cross-compile for deployment

Build on your Mac, deploy to a Linux server:

```bash
GOOS=linux GOARCH=amd64 go build -o context-share .
scp context-share your-server:/opt/context-share/
```

### Run with systemd (Linux)

Create `/etc/systemd/system/context-share.service`:

```ini
[Unit]
Description=context-share server
After=network.target

[Service]
ExecStart=/opt/context-share/context-share
Environment=CONTEXT_SHARE_API_KEY=your-secret-api-key
Environment=DB_PATH=/opt/context-share/data/context-share.db
WorkingDirectory=/opt/context-share
Restart=always

[Install]
WantedBy=multi-user.target
```

Then:

```bash
sudo systemctl enable --now context-share
```

### Docker

```bash
cd server
docker build -t context-share .
docker run -p 8787:8787 -e CONTEXT_SHARE_API_KEY=your-secret -v context-data:/data -e DB_PATH=/data/context-share.db context-share
```

## Plugin Setup

### Install the plugin

```bash
# 1. Add the marketplace (from GitHub)
claude plugin marketplace add sam-david/claude-code-context-share-plugin

# Or from a local path if you've cloned the repo
claude plugin marketplace add /path/to/claude-code-context-share-plugin

# 2. Install the plugin
claude plugin install context-share
```

### Configure environment variables

Add to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.):

```bash
export CONTEXT_SHARE_URL="http://your-server:8787"    # or https:// if behind a reverse proxy
export CONTEXT_SHARE_API_KEY="your-secret-api-key"     # same key the server uses
```

Both variables are required on every machine that will use the plugin.

## Usage

### Save context

In any Claude Code session:

```
/context-save my-feature
```

The agent compiles a structured context bundle from the conversation — summary, decisions made, files touched, current state, next steps, key code snippets, and warnings — then uploads it to the server under the key `my-feature`.

Optional TTL (auto-expire after N hours):

```
/context-save my-feature --ttl 72
```

### Load context

In a new Claude Code session (same machine or different):

```
/context-load my-feature
```

The agent fetches the context, presents it, reads relevant local files if they exist, and offers to continue where the previous session left off.

### Delete context

```
/context-delete my-feature
```

Removes the context from the server.

## API Reference

All endpoints require `Authorization: Bearer <api-key>` header.

| Method | Path | Description |
|--------|------|-------------|
| `PUT` | `/context/{key}` | Save context. Body: `{"context": <any JSON>, "ttl_hours": <int or null>}` |
| `GET` | `/context/{key}` | Load context. Returns `{"key", "context", "created_at", "expires_at?"}` |
| `DELETE` | `/context/{key}` | Delete context. |
| `GET` | `/health` | Health check. Returns `{"status":"ok"}` |

## Architecture Notes

- **SQLite with WAL mode** — handles concurrent reads/writes safely, no external database needed
- **Constant-time auth comparison** — prevents timing attacks on API key
- **Expired contexts** are cleaned up lazily on next read (no background job needed)
- **Context is stored as arbitrary JSON** — the server is schema-agnostic, all structure is defined by the plugin commands
- **14MB binary, ~5MB RAM** — runs comfortably on the smallest VPS or even a Raspberry Pi
