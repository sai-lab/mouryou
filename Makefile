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

link:
	mkdir -p $(GOPATH)/src/github.com/sai-lab
	ln -s $(CURDIR) $(GOPATH)/src/github.com/sai-lab/mouryou

unlink:
	rm $(GOPATH)/src/github.com/sai-lab/mouryou
	rmdir $(GOPATH)/src/github.com/sai-lab
