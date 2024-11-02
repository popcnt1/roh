PROJECT_NAME := roh
BIN_DIR := bin
CMD_DIR := cmd
BUILD_FLAGS := -ldflags "-s -w"

all: $(BIN_DIR)/getda

$(BIN_DIR)/getda: $(CMD_DIR)/getda/main.go
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/getda $(CMD_DIR)/getda/main.go

clean:
	rm -rf $(BIN_DIR)

TOOLS := getda

.PHONY: all clean $(TOOLS)

$(TOOLS):
	@mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/$@ $(CMD_DIR)/$@/main.go

