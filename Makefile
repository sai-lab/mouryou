GO_BUILDOPT := -ldflags '-s -w'

dep:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

run:
	go run main.go ${ARGS}

imports:
	go get -u golang.org/x/tools/cmd/goimports

fmt:
	goimports -w *.go lib/*/*.go

build: fmt
	go build $(GO_BUILDOPT) -o bin/mouryou main.go

clean:
	rm -f bin/mouryou

link:
	mkdir -p $(GOPATH)/src/github.com/sai-lab
	ln -s $(CURDIR) $(GOPATH)/src/github.com/sai-lab/mouryou

unlink:
	rm $(GOPATH)/src/github.com/sai-lab/mouryou
	rmdir $(GOPATH)/src/github.com/sai-lab
