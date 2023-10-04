CGO_CPPFLAGS ?= ${CPPFLAGS}
export CGO_CPPFLAGS
CGO_CFLAGS ?= ${CFLAGS}
export CGO_CFLAGS
CGO_LDFLAGS ?= $(filter -g -L% -l% -O%,${LDFLAGS})
export CGO_LDFLAGS

EXE =
ifeq ($(GOOS),windows)
EXE = .exe
endif

## The following tasks delegate to `script/build.go` so they can be run cross-platform.

.PHONY: bin/inst$(EXE)
bin/inst$(EXE): script/build
	@script/build $@

script/build: script/build.go
	GOOS= GOARCH= GOARM= GOFLAGS= CGO_ENABLED= go build -o $@ $<

.PHONY: clean
clean:
	rm -f script/build
	rm -f bin/inst

# just a convenience task around `go test`
.PHONY: test
test:
	@go test -v -race -coverpkg=./... -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out
	@go tool cover -html=coverage.out
	@rm coverage.out
## Install/uninstall tasks are here for use on *nix platform. On Windows, there is no equivalent.
DESTDIR :=
prefix  := /usr/local
bindir  := ${prefix}/bin
mandir  := ${prefix}/share/man

.PHONY: install
install: bin/inst
	install -d ${DESTDIR}${bindir}
	install -m755 bin/inst ${DESTDIR}${bindir}/

.PHONY: uninstall
uninstall:
	rm -f ${DESTDIR}${bindir}/inst
