.DEFAULT_GOAL := help

PROJECT_BASE := github.com/skycoin/cx-tracker

install: ## Installs cxchain and cxchain-cli
	go install ./cmd/...

install-linters: ## Install code linters
	- VERSION=latest ./ci_scripts/install-golangci-lint.sh
	# GO111MODULE=off go get -u github.com/FiloSottile/vendorcheck
	# For some reason this install method is not recommended, see https://github.com/golangci/golangci-lint#install
	# However, they suggest `curl ... | bash` which we should not do
	# ${OPTS} go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/incu6us/goimports-reviser

dep: ## Tidy and update dependencies
	go mod tidy -v
	go mod vendor -v

format: ## Formats code
	goimports -w -local $(PROJECT_BASE) ./pkg
	goimports -w -local $(PROJECT_BASE) ./cmd
	find . -type f -name '*.go' -not -path "./vendor/*"  -exec goimports-reviser -project-name ${PROJECT_BASE} -file-path {} \;

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
