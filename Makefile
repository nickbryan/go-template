.DEFAULT_GOAL := help

# Colours used in help
GREEN    := $(shell tput -Txterm setaf 2)
WHITE    := $(shell tput -Txterm setaf 7)
YELLOW   := $(shell tput -Txterm setaf 3)
RESET    := $(shell tput -Txterm sgr0)

HELP_FUN = %help; \
	while(<>) { push @{$$help{$$2 // 'Misc'}}, [$$1, $$3] \
	if /^([a-zA-Z\-]+)\s*:.*\#\#(?:@([a-zA-Z\-]+))?\s(.*)$$/ }; \
	for (sort keys %help) { \
	print "${WHITE}$$_${RESET}\n"; \
	for (@{$$help{$$_}}) { \
	$$sep = " " x (32 - length $$_->[0]); \
	print "  ${YELLOW}$$_->[0]${RESET}$$sep${GREEN}$$_->[1]${RESET}\n"; \
	}; \
	print "\n"; } \
	$$sep = " " x (32 - length "help"); \
	print "${WHITE}Options${RESET}\n"; \
	print "  ${YELLOW}help${RESET}$$sep${GREEN}Prints this help${RESET}\n";

help:
	@echo "\nUsage: make ${YELLOW}<target>${RESET}\n\nThe following targets are available:\n";
	@perl -e '$(HELP_FUN)' $(MAKEFILE_LIST)

docker-up: ##@Local-development
	docker-compose -f deploy/docker-compose.yml up -d

docker-down: ##@Local-development
	docker-compose -f deploy/docker-compose.yml down

docker-logs: ##@Local-development
	docker-compose -f deploy/docker-compose.yml logs -f

migrate: ##@Migrations
	docker exec -it gotemplate /bin/bash -c 'migrate -path app/migrations -database=$$DATABASE_URL $(cmd)'

migrate-up: ##@Migrations
	docker exec -it gotemplate /bin/bash -c 'migrate -path app/migrations -database=$$DATABASE_URL up'

migrate-down: ##@Migrations
	docker exec -it gotemplate /bin/bash -c 'migrate -path app/migrations -database=$$DATABASE_URL down'

migrate-fresh: ##@Migrations
	make migrate-down
	make migrate-up

migration: ##@Migrations
	docker-compose -f deploy/docker-compose.yml exec gotemplate migrate create -ext sql -dir app/migrations $(name)

test: ##@Test
	docker-compose -f deploy/docker-compose.yml exec -e APP_ENV="test" gotemplate go test ./...
