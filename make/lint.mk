.PHONY: lint
## Runs linters on Go code files and YAML files
lint:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint
	${GOPATH}/bin/golangci-lint ${V_FLAG} run