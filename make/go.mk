# By default the project should be build under GOPATH/src/github.com/<orgname>/<reponame>
GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}

GO111MODULE?=on
export GO111MODULE

.PHONY: build 
## Build the operator
build: 
	@echo "building..."
	@-CGO_ENABLED=0 \
		go build \
		-o $(OUT_DIR)/kubectl-terminate \
		cmd/main.go

.PHONY: install
## Builds and installs the operator in $GOPATH/bin
install: build
	@echo "installing..."
	 mv $(OUT_DIR)/kubectl-terminate $(GOPATH)/bin

