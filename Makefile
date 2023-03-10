export CGO_ENABLED ?= 0
GOFLAGS += -trimpath
LDFLAGS += -X main.version=$(VERSION)
INSTALL ?= install
INSTALL_PROGRAM ?= $(INSTALL)

prefix = /usr/local
bindir ?= $(prefix)/bin

builddir = bin
distdir = dist
tmpdir = tmp

all: test build

build: gen
	@mkdir -p "$(builddir)"
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o "$(builddir)/" ./...

check: gen
	golangci-lint run ./...

clean:
	find . -type f -name '*_gen.go' -delete
	rm -rf "$(builddir)" "$(distdir)" "$(tmpdir)"

dist:
	$(MAKE) bindir="$(distdir)/$(notdir $(CURDIR))" install
	tar -C $(distdir) -cvf "$(distdir)/$(notdir $(CURDIR)).tar.gz" "$(notdir $(CURDIR))"

gen: server/wsdl_gen.go

install: all
	$(INSTALL_PROGRAM) -Dt "$(DESTDIR)$(bindir)" "$(builddir)"/*

install-strip:
	$(MAKE) INSTALL_PROGRAM='$(INSTALL_PROGRAM) -s' install

test: gen
	@mkdir -p "$(tmpdir)/reports"
	go test $(GOFLAGS) -ldflags "$(LDFLAGS)" -coverprofile "$(tmpdir)/reports/coverage.out" ./...
	go tool cover -html "$(tmpdir)/reports/coverage.out" -o "$(tmpdir)/reports/coverage.html"

uninstall:
	rm -fv "$(bindir)/$(notdir $(CURDIR))"

server/wsdl_gen.go:
	go generate ./...

.PHONY: all build check clean dist gen install install-strip test uninstall
