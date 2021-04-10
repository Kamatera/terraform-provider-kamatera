VERSION=0.0.4
LOCAL_PROVIDERS="$$HOME/.terraform.d/plugins_local"
BINARY_PATH="registry.terraform.io/kamatera/kamatera/${VERSION}/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-kamatera_v${VERSION}"

# Builds the provider and adds it to an independently configured filesystem_mirror folder.
build:
	@echo "Please configure your .terraformrc file to contain a filesystem_mirror block pointed at '${LOCAL_PROVIDERS}' for 'registry.terraform.io/kamatera/kamatera'"
	@echo "You MUST delete existing cached plugins from any .terraform directories in Terraform installations you want to test against so that it will perform a lookup on the local mirror"
	go build -o "${LOCAL_PROVIDERS}/${BINARY_PATH}"
