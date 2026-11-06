# Local dev clusters. All of them bind port 9092, so only one runs at a time:
# every up-* target stops the others first.

COMPOSE       := docker compose
DEFAULT_STACK := $(COMPOSE) -p ktea-dev -f docker/docker-compose.yml
SSL_STACK     := $(COMPOSE) -p ktea-ssl -f docker/docker-compose-ssl.yml
SCRAM_STACK   := $(COMPOSE) -p ktea-scram -f docker/docker-compose-scram.yml

.DEFAULT_GOAL := help
.PHONY: help up up-ssl up-scram down clean seed logs generate run run-debug

help: ## List the goals
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F ':.*## ' '{printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'

up: ## Start the default cluster: broker, schema registry, Kafka Connect
	@$(SSL_STACK) down
	@$(SCRAM_STACK) down
	$(DEFAULT_STACK) up -d --wait

up-ssl: ## Start the SSL / SASL_SSL cluster (ports 9092, 9093, 9094)
	@$(DEFAULT_STACK) down
	@$(SCRAM_STACK) down
	$(SSL_STACK) up -d --wait

up-scram: ## Start Redpanda with SASL/SCRAM (admin / adminadmin)
	@$(DEFAULT_STACK) down
	@$(SSL_STACK) down
	$(SCRAM_STACK) up -d --wait

down: ## Stop all clusters
	$(DEFAULT_STACK) down
	$(SSL_STACK) down
	$(SCRAM_STACK) down

clean: ## Stop all clusters and delete their data
	$(DEFAULT_STACK) down -v
	$(SSL_STACK) down -v
	$(SCRAM_STACK) down -v

seed: ## Create topics, consumer groups and offsets on the default cluster
	$(DEFAULT_STACK) --profile kafka-init run --rm kafka-init

logs: ## Follow the logs of the running cluster
	@stack=$$($(COMPOSE) ls -q | grep -E '^ktea-(dev|ssl|scram)$$' | head -1); \
	if [ -z "$$stack" ]; then echo "no cluster running"; exit 1; fi; \
	$(COMPOSE) -p $$stack logs -f

generate: ## Produce test records to the running cluster
	go run ./cmd/generate

run: ## Run ktea with the dev tag (simulated slow network)
	go run -tags dev ./cmd/ktea

run-debug: ## Run ktea with the dev tag and --debug (writes debug.log)
	go run -tags dev ./cmd/ktea --debug
