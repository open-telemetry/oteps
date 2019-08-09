# All documents to be used in spell check.
ALL_DOCS := $(shell find . -name '*.md' -type f | sort)

TOOLS_DIR := ./.tools

$(TOOLS_DIR)/misspell: go.mod go.sum internal/tools.go
	go build -o $(TOOLS_DIR)/misspell github.com/client9/misspell/cmd/misspell

.PHONY: precommit
precommit: $(TOOLS_DIR)/misspell misspell

.PHONY: misspell
misspell:
	$(TOOLS_DIR)/misspell -error $(ALL_DOCS)

.PHONY: misspell-correction
misspell-correction:
	$(TOOLS_DIR)/misspell -w $(ALL_DOCS)

