.PHONY: lint
## Runs linters on Go code files and YAML files
lint: lint-go-code lint-yaml

YAML_FILES := $(shell find . -type f -regex ".*y[a]ml" -print)
.PHONY: lint-yaml
## runs yamllint on all yaml files
lint-yaml: ${YAML_FILES}
	yamllint -c .yamllint $(YAML_FILES)

.PHONY: lint-go-code
## Checks the code with golangci-lint
lint-go-code:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	${GOPATH}/bin/golangci-lint ${V_FLAG} run