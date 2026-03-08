---
description: Save current conversation context to share with another Claude Code agent
argument-hint: <key> [specific context to save] [--ttl <hours>]
allowed-tools: [Bash, Read, Glob]
---

# Context Save

Save the current conversation context so another Claude Code agent (on any machine) can load it and continue the work.

## Arguments

The user invoked this command with: $ARGUMENTS

Parse the first argument as the context key (required). Optionally parse `--ttl <hours>` for automatic expiration. Any remaining text after the key (and before --ttl if present) is a description of what specific context to save.

## Required Environment Variables

- `CONTEXT_SHARE_URL`: Base URL of the context-share server (e.g., `http://localhost:8787`)
- `CONTEXT_SHARE_API_KEY`: API key for authentication

## Instructions

1. First, verify environment variables are set by running:
   ```
   echo "URL: $CONTEXT_SHARE_URL / KEY set: $([ -n "$CONTEXT_SHARE_API_KEY" ] && echo yes || echo no)"
   ```
   If either is missing, tell the user to set them and stop.

2. Compile a context object from this conversation. If the user provided a description of specific context to save, focus **only** on that topic — include only the decisions, files, snippets, and details relevant to what they described. If no description was provided, save everything. Build a JSON object with these fields:

   - **summary** (string): 2-4 sentence overview of this conversation — what was the goal, what was accomplished, where things stand now.
   - **decisions** (array of strings): Key technical decisions made during this conversation. Include the reasoning, not just the choice.
   - **files_touched** (array of objects): Each with `path` (string) and `description` (string) for every file that was read, created, or modified. Include what was done and why.
   - **current_state** (string): What's working, what's broken, what's in progress right now.
   - **next_steps** (array of strings): Concrete actions the next agent should take to continue this work.
   - **key_snippets** (array of objects): Each with `path` (string), `description` (string), and `code` (string) for small critical code sections that the next agent will need. Keep snippets short — just the essential parts, not entire files.
   - **project** (object): `directory` (string — working directory path), `type` (string — e.g., "go", "node", "python"), and `structure` (string — brief description of project layout).
   - **warnings** (array of strings): Gotchas, pitfalls, or things the next agent should be careful about.

3. Send the context to the server. Write the JSON to a temp file and use curl:

   ```bash
   TMPFILE=$(mktemp)
   cat > "$TMPFILE" << 'CONTEXT_JSON'
   {"context": <your compiled JSON object>, "ttl_hours": <TTL or null>}
   CONTEXT_JSON
   curl -s -X PUT "${CONTEXT_SHARE_URL}/context/<KEY>" \
     -H "Authorization: Bearer ${CONTEXT_SHARE_API_KEY}" \
     -H "Content-Type: application/json" \
     -d @"$TMPFILE"
   rm "$TMPFILE"
   ```

4. Report the result. Tell the user:
   - The key to share: `<KEY>`
   - The command the other agent should run: `/context-load <KEY>`
   - When it expires (if TTL was set)
