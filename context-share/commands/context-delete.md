---
description: Delete a shared context from the server
argument-hint: <key>
allowed-tools: [Bash]
---

# Context Delete

Delete a previously saved context from the context-share server.

## Arguments

The user invoked this command with: $ARGUMENTS

The first argument is the context key (required).

## Required Environment Variables

- `CONTEXT_SHARE_URL`: Base URL of the context-share server (e.g., `http://localhost:8787`)
- `CONTEXT_SHARE_API_KEY`: API key for authentication

## Instructions

1. Verify environment variables are set by running:
   ```
   echo "URL: $CONTEXT_SHARE_URL / KEY set: $([ -n "$CONTEXT_SHARE_API_KEY" ] && echo yes || echo no)"
   ```
   If either is missing, tell the user to set them and stop.

2. Delete the context from the server:
   ```bash
   curl -s -X DELETE "${CONTEXT_SHARE_URL}/context/<KEY>" \
     -H "Authorization: Bearer ${CONTEXT_SHARE_API_KEY}"
   ```

3. Report the result. If successful, confirm the context was deleted. If it returned "not found", let the user know that key doesn't exist.
