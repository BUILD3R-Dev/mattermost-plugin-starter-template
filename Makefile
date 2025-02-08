# Makefile for packaging the Mattermost Ticket Plugin
# Determine the plugin version using Git tags (default to 0.0.0 if unavailable)
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "0.0.0")

# Plugin binary name
BINARY := mattermost-plugin-ticketing

.PHONY: build package clean

# Build target – change directory to "server" where the Go files reside,
# then compile the binary. The output is placed in the parent directory.
build:
	@echo "Building $(BINARY)..."
	cd server && go build -o ../$(BINARY) .

# Package target – builds the binary first then archives the binary, plugin manifest, and necessary assets.
package: build
	@echo "Packaging $(BINARY) version $(VERSION)..."
	tar -czvf $(BINARY)-$(VERSION).tar.gz $(BINARY) plugin.json server/email/template.html server/configuration.go server/plugin.go server/email/service.go server/tickets/utils.go server/email/smtp.go
	@echo "Package created: $(BINARY)-$(VERSION).tar.gz"

# Clean target – remove build artifacts.
clean:
	@echo "Cleaning..."
	rm -f $(BINARY) $(BINARY)-*.tar.gz