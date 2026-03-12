# short.ly

Local development uses Docker Compose for a Go API with hot reload.

## Quickstart

```bash
cp .env.example .env
make dev
```

Then:

```bash
curl -s http://localhost:8080/health | jq .
```

## Common commands

```bash
make logs
make down
```
