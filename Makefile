
VERSION=v0.0.1
ARCH=amd64
OS=linux

TARGET_BINARY=terraform-provider-circleci_$(VERSION)

TERRAFORM_PLUGIN_DIR=$(HOME)/.terraform.d/plugins/$(OS)_$(ARCH)/

.PHONY: $(TARGET_BINARY)

build: $(TARGET_BINARY)

$(TARGET_BINARY):
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags="-s -w" -a -o $(TARGET_BINARY)

test:
	go test -v -cover ./...

install_plugin_locally: $(TARGET_BINARY)
	mkdir -p $(TERRAFORM_PLUGIN_DIR)
	cp ./$(TARGET_BINARY) $(TERRAFORM_PLUGIN_DIR)/