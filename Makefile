ifneq ($(GOPATH),)
  prefix ?= $(GOPATH)
endif
prefix ?= /usr/local
exec_prefix ?= $(prefix)
ifneq ($(GOBIN),)
  bindir ?= $(GOBIN)
endif
bindir ?= $(exec_prefix)/bin

treehash:=$(shell env CODECHAIN_DIR=.codechain_mainnet CODECHAIN_EXCLUDE=.codechain_testnet codechain treehash)

.PHONY: all install uninstall clean test update-vendor

all:
	env GO111MODULE=on go build -mod vendor -v -ldflags "-X main.treehash=$(treehash)" . ./cmd/...

install:
	env GO111MODULE=on GOBIN=$(bindir) go install -mod vendor -v -ldflags "-X main.treehash=$(treehash)" . ./cmd/...

uninstall:
	rm -f $(bindir)/addblock $(bindir)/gencerts $(bindir)/promptsecret $(bindir)/fnod $(bindir)/addr2pkscript $(bindir)/gennonce $(bindir)/fnochain $(bindir)/fnoupdate $(bindir)/findcheckpoint $(bindir)/printunixtime $(bindir)/fnoctl

clean:
	rm -f fnod

test:
	env GO111MODULE=on ./run_tests.sh

update-vendor:
	rm -rf vendor
	env GO111MODULE=on go get -u
	env GO111MODULE=on go mod tidy -v
	env GO111MODULE=on go mod vendor
