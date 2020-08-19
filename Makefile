snapshot:
	goreleaser build --snapshot --rm-dist
build:
	goreleaser build --rm-dist
release:
	goreleaser release --rm-dist
dev:
	air

init:
	go get -u golang.org/x/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	@echo "Install pre-commit hook"
	@rm -f $(shell pwd)/.git/hooks/pre-commit ||true
	@ln -s $(shell pwd)/pre-commit $(shell pwd)/.git/hooks/pre-commit || true
	@chmod +x ./check.sh

getdeps:
	@mkdir -p ${GOPATH}/bin
	@which golangci-lint 1>/dev/null || (echo "Installing golangci-lint" && curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh |sudo sh -s -- -b ${GOPATH}/bin v1.27.0)

lint:
	@echo "Running $@ check"
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint cache clean
	@GO111MODULE=on ${GOPATH}/bin/golangci-lint run --timeout=5m --config ./.golangci.yml
