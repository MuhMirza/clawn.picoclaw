---
name: brain
description: "Query the user's second brain memory vault using vector search. Use when: user asks to recall something from their bookmarks, memories, or saved links."
homepage: https://aiagenz.cloud
metadata: { "openclaw": { "emoji": "🧠", "requires": { "bins": ["curl"] } } }
---

# Brain Skill (Memory Recall)

Recall information from the user's second brain (saved bookmarks) using semantic vector search.

## When to Use

✅ **USE this skill when the user says:**
- "Cari di memori tentang AI"
- "Recall what I saved about Next.js"
- "Search my bookmarks for 'Golang'"
- "Apa yang pernah gua simpen soal database?"
- "Get links related to machine learning"

## Command

To search the bookmarks, use the `curl` command to hitting the `/api/bookmarks/search` endpoint. Ensure the `q` parameter is properly URL-encoded.

```bash
curl -G http://localhost:8085/api/bookmarks/search \
  --data-urlencode "q=USER_QUERY_HERE" \
  --data-urlencode "user_id=telegram_user"
```

## How to Handle Results

The API will return a JSON array containing the top 5 most relevant bookmarks. Each result includes `title`, `url`, `shortSummary`, `tags`, and a `similarity` score.

1. **Ingest the results**: Read the returned JSON output.
2. **Synthesize**: Formulate a response based on the search results. DO NOT just dump the raw JSON to the user.
3. **Draft the answer**:
   - If results are found, mention them nicely: "Ini beberapa link yang ada di Memory lu bos: \n- [Title](URL) - _Summary_"
   - If the results are irrelevant (e.g. similarity < 0.6) or empty, tell the user gracefully that you couldn't find anything highly related.
