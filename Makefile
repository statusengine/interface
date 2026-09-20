# Statusengine Web Interface

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X github.com/statusengine/interface/internal/httpapi.Version=$(VERSION)

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: frontend ## Build the binary with the frontend embedded
	go build -trimpath -ldflags "$(LDFLAGS)" -o seid ./cmd/seid

.PHONY: build-api
build-api: ## Build the binary without a frontend (API only, fast)
	go build -ldflags "$(LDFLAGS)" -o seid ./cmd/seid

.PHONY: frontend
frontend: ## Build the Angular bundle into internal/webui/dist
	cd frontend && npm run build

.PHONY: deps
deps: ## Install frontend dependencies
	cd frontend && npm ci

.PHONY: dev-api
dev-api: ## Run the API on :8090 (pair with `make dev-ui`)
	go run ./cmd/seid serve -config seid.yaml

.PHONY: dev-ui
dev-ui: ## Run the Angular dev server on :4200, proxying /api to :8090
	cd frontend && npm start

.PHONY: test
test: test-go test-ui ## Run every test

.PHONY: test-go
test-go: ## Run the Go tests
	go test ./...

.PHONY: test-integration
test-integration: ## Run the repository tests against a real Statusengine schema (read-only)
	@test -n "$(SEI_TEST_DSN)" || { \
		echo "SEI_TEST_DSN is not set. Example:"; \
		echo "  make test-integration SEI_TEST_DSN='user:pass@tcp(127.0.0.1:3306)/statusengine'"; \
		exit 1; }
	SEI_TEST_DSN="$(SEI_TEST_DSN)" go test -count=1 ./internal/repository/...

.PHONY: test-ui
test-ui: ## Run the frontend tests
	cd frontend && npx ng test --watch=false

.PHONY: test-a11y
test-a11y: ## Audit accessibility against a running instance (see tools/a11y/README.md)
	@test -n "$(SEI_PASS)" || { \
		echo "SEI_PASS is not set. Example:"; \
		echo "  make test-a11y SEI_URL=http://127.0.0.1:8090 SEI_USER=ops SEI_PASS=secret"; \
		exit 1; }
	cd tools/a11y && npm install --silent && \
		SEI_URL="$(SEI_URL)" SEI_USER="$(SEI_USER)" SEI_PASS="$(SEI_PASS)" node audit.mjs && \
		SEI_URL="$(SEI_URL)" SEI_USER="$(SEI_USER)" SEI_PASS="$(SEI_PASS)" node dialogs.mjs && \
		SEI_URL="$(SEI_URL)" SEI_USER="$(SEI_USER)" SEI_PASS="$(SEI_PASS)" node keyboard.mjs

.PHONY: lint
lint: ## Vet the Go code and check frontend formatting
	go vet ./...
	gofmt -l cmd internal
	cd frontend && npm run lint:format

.PHONY: format
format: ## Format everything
	gofmt -w cmd internal
	cd frontend && npm run format

.PHONY: migrate
migrate: ## Apply the sei_* schema and exit
	go run ./cmd/seid migrate -config seid.yaml

.PHONY: docker
docker: ## Build the container image
	docker build --build-arg VERSION=$(VERSION) -t statusengine/interface:$(VERSION) .

.PHONY: clean
clean: ## Remove build output
	rm -f seid
	find internal/webui/dist -mindepth 1 ! -name .gitkeep -delete
	rm -rf frontend/dist frontend/.angular
