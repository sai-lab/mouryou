GO_BUILDOPT := -ldflags '-s -w'

gom:
	go get github.com/mattn/gom
	gom install

run:
	gom run main.go ${ARGS}

fmt:
	gom exec goimports -w *.go lib/*/*.go

build: fmt
	gom build $(GO_BUILDOPT) -o bin/mouryou main.go

clean:
	rm -f bin/mouryou

install: build
	cp bin/mouryou /usr/local/bin/

uninstall: clean
	rm -f /usr/local/bin/mouryou
