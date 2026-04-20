# sage-wiki as a Memory Layer for AI Agents

sage-wiki runs as an MCP server with 17 tools. But tools alone aren't enough — agents won't proactively use the wiki unless their context tells them *when* to check it, *what* to capture, and *how* to query effectively. This guide covers the full setup: connecting MCP, generating skill files, and establishing the read-capture-evolve loop that turns sage-wiki into compounding institutional memory.

## The Problem Skill Files Solve

MCP handles tool discovery and invocation. What it doesn't handle is *when* the agent should voluntarily reach for a tool. A developer installs sage-wiki, adds it to `.mcp.json`, and the agent ignores it because it treats the wiki as one more tool in a pile of tools.

Skill files bridge that gap. They're behavioral instructions — 30-50 lines appended to the agent's instruction file — that create three changes:

1. **When to check the wiki** — without this, agents never look
2. **What to capture** — without this, agents capture nothing or capture everything
3. **How to query effectively** — without this, agents send bad queries and get bad results

## Project Setup for Repo-as-Wiki

The recommended setup keeps sage-wiki as a subdirectory within your project, with sources pointing to your docs and code:

```
my-project/
├── .mcp.json                    # sage-wiki MCP server config
├── CLAUDE.md                    # agent instructions (includes wiki skill)
├── docs/
│   ├── adrs/                    # architecture decision records
│   ├── guides/                  # engineering guides
│   └── architecture.md
├── src/                         # application code
└── .sage-wiki/                  # sage-wiki project root
    ├── config.yaml
    ├── .sage/wiki.db
    ├── .manifest.json
    └── _wiki/                   # compiled output
        ├── concepts/
        ├── summaries/
        └── CHANGELOG.md
```

### Initialize

```bash
cd my-project
mkdir .sage-wiki && cd .sage-wiki
sage-wiki init --skill claude-code
```

This creates the project structure, config.yaml, and appends a skill section to `CLAUDE.md` in one step.

### config.yaml

```yaml
project: my-project
sources:
  - path: ../docs
    type: article
    watch: true
  - path: ../src
    type: code
    watch: true
ignore:
  - node_modules
  - dist
  - .git
  - "*.test.*"
output: _wiki
compiler:
  default_tier: 1              # index + embed everything, compile on demand
  auto_promote: true
  promote_signals:
    query_hit_count: 3
    cluster_size: 5
```

Setting `default_tier: 1` indexes everything fast (FTS5 + vector embedding). Articles compile on demand when an agent queries a topic and `wiki_compile_topic` fires.

### .mcp.json

```json
{
  "mcpServers": {
    "sage-wiki": {
      "command": "sage-wiki",
      "args": ["serve", "--project", ".sage-wiki"]
    }
  }
}
```

## MCP Setup by Agent

### Claude Code

Add to your project's `.mcp.json`:

```json
{
  "mcpServers": {
    "sage-wiki": {
      "command": "sage-wiki",
      "args": ["serve", "--project", "/path/to/your/wiki"]
    }
  }
}
```

Generate the skill file:

```bash
sage-wiki init --skill claude-code
# Or for an existing project:
sage-wiki skill refresh --target claude-code
```

This appends behavioral instructions to your CLAUDE.md. The agent will now proactively search the wiki before architectural decisions, capture learnings after significant work, and use ontology queries to explore relationships.

### Cursor

Add to `.cursor/mcp.json` in your project:

```json
{
  "mcpServers": {
    "sage-wiki": {
      "command": "sage-wiki",
      "args": ["serve", "--project", "/path/to/your/wiki"]
    }
  }
}
```

Generate the skill file:

```bash
sage-wiki skill refresh --target cursor
```

This creates `.cursorrules` with behavioral instructions for when to use the wiki. Cursor uses plain text format — markdown headers are automatically converted.

### Windsurf

Same MCP config pattern. Generate the skill file:

```bash
sage-wiki skill refresh --target windsurf
```

Creates `.windsurfrules` in plain text format.

### ChatGPT

ChatGPT supports MCP servers via its remote server protocol. Point it to the SSE endpoint:

```bash
sage-wiki serve --transport sse --port 3333 --project /path/to/your/wiki
```

Then add `http://localhost:3333/sse` as an MCP server in ChatGPT settings.

### Gemini CLI

```bash
sage-wiki skill refresh --target gemini
```

Creates `GEMINI.md` with wiki skill instructions.

### Antigravity / Codex

```bash
sage-wiki skill refresh --target agents-md
# or equivalently:
sage-wiki skill refresh --target codex
```

Both write to `AGENTS.md`.

### Other MCP Clients

Any client that supports MCP can connect via stdio or SSE:

```bash
# stdio (default)
sage-wiki serve --project /path/to/wiki

# SSE (network)
sage-wiki serve --transport sse --port 3333 --project /path/to/wiki
```

For agents without a skill file convention, generate a generic skill and include it manually:

```bash
sage-wiki skill preview --target generic > sage-wiki-skill.md
```

## Skill Packs

sage-wiki ships 4 domain-specific skill packs. Each generates different behavioral triggers, capture guidelines, and query examples tailored to the domain.

### codebase-memory (default for code projects)

For repositories. Teaches the agent to:

- Check wiki before modifying public APIs or cross-module interfaces
- Capture ADR-style decisions after architectural changes
- Query dependency relationships via ontology ("what depends on this module?")
- Compile relevant topic clusters when search returns uncompiled sources
- Use `wiki_provenance` to trace source-article relationships

Trigger phrases: "refactor", "redesign", "why is this", "what depends on", "breaking change"

MCP tools referenced: `wiki_search`, `wiki_read`, `wiki_status`, `wiki_ontology_query`, `wiki_learn`, `wiki_compile_topic`, `wiki_provenance`

### research-library (default for article/paper projects)

For research projects. Teaches the agent to:

- Search wiki before answering domain questions — brain-first, not hallucinate-first
- Capture new findings after reading papers or running experiments
- Query prerequisite chains ("what do I need to understand before X?")
- Query contradiction edges ("what challenges this claim?")
- Compile on-demand when a topic cluster has enough uncompiled sources

Trigger phrases: "related work", "prior art", "what's known about", "contradicts", "builds on"

MCP tools referenced: `wiki_search`, `wiki_read`, `wiki_status`, `wiki_ontology_query`, `wiki_learn`, `wiki_capture`, `wiki_compile_topic`

### meeting-notes (override only)

For operational use. Select with `--pack meeting-notes`. Teaches the agent to:

- Search wiki for person/company context before meetings
- Capture decisions and action items after meetings
- Generate pre-meeting briefings from wiki context
- Enrich person pages when new information surfaces

Trigger phrases: "prep me for", "what do we know about [person]", "meeting with", "action items"

MCP tools referenced: `wiki_search`, `wiki_read`, `wiki_status`, `wiki_learn`, `wiki_capture`, `wiki_list`

### documentation-curator (override only)

For maintaining team documentation. Select with `--pack documentation-curator`. Teaches the agent to:

- Check wiki before writing new docs (avoid duplication)
- Capture new concepts and definitions when authoring guides
- Link new documentation to existing concept articles via ontology
- Flag stale or contradictory documentation via linter integration

Trigger phrases: "document this", "write a guide", "update the docs", "what's documented about"

MCP tools referenced: `wiki_search`, `wiki_read`, `wiki_status`, `wiki_ontology_query`, `wiki_learn`, `wiki_lint`, `wiki_compile_topic`

### Auto-selection

The generator picks a pack based on source types in your config.yaml:

| Source types | Pack selected |
|---|---|
| All code | codebase-memory |
| All article/paper | research-library |
| Mixed code + article | codebase-memory |
| Empty or auto-only | codebase-memory |

Override with `--pack` for meeting-notes or documentation-curator:

```bash
sage-wiki init --skill claude-code --pack meeting-notes
```

## What the Generator Produces

The skill content is generated from your project's config.yaml, not hand-written. The generator reads:

- **Project name** — referenced in the skill header
- **Source types** — determines which pack to use and what triggers to emphasize
- **Entity types** — built-in (concept, technique, source, claim, artifact) + custom types from `ontology.entity_types`
- **Relation types** — built-in (implements, extends, contradicts, etc.) + custom from `ontology.relation_types`
- **Graph expansion** — whether ontology queries are available (affects query examples)
- **Default tier** — compilation tier (affects compile-on-demand guidance)

The output is 30-50 lines of behavioral instructions with three sections: when to check, what to capture, how to query. It references MCP tool names but not full schemas (agents get those from MCP discovery).

### Marker-based updates

The generated skill section is wrapped in markers:

```
<!-- sage-wiki:skill:start -->
...generated content...
<!-- sage-wiki:skill:end -->
```

For plain-text files (.cursorrules, .windsurfrules):

```
# sage-wiki:skill:start
...
# sage-wiki:skill:end
```

Running `sage-wiki skill refresh` replaces only the content between markers. Your other instructions are preserved.

## Capture Workflows

sage-wiki provides several MCP tools for different capture patterns:

### Quick Capture: `wiki_capture`

The primary tool for saving knowledge from conversations. Give it a chunk of text and it uses the LLM to extract the key learnings automatically.

**Example prompts you can say to your AI:**

> "Save what we just figured out about connection pooling to my wiki"

> "Capture the key decisions from this conversation"

> "Extract the important findings from our debugging session and add them to my wiki"

The AI will call `wiki_capture` with the relevant text. The tool:
1. Sends the text to your configured LLM for knowledge extraction
2. Writes each extracted item as a source file in `raw/captures/`
3. Returns a summary of what was captured

Run `sage-wiki compile` afterward to process captures into wiki articles with concepts, cross-references, and search indexing.

### Single Nugget: `wiki_learn`

For storing a specific learning or insight without LLM extraction:

> "Remember that SQLite FTS5 requires double-quoting for phrase search"

The AI calls `wiki_learn` with a type (gotcha, correction, convention, error-fix, api-drift) and the content. These are stored in the learning database and surfaced during linting.

### Full Document: `wiki_add_source`

For adding an existing file as a wiki source:

> "Add the file at raw/papers/attention.pdf to my wiki sources"

### Compile: `wiki_compile`

After capturing knowledge, compile to process everything:

> "Compile my wiki to process the new captures"

## The Read-Capture-Evolve Loop

The skill file's purpose is to create a virtuous cycle:

```
Session starts
  → Agent reads CLAUDE.md (includes wiki skill)
  → Agent checks wiki_status (knows what's available)
  │
  ├─ Agent gets a task
  │  → Checks wiki before acting (behavioral trigger)
  │  → Finds relevant context (or finds nothing — signals gap)
  │  → Completes the task with wiki context
  │  → Captures decision/gotcha/convention (capture trigger)
  │
  ├─ Next session: wiki is smarter
  │  → New capture is indexed (Tier 1) or compiled (Tier 3)
  │  → Agent finds more context next time
  │  → Better decisions, fewer repeated mistakes
  │
  └─ Over time: wiki compounds
     → Architectural decisions accumulate
     → Convention knowledge densifies
     → New team members get context from day 1
     → The wiki is the institutional memory
```

The difference between "sage-wiki is installed" and "sage-wiki is actively used" is entirely in whether this loop runs. The skill file is the bootstrap that starts it.

## Adding a Custom Skill Pack

Skill packs are Go templates embedded in the binary at `internal/skill/packs/`. To add a new pack:

1. Create `internal/skill/packs/your-pack-name.md.tmpl` following the three-section structure (when/what/how). Use Go `text/template` syntax for project-specific values:

```
## sage-wiki — project knowledge base

This project ({{.Project}}) uses sage-wiki as its knowledge layer.
Sources: {{.SourceTypes}}.

### When to check the wiki

[Your domain-specific triggers here]

1. Call `wiki_search` with the relevant term
2. Read articles with `wiki_read`{{if .HasOntology}}
3. Use `wiki_ontology_query` for relationship queries{{end}}

### What to capture

[Your domain-specific capture guidelines]

### How to query effectively

- Entity types: {{range $i, $e := .EntityTypes}}{{if $i}}, {{end}}{{$e}}{{end}}{{if .HasOntology}}
- Relation types: {{range $i, $r := .RelationTypes}}{{if $i}}, {{end}}{{$r}}{{end}}{{end}}
```

2. Register the pack in `internal/skill/packs.go`:

```go
var packFiles = map[PackName]string{
    // ...existing packs...
    PackName("your-pack-name"): "packs/your-pack-name.md.tmpl",
}
```

3. (Optional) Add auto-selection logic in `SelectPack()` in `internal/skill/skill.go`, or leave it as override-only via `--pack your-pack-name`.

Available template variables:

| Variable | Type | Description |
|---|---|---|
| `{{.Project}}` | string | Project name from config |
| `{{.SourceTypes}}` | string | Comma-separated source types (e.g., "article, code") |
| `{{.EntityTypes}}` | []string | Built-in + custom entity type names |
| `{{.RelationTypes}}` | []string | Built-in + custom relation type names |
| `{{.HasOntology}}` | bool | Whether ontology queries are available (always true) |
| `{{.DefaultTier}}` | int | Default compilation tier (0-3) |
| `{{.HasGraphExpansion}}` | bool | Whether graph-based search expansion is enabled |

Keep packs under 50 lines. The agent reads this on every session start — bloating the context window wastes tokens and dilutes the behavioral signal.

## Tips for Effective Capture

1. **Be specific about what to save.** "Save the part about retry backoff" is better than "save everything."

2. **Add context.** "We were debugging the auth middleware" helps the extraction focus on what matters.

3. **Tag your captures.** Tags help with search and organization later: "tag this with go, performance."

4. **Compile regularly.** Captured items sit in `raw/captures/` until compiled. The compiler extracts concepts, discovers connections to existing articles, and builds the wiki graph.

5. **Review captures.** Check `raw/captures/` occasionally. The LLM extraction is good but not perfect — you may want to edit or merge items.

## What Gets Captured

The extraction prompt focuses on:

- **Decisions** made during the conversation
- **Discoveries** or "aha moments"
- **Corrections** where an assumption was wrong
- **Technical facts** that were established
- **Patterns and anti-patterns** identified

It skips greetings, retries, debugging dead-ends, and other noise.

## Example

Say you're debugging a performance issue with your AI and discover that the bottleneck is in the database connection pool, not the query itself. At the end of the session:

> "Capture the key findings from this debugging session. Tag with postgres, performance."

The AI extracts items like:
- "connection-pool-bottleneck" — The actual performance issue was exhausted connections, not slow queries
- "pgbouncer-transaction-mode" — Transaction-level pooling resolved the issue; session-level was causing connection hoarding

These become source files that the compiler weaves into your wiki's knowledge graph. Next time an agent encounters a database performance question, `wiki_search("connection pooling")` surfaces these findings — the wiki remembers what you learned.
