BIN_DIR := bin
CMD_DIR := cmd
BUILD_FLAGS := -ldflags "-s -w"

# Define all tools
TOOLS := getda getreorg idxblk

# The default target builds all tools.  It now depends on the binaries
# in the bin directory.
all: $(addprefix $(BIN_DIR)/, $(TOOLS))

# Phony targets for all, clean, and each tool
.PHONY: all clean install $(TOOLS)


# Build rule for each tool.  This uses a pattern rule to simplify
# the creation of multiple similar targets.
$(BIN_DIR)/%: $(CMD_DIR)/%/main.go
	@mkdir -p $(BIN_DIR) # Ensure bin/ exists
	go build $(BUILD_FLAGS) -o $@ $<

clean:
	rm -rf $(BIN_DIR)

install: all
		@mkdir -p $(GOPATH)/bin # Ensure $GOPATH/bin exists
		@for tool in $(TOOLS); do \
			cp $(BIN_DIR)/$$tool $(GOPATH)/bin/; \
			echo "Installed $$tool to $(GOPATH)/bin"; \
		done
