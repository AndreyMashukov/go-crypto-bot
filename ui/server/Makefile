#!/usr/bin/make
SHELL = /bin/sh

UID ?= $(shell id -u)
GID ?= $(shell id -g)

# All targets run inside the docker-compose stack (service name: backend).
# Bring the stack up first with: make up.

.PHONY: up down build run clean db database cs cs-dry phpstan rector rector-dry test check shell composer-install

up:
	docker compose up -d

down:
	docker compose down

build:
	docker compose build --no-cache

run: up

clean:
	docker compose exec backend rm -rf var/cache/* var/log/*

composer-install:
	docker compose exec backend composer install

shell:
	docker compose exec backend sh

# --- DB ---
db: database

database:
	docker compose exec backend php bin/console doctrine:database:drop --force --if-exists
	docker compose exec backend php bin/console doctrine:database:create
	docker compose exec backend php bin/console doctrine:migrations:migrate -n
	docker compose exec backend php bin/console doctrine:fixtures:load -n --group=first
	docker compose exec backend php bin/console doctrine:schema:validate --skip-sync

# --- Quality gates ---
cs:
	docker compose exec backend vendor/bin/php-cs-fixer fix -v --using-cache=no --diff --config=.php-cs-fixer.dist.php --allow-risky=yes

cs-dry:
	docker compose exec backend vendor/bin/php-cs-fixer fix --dry-run --diff --config=.php-cs-fixer.dist.php --allow-risky=yes

phpstan:
	docker compose exec backend vendor/bin/phpstan analyse --memory-limit=1G

rector:
	docker compose exec backend vendor/bin/rector process

rector-dry:
	docker compose exec backend vendor/bin/rector process --dry-run

test:
	docker compose exec backend vendor/bin/phpunit -c phpunit.xml.dist

check: cs-dry phpstan rector-dry test
