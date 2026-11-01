# Roadmap

Written 2026-11-01.

**Goal.** Become the MCP server people reach for when they need process mining: an event log in, a discovered process model and a conformance report out.

**Rules.**

1. **Reuse the engine.** Mining logic lives in go-pflow and pflow-rs. If something is missing, it is added there and this repo depends on it. mine owns the product surface: ingest ergonomics, reports, HTTP and MCP.
2. **Outcomes first in public copy.** Pages and tool descriptions talk about the process, the cases and the deviations. Petri nets are the engine, not the pitch. `TestLandingCopyDoesNotPitchTheEngine` holds the landing page to this.
3. **A mechanism decides.** Every number in a report comes from an engine call. Any wording an LLM adds is labelled as judgment, never as engine output.

## What exists to reuse

Checked against go-pflow v0.34.0 and pflow-rs on 2026-11-01.

| Need | go-pflow | pflow-rs |
|------|----------|----------|
| CSV log | `eventlog.ParseCSV`, `ParseCSVReader` | `pflow-eventlog` |
| JSONL log | `eventlog.ParseJSONL`, `ParseJSONLReader` | `pflow-eventlog` |
| **XES log** | **missing** (`eventlog/README.md` lists it as planned) | **missing** |
| Discovery | `mining.Discover`: `sequential`, `common-path`, `alpha`, `heuristic` | `pflow-mining` |
| Fitness and precision | `mining.CheckFullConformance` | `pflow-mining`, and `petri_conformance` in `pflow-mcp` |
| Timing | `mining.ExtractTiming`, `LearnRatesFromLog` | `pflow-mining` |

The Go server calls go-pflow, as the other pflow backends do. pflow-rs is the parity check, not a second runtime.

## MVP

### M0: Foundation

The landing page, `/health`, push-to-deploy and the `mine.pflow.xyz` vhost with TLS.

*Exit:* `curl https://mine.pflow.xyz/health` returns `ok`, and a push to `main` redeploys.

### M1: Ingest

- Add the XES parser **to go-pflow's `eventlog` package** (IEEE 1849), then port it to `pflow-eventlog` with shared fixtures.
- Add go-pflow to `go.mod`. It raises this module's minimum Go version to 1.24.9.
- Map CSV columns to case, activity and timestamp. Guess from headers, show the guess, and let the caller override it.
- Return a summary: cases, events, activities, variants, time range and rows that could not be read.

*Exit:* a public XES log and its CSV export produce the same summary.

### M2: Discover

- Expose `mining.Discover` with its four methods, and return the top variants with their share of cases.
- Return the model with a stable content ID (CID) so a report can cite exactly what it was checked against.
- Export to pflow.xyz, so a discovered model opens in the editor.

*Exit:* the same log and method give the same CID on every run, enforced by a test. If an engine method turns out not to be deterministic, the fix goes in go-pflow.

### M3: Conformance

- Expose `mining.CheckFullConformance` against a discovered model or a supplied one.
- Report fitness, precision, F-score, the fitting share and each non-fitting case with its missing and remaining tokens, in plain language.

*Exit:* golden results on shared fixtures match between go-pflow and pflow-rs (`pflow-mining` already has `tests/conformance_golden.rs` to build on).

### M4: MCP server

- A remote MCP endpoint at `/mcp` with read-only, bounded tools: `mine_ingest`, `mine_discover`, `mine_conformance`.
- Tool descriptions state the outcome and when to use the tool. Results carry the model CID and the numbers the engine returned.
- Set the upload ceiling on purpose: the vhost's default body limit is 10 MB, so large logs need a `max_body` in the manifest or a separate upload path.

*Exit:* from a Claude client, a public log goes from ingest to a conformance answer with no other tool.

### M5: Directory listing

mine.pflow.xyz is a **separate listing** from the main pflow connector, aimed at the **Automation & workflows** and **Logs & errors** categories of the Claude connectors directory. The plan for both listings is in the [pflow-xyz roadmap](https://github.com/pflow-xyz/pflow-xyz/blob/main/ROADMAP.md#claude-connectors-directory).

*Exit:* M4 is live, with published docs, terms and privacy pages that describe how uploaded logs are handled.

## After the MVP

- Bottlenecks and waiting time from `ExtractTiming`.
- Hand a discovered, timed model to [sim.pflow.xyz](https://sim.pflow.xyz) for what-if questions.
- Object-centric logs (OCEL) and streaming conformance.

## Open decisions

1. **Log retention.** Event logs often hold business or personal data. The proposed default is to process a log in memory and keep only what the caller asks to keep. This has to be settled before M4, because the privacy page depends on it.
2. **Authentication and limits for the hosted MCP.** pflow.xyz's `/mcp` is the reference. A local, self-hosted mode may suit sensitive logs better.
3. **Where it runs at scale.** It runs on pflow.dev for now. Mining a large log is CPU heavy, and pflow.dev is shared with every other app, so M4 needs a concurrency cap and may justify moving to Cloud Run.

## Sources and licensing

This project uses public, reuse-licensed sources only. No employer, client or otherwise private material goes into code, fixtures or docs.

- **Code.** pm4py is AGPL-3.0, so it is not a source to copy or port from here. Algorithms come from the published literature and from go-pflow and pflow-rs (MIT).
- **Formats.** XES is the public IEEE 1849 standard.
- **Test data.** Generated fixtures first. A public log is vendored only after checking that dataset's own license, which is then recorded next to the file. None are vendored yet.
