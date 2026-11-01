# mine.pflow.xyz

Process mining on the pflow stack: **an event log goes in, a discovered process model and a conformance report come out.**

- **What actually happened?** Discover the process from the log, including the paths nobody drew on the whiteboard.
- **Did it go the way it should?** Check the log against a model: which cases fit, which did not, and by how much.
- **Ask from where you work.** The goal is the MCP server people reach for when they need process mining, so an assistant can answer from the log itself.

**Status: early.** The landing page, health check and deploy hook are all that exist today. Nothing below is available yet; [ROADMAP.md](ROADMAP.md) has the plan for the first release.

## Planned

| Step | In | Out |
|------|----|-----|
| Ingest | XES or CSV, plus which columns mean case, activity and time | A checked log summary: cases, events, variants |
| Discover | A log and a method | A process model with a stable content ID |
| Conformance | A log and a model | Fitness, precision and the non-fitting cases |
| MCP | The three steps above as tools | The same answers from Claude or any MCP client |

## The engine

mine does not reimplement process mining. Discovery and conformance come from the `eventlog` and `mining` packages of [go-pflow](https://github.com/pflow-xyz/go-pflow), with [pflow-rs](https://github.com/pflow-xyz/pflow-rs) as the independent second implementation. Models are Petri nets under the hood, which is why replay against them is exact.

## Run it

```bash
make build
make run     # serves http://localhost:8099
make test
```

## License

MIT. Sources and test data are limited to public, reuse-licensed material; see [ROADMAP.md](ROADMAP.md#sources-and-licensing).
