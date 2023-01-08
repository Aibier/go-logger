GO		:= $(shell which go)
GOPATH		:= $(shell go env GOPATH)
GOBIN		:= $(GOPATH)/bin
GOLINT		:= $(GOBIN)/golint

# list go source directories here
COVERAGE_FILE	:= coverage.out
COVERAGE_HTML	:= coverage.html

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: clean
clean:
	rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)

.PHONY: test
test: clean lint vet unit

.PHONY: lint-install
lint-install:
	test -e $(GOLINT) || $(GO) get -u golang.org/x/lint/golint

.PHONY: lint
lint: lint-install
	$(GOLINT) -set_exit_status ./...

.PHONY: unit
unit:
	$(GO) test -race -v ./... -coverprofile=$(COVERAGE_FILE)

.PHONY: coverage
coverage: unit
	$(GO) tool cover -func=$(COVERAGE_FILE)

.PHONY: coverage-html
coverage-html: unit
	$(GO) tool cover -html=$(COVERAGE_FILE)

.PHONY: tidy
tidy:
	$(GO) mod tidy

