.PHONY: test
## runs the tests without coverage
test:
	@echo "running the tests without coverage..."
	$(Q)go test ${V_FLAG} -race $(shell go list ./...) -failfast
	
# Output directory for coverage information
COV_DIR = $(OUT_DIR)/coverage

.PHONY: test-with-coverage
## runs the tests *with* coverage
test-with-coverage: 
	@echo "running the tests with coverage..."
	@-mkdir -p $(COV_DIR)
	@-rm $(COV_DIR)/coverage.txt
	$(Q)go test -vet off ${V_FLAG} $(shell go list ./...) -coverprofile=$(COV_DIR)/coverage.txt -covermode=atomic ./...