PKG = $(shell cat go.mod | grep "^module " | sed -e "s/module //g")
VERSION = v$(shell cat .version)
COMMIT_SHA ?= $(shell git describe --always)-devel
GOTEST = go test -v -race -failfast

test: download
	GODEBUG=x509ignoreCN=0 $(GOTEST) ./...

cover:
	$(GOTEST) -coverprofile=coverage.txt -covermode=atomic ./...

fmt:
	goimports -l -w .
	gofmt -l -w .

download:
	go mod download