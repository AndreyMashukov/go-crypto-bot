#!/usr/bin/make
SHELL = /bin/sh

# All targets run inside the docker-compose stack (service name: frontend).
# Bring the stack up first with: make up.

.PHONY: up down build run dev install lint lint-fix typecheck generate preview check shell

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build --no-cache

run: up

dev: up

install:
	docker compose run --rm frontend yarn install --frozen-lockfile

lint:
	docker compose exec frontend yarn lint

lint-fix:
	docker compose exec frontend yarn lint:fix

typecheck:
	docker compose exec frontend yarn typecheck

generate:
	docker compose exec frontend yarn generate

preview:
	docker compose exec frontend yarn preview

shell:
	docker compose exec frontend sh

check: lint typecheck
