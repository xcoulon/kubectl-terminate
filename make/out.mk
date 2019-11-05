# Create output directory for artifacts and test results. `./build/_output` is supposed to
# be a safe place for all targets to write to while knowing that all content
# inside of `./build/_output` is wiped once "make clean" is run.
OUT_DIR := ./bin
$(shell mkdir -p $(OUT_DIR))