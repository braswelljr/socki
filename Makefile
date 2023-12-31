HASGOCILINT := $(shell which golangci-lint 2> /dev/null)
ifdef HASGOCILINT
    GOLINT=golangci-lint
else
    GOLINT=bin/golangci-lint
endif

.PHONY: install
install:
	go install -v ./...

.PHONY: serve
serve:
	air -c .air.toml

.PHONY: run
run:
	go run main.go

.PHONY: download
download:
	go mod download

.PHONY: upgrade
upgrade:
	go get -u ./...

.PHONY: build
build:
	go build -o ./tmp/main ./.

.PHONY: run/build
run/build:
	@make build
	./tmp/main

.PHONY: test
test:
	go test -race ./...

.PHONY: fix
fix: ## Fix lint violations
	gofmt -s -w .
	goimports -w $$(find . -type f -name '*.go' -not -path "*/vendor/*")

.PHONY: lint
lint: ## Run linters
	$(GOLINT) run