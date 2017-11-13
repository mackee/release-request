VERSION := $(shell git describe --tags)

_bin/release-request: *.go
	go generate
	go build -o _bin/release-request -ldflags="-X main.Version=$(VERSION)"

.PHONY: clean install get-deps test build

test:
	go test -v -race
	go vet

get-deps:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

clean:
	rm -Rf _bin/* _artifacts/*

install: _bin/release-request
	install _bin/release-request $(GOPATH)/bin

build: clean get-deps test
	go generate
	gox -output "_artifacts/{{.Dir}}-{{.OS}}-{{.Arch}}-${VERSION}/release-request" -ldflags "-w -s -X main.Version=$(VERSION)"
	cd _artifacts/ && find . -name 'release-request*' -type d | sed 's/\.\///' | xargs -I{} zip -m -q -r {}.zip {}

release:
	ghr ${VERSION} _artifacts
