VERSION=0.0.4
LOCAL_PROVIDERS="$$HOME/.terraform.d/plugins_local"
BINARY_PATH="registry.terraform.io/kamatera/kamatera/${VERSION}/$$(go env GOOS)_$$(go env GOARCH)/terraform-provider-kamatera_v${VERSION}"

# Builds the provider and adds it to an independently configured filesystem_mirror folder.
build:
	@echo "Please configure your .terraformrc file to contain a filesystem_mirror block pointed at '${LOCAL_PROVIDERS}' for 'registry.terraform.io/kamatera/kamatera'"
	@echo "You MUST delete existing cached plugins from any .terraform directories in Terraform installations you want to test against so that it will perform a lookup on the local mirror"
	go build -o "${LOCAL_PROVIDERS}/${BINARY_PATH}"


# New script
TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=register.terraform.io
NAMESPACE=kamatera
NAME=kamatera
BINARY=terraform-provider-${NAME}
VERSION=0.5
OS_ARCH=darwin_amd64

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test -i $(TEST) || exit 1
	echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m
