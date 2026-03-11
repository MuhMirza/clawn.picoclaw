---
name: bookmark
description: "Save a URL/link to the user's second brain memory vault. Use when: user asks to save, remember, or bookmark a link."
homepage: https://aiagenz.cloud
metadata: { "openclaw": { "emoji": "🔖", "requires": { "bins": ["curl"] } } }
---

# Bookmark Skill

Save URLs to the user's second brain memory vault for later retrieval and AI processing.

## When to Use

✅ **USE this skill when the user says:**
- "Save this link"
- "Bookmark this: https://..."
- "Remember this article"
- "Tolong simpen link ini"
- Or just sends a raw URL without much context (e.g. just pasting a link).

## Commands

You have two options to save a bookmark, depending on the user's request. Always execute the `curl` command using your shell/exec tool.

### Option 1: Fast Save (Default)
If the user just pastes a link or says "save this", simply send the URL. The system will process it in the background later.

```bash
curl -X POST http://localhost:8085/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{"url": "THE_URL_HERE", "userId": "telegram_user"}'
```

### Option 2: Agent Pre-Processed Save
If the user specifically asks YOU (the Agent) to process, summarize, or tag the bookmark before saving (e.g., "baca link ini terus simpen ke bookmark sama summary-nya"), you first use your web-fetch tools to read the URL content. 
Then, generate a title, short summary (2-3 sentences), and some tags. Send all of it to the endpoint to bypass the background worker's parsing step!

```bash
curl -X POST http://localhost:8085/api/bookmarks \
  -H "Content-Type: application/json" \
  -d '{
    "url": "THE_URL_HERE", 
    "userId": "telegram_user",
    "title": "Your generated title",
    "shortSummary": "Your generated summary here",
    "tags": ["tag1", "tag2", "tag3"]
  }'
```

## Quick Responses

After successfully executing the command (check that the API returns `201 Created` or `200 OK`), respond cheerfully to the user:

"🔖 Link disimpen ke Memory Bookmark ya bos! Biar Agent proses dulu..."

## Notes
- Ensure the URL starts with `http://` or `https://`.
- If the API returns an error, inform the user that the bookmark service might be down.
