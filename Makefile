


all: install

full: generate lint vet test install



generate:
	go generate ./...


lint:
	golint ./...

test:
	go test ./...

install: generate depend
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go install ./...

vet:
	go vet ./...

depend:
	go get -d -v ./...
