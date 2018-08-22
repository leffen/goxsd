
include .env


APP=goxsd


SOURCES= generate.go goxsd.go xsd.go 


VERSION=$(shell git describe --always --long)
BuildTime=$(shell date +%Y-%m-%d.%H:%M:%S)

bump:
	bump_version patch goxsd.go 

build: test
	go build -ldflags  "-s -X main.CommitHash=$(VERSION) -X main.BuildTime=$(BuildTime) -w"  -o bin/$(APP) .

run: 
	go run  $(SOURCES)  mediainfo.xsd 

test: vet
	@# this target should always be listed first so "make" runs the tests.
	go test -cover -race -v ./...

vet: 
	@# We can't vet the vendor directory, it fails.
	go list ./... | grep -v vendor | xargs go vet
		