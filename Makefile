BINARY_NAME := check_nextcloud
BUILD_DIR   := build
INSTALL_DIR := /usr/local/lib/nagios/plugins

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	install -Dm755 $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
