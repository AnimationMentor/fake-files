


all: install

full: generate lint vet test install



generate:
	go generate ./...


lint:
	golint ./...

test:
	go test ./...

install: generate depend
	go install ./...

vet:
	go tool vet .

depend:
	go get -d -v ./...
