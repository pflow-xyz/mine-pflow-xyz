# CLAUDE.md

Process mining on the pflow stack (mine.pflow.xyz): event logs in, discovered process model and conformance report out. Read `README.md` and `ROADMAP.md` first; only the landing page exists so far.

Lifecycle is managed declaratively by `stackdump-com/apps/mine-pflow-xyz.yaml` via the `console` MCP server. Do not hand-edit the `~/services` entry or the nginx vhost: change the manifest and re-run the matching lifecycle step, or `app_drift` will flag the divergence.

- Runtime: pflow.dev
- Port: 8099
- Public: https://mine.pflow.xyz

## Build

```bash
make build   # build the binary
make test    # go tests
```

## Rules

- **Pushing `main` deploys.** The running binary handles `POST /webhook/github`, refuses a tip that fails `git verify-commit`, then pulls, builds and restarts. Land changes with a local signed merge, not the GitHub merge button (see the `merge-to-a-signed-deploy-main` skill).
- **Do not reimplement mining here.** Fix or extend go-pflow's `eventlog` and `mining` packages, and pflow-rs to match.
- **Public copy leads with outcomes.** The landing page and tool descriptions do not pitch Petri nets; a test enforces this for `web/`.
- **Public, reuse-licensed sources only.** See `ROADMAP.md`. pm4py is AGPL-3.0, so do not copy from it.

## Decommissioning

State not in git:

- `~/.env-services/mine-pflow-xyz` on pflow.dev holds `DEPLOY_SECRET`, which matches the repo's GitHub webhook secret. It can be regenerated, so there is nothing to archive.

The app is stateless today, so the manifest has no `data:` block. Add any data path here and to that block the moment one exists: `archive` copies exactly what is declared, and `retire` is unreachable until that copy is verified by content hash. Retiring must also delete the GitHub webhook.
