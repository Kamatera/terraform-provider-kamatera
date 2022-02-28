VERSION=0.7.4
LOCAL_PROVIDERS="$$HOME/.terraform.d/plugins"
BINARY_PATH="registry.terraform.io/kamatera/kamatera/${VERSION}/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-kamatera_v${VERSION}"

# Builds the provider and adds it to an independently configured filesystem_mirror folder.
build:
	go build -o "${LOCAL_PROVIDERS}/${BINARY_PATH}"
