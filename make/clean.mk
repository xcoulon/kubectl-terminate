.PHONY: clean
clean:
	$(Q)-rm -rf ${V_FLAG} $(OUT_DIR) ./vendor
	$(Q)go clean ${X_FLAG} ./...
	$(Q)-rm deploy/templates/nstemplatetiers/metadata.yaml
	$(Q)-rm pkg/templates/nstemplatetiers/nstemplatetier_assets.go 2>/dev/null
	$(Q)-rm test/templates/nstemplatetiers/nstemplatetier_assets.go 2>/dev/null
