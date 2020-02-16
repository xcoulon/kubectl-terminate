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
	$(Q)CGO_ENABLED=0 \
		go build ${V_FLAG} \
		-ldflags "-X ${GO_PACKAGE_PATH}/cmd.Commit=${GIT_COMMIT_ID} -X ${GO_PACKAGE_PATH}/BuildTime=${BUILD_TIME}" \
		-o $(OUT_DIR)/kubectl-terminate \
		main.go

.PHONY: install
## Builds and installs the operator in $GOPATH/bin
install: build
	@echo "installing..."
	$(Q) mv $(OUT_DIR)/kubectl-terminate $(GOPATH)/bin

