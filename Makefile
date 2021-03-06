#
# Simple Makefile
#
PROG = xlquery

build: fmt
	go build
	go build ./rss2
	go build -o bin/$(PROG) cmds/$(PROG)/$(PROG).go

test: fmt
	go test
	go test ./rss2

fmt: 
	gofmt -w $(PROG).go
	gofmt -w $(PROG)_test.go
	gofmt -w rss2/rss2.go
	gofmt -w rss2/rss2_test.go
	gofmt -w cmds/$(PROG)/$(PROG).go
	goimports -w $(PROG).go
	goimports -w $(PROG)_test.go
	goimports -w rss2/rss2.go
	goimports -w rss2/rss2_test.go
	goimports -w cmds/$(PROG)/$(PROG).go

save:
	./mk-website.bash
	git commit -am "quick save"
	git push origin master

clean:
	if [ -d bin ]; then rm -fR bin; fi
	if [ -d dist ]; then rm -fR dist; fi
	if [ -f $(PROG)-binary-release.zip ]; then rm -f $(PROG)-binary-release.zip; fi

install:
	env GOBIN=$(HOME)/bin go install cmds/$(PROG)/$(PROG).go

release:
	./mk-release.bash

publish:
	./mk-website.bash
	./publish.bash
