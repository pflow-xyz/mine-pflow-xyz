# CLAUDE.md

Process mining on the pflow stack (mine.pflow.xyz): event logs in, discovered process model and conformance report out

Registered in the workspace ecosystem table; lifecycle is managed declaratively
by `stackdump-com/apps/mine-pflow-xyz.yaml` via the `console` MCP server. Do not hand-edit
the `~/services` entry or the nginx vhost — change the manifest and re-run the
matching lifecycle step, or `check_drift` will flag the divergence.

- Runtime: pflow.dev
- Port: 8099
- Public: https://mine.pflow.xyz

## Build

```bash
make build   # build the binary
make test    # go tests
```

## Decommissioning

State not in git: (list every untracked data path here, and mirror it into the
`data:` block of the manifest — `archive` copies exactly what is declared there,
and `retire` is unreachable until that copy is verified by content hash).
