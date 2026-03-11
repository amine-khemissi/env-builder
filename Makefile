BINARY_NAME := eb
BUILD_OUT   := ./bin/$(BINARY_NAME)
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build install clean

build:
	go build -o $(BUILD_OUT) .

install: build
	mkdir -p $(INSTALL_DIR)
	cp $(BUILD_OUT) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed to $(INSTALL_DIR)/$(BINARY_NAME)"

clean:
	rm -f $(BUILD_OUT)
