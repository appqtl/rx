GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOTIDY=$(GOMOD) tidy
GOTOOL=$(GOCMD) tool

.PHONY: test
test:
	$(GOTEST) -coverprofile cover.out -v ./...

.PHONY: tidy
tidy:
	$(GOTIDY)

.PHONY: cover
cover: test
	$(GOTOOL) cover -html=cover.out