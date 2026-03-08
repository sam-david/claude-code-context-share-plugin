---
description: Load shared context from another Claude Code agent's session
argument-hint: <key>
allowed-tools: [Bash, Read, Glob]
---

# Context Load

Load context that was saved by another Claude Code agent session, potentially from a different machine.

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

2. Fetch the context from the server:
   ```bash
   curl -s "${CONTEXT_SHARE_URL}/context/<KEY>" \
     -H "Authorization: Bearer ${CONTEXT_SHARE_API_KEY}"
   ```

3. Parse the returned JSON. If the request failed (not found, expired, unauthorized), report the error clearly and stop.

4. Present the loaded context to the user in a clear, structured format:

   - **Summary**: What the previous session was about and where it left off
   - **Key Decisions**: Important choices already made and their reasoning
   - **Files Involved**: What files were worked on, with descriptions
   - **Current State**: What's working, what's broken, what's in progress
   - **Warnings**: Any gotchas or pitfalls to watch out for
   - **Next Steps**: What needs to be done next

5. After presenting the context, if the context includes file paths and you are on a machine where those files might exist, proactively read the most important ones (up to 5) to build deeper understanding of the current codebase state. If the files don't exist locally, note this — the user may be on a different machine.

6. Offer to continue where the previous session left off. If there are next steps listed, suggest starting with the first one.
