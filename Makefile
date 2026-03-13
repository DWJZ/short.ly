.PHONY: help dev up down restart logs ps test fmt tidy

PROJECT ?= shortly
COMPOSE ?= docker compose

BACKEND_PORT ?= 8080

help:
	@printf "%s\n" \
	  "Targets:" \
	  "  make dev         - start api (hot reload)" \
	  "  make down        - stop stack" \
	  "  make logs        - follow logs" \
	  "  make fmt         - gofmt backend" \
	  "  make test        - run backend tests" \
	  "  make tidy        - go mod tidy"

dev:
	BACKEND_PORT=$(BACKEND_PORT) $(COMPOSE) up --build

up: dev

down:
	$(COMPOSE) down

restart:
	$(COMPOSE) down && $(COMPOSE) up --build

logs:
	$(COMPOSE) logs -f --tail=200

ps:
	$(COMPOSE) ps

fmt:
	$(COMPOSE) run --rm backend sh -c "/usr/local/go/bin/gofmt -w ."

test:
	$(COMPOSE) run --rm backend sh -c "/usr/local/go/bin/go test -v ./..."

tidy:
	$(COMPOSE) run --rm backend sh -c "/usr/local/go/bin/go mod tidy"
