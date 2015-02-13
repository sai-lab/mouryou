GO_BUILDOPT := -ldflags '-s -w'

run:
	go run main.go ${ARGS}

fmt:
	go fmt ./...

build: fmt
	go build $(GO_BUILDOPT) -o bin/tenbin main.go

clean:
	rm -f bin/tenbin

install: build
	cp bin/tenbin /usr/local/bin/

uninstall: clean
	rm -f /usr/local/bin/tenbin
