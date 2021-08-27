# By default the project should be build under GOPATH/src/github.com/<orgname>/<reponame>
GO_PACKAGE_ORG_NAME ?= $(shell basename $$(dirname $$PWD))
GO_PACKAGE_REPO_NAME ?= $(shell basename $$PWD)
GO_PACKAGE_PATH ?= github.com/${GO_PACKAGE_ORG_NAME}/${GO_PACKAGE_REPO_NAME}

GO111MODULE?=on
export GO111MODULE

CUR_DIR=$(shell pwd)
INSTALL_PREFIX=$(CUR_DIR)/bin

ifeq ($(OS),Windows_NT)
BINARY_PATH=$(INSTALL_PREFIX)/kubectl-terminate.exe
else
BINARY_PATH=$(INSTALL_PREFIX)/kubectl-terminate
endif

.PHONY: build 
## Build the operator
build: 
	$(eval BUILD_COMMIT:=$(shell git rev-parse --short HEAD))
	$(eval BUILD_TAG:=$(shell git tag --contains $(BUILD_COMMIT)))
	$(eval BUILD_TIME:=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ'))
	@echo "building with commit:$(BUILD_COMMIT) / tag:$(BUILD_TAG) / time:$(BUILD_TIME)"
	@CGO_ENABLED=0 \
		go build -ldflags \
		"-X github.com/xcoulon/kubectl-terminate/BuildCommit=$(BUILD_COMMIT) \
	    -X github.com/xcoulon/kubectl-terminate/main.BuildTag=$(BUILD_TAG) \
	    -X github.com/xcoulon/kubectl-terminate/main.BuildTime=$(BUILD_TIME)" \
		-o $(BINARY_PATH) \
		cmd/main.go
	@echo "$(BINARY_PATH) is ready to use"

.PHONY: install
## Builds and installs the operator in $GOPATH/bin
install: build
	@echo "installing..."
	 mv $(BINARY_PATH) $(GOPATH)/bin

