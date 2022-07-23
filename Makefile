.PHONY: help pkgsite report test

help: ## list available targets
	@# Shamelessly stolen from Gomega's Makefile
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

pkgsite: ## serves Go documentation on port 6060
	@echo "navigate to: http://localhost:6060/github.com/thediveo/whalewatcher"
	@scripts/pkgsite.sh

report: ## run goreportcard on this module
	@scripts/goreportcard.sh

test: ## run unit tests
	go test -v -p=1 -race ./...
