GO_BUILDOPT := -ldflags '-s -w'

run:
	go run main.go ${ARGS}

fmt:
	go fmt ./...

build: fmt
	go build $(GO_BUILDOPT) -o bin/mouryou main.go

clean:
	rm -f bin/mouryou

install: build
	cp bin/mouryou /usr/local/bin/

uninstall: clean
	rm -f /usr/local/bin/mouryou
