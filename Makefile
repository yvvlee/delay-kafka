.PHONY: init
# init tool
init:
	go mod tidy
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/google/wire/cmd/wire@latest

# build
build:
	mkdir -p bin/
	go build -o ./bin/ ./...

.PHONY: wire
# generate DI code
wire:
	wire ./...

.PHONY: all
# generate all, lint and build
all:
	make wire
	make lint
	go build --race -o ./bin/ ./...
.PHONY: lint
# lint
lint:
	golangci-lint run -v

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
