# Raccourcis de développement.
# Les variables d'environnement sont chargées depuis le fichier .env.

include .env
export

.DEFAULT_GOAL := help

.PHONY: help
help: ## Affiche la liste des commandes disponibles
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: db-up
db-up: ## Démarre la base de données PostgreSQL
	docker compose up -d

.PHONY: db-down
db-down: ## Arrête la base de données
	docker compose down

.PHONY: db-reset
db-reset: ## Supprime et recrée la base (rejoue les migrations et les données de démo)
	docker compose down -v
	docker compose up -d

.PHONY: run
run: ## Lance le serveur API
	go run ./cmd/api

.PHONY: build
build: ## Compile le serveur dans bin/
	go build -o bin/api ./cmd/api

.PHONY: test
test: ## Lance l'ensemble des tests
	go test ./... -v

.PHONY: test-coverage
test-coverage: ## Lance les tests et affiche le taux de couverture
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out

.PHONY: fmt
fmt: ## Formate le code source
	go fmt ./...

.PHONY: vet
vet: ## Analyse statique du code
	go vet ./...

.PHONY: test-integration
test-integration: ## Lance les tests d'intégration (nécessite la base démarrée)
	TEST_DATABASE_URL="$(DATABASE_URL)" go test ./internal/repository -v
