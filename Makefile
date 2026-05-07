BINARY ?= memdb

BINARIES := client memdb

# Output directory
BIN_DIR := bin

CMD_DIR = ./cmd/$(BINARY)

# Show variables (debugging)
.PHONY: .vars
.vars:
	@echo "BINARY=$(BINARY)"
	@echo "CMD_DIR=$(CMD_DIR)"
	@echo "BIN_DIR=$(BIN_DIR)"
	@echo "BINARIES=$(BINARIES)"

# Build binary into bin/
.PHONY: .build
.build:
	@echo ">> building $(BINARY)"
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)


# Default target
.PHONY: build
build: clean
	@for bin in $(BINARIES); do \
    		$(MAKE) .build BINARY=$$bin; \
    done


# Clean build artifacts
.PHONY: clean
clean:
	@echo ">> cleaning"
	@rm -rf $(BIN_DIR)

.PHONY: test
test:
	go test -race -count=1 -v ./...

.PHONY: .run-memdb
.run-memdb:
	go run $(CMD_DIR) -c $(PWD)/memdb.conf.yml

.PHONY: run-memdb
run-memdb:
	$(MAKE) .run-memdb BINARY=memdb

.PHONY: .run-client
.run-client:
	go run $(CMD_DIR)

.PHONY: run-client
run-client:
	$(MAKE) .run-client BINARY=client
