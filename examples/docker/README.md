# Hosted Fitter MCP via Docker

Runs the [Fitter MCP server](../../README.md#how-to-use-fitter_mcp) as a shared HTTP endpoint — one fitter for the whole team instead of a binary per machine.

## Start

```bash
export FITTER_MCP_AUTH_TOKEN=$(openssl rand -hex 16)
echo "token: $FITTER_MCP_AUTH_TOKEN"
docker compose up -d
curl http://localhost:8080/healthz   # -> ok
```

## Register in Claude Code

```bash
claude mcp add --transport http fitter http://localhost:8080/mcp \
  --header "Authorization: Bearer $FITTER_MCP_AUTH_TOKEN"
```

Then ask for data:

> Run /configs/config_morning_briefing.json with fitter for Berlin

(the compose file mounts the repo's [examples/](..) folder read-only at `/configs`, so `fitter_run_file` can reach every example config)

## Notes

- The container image is slim: server/static/file connectors work, browser connectors (`chromium`/`docker`/`playwright`) do not — use a release binary on a host with a browser for those.
- The image ships a `HEALTHCHECK` against `/healthz`; `docker ps` shows the health state.
- Scaling out behind a load balancer? Add `FITTER_MCP_STATELESS: "true"` so replicas don't need sticky sessions.
- Without `FITTER_MCP_AUTH_TOKEN` the endpoint is unauthenticated — never expose it publicly like that.
