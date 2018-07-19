all: clean example test

image = go-swagger:strato

swagger = docker run --rm \
	-e GOPATH=$(GOPATH):/go \
	-v $(PWD):$(PWD) \
	-w $(PWD)/example \
	-u $(shell id -u):$(shell id -g) \
	$(image)

build:
	docker build . -t $(image)

deps:
	go get ./...

test:
	go build ./example/main.go
	go test ./...

example: build clean
	cd example ; \
		$(swagger) generate server ; \
		$(swagger) generate client ; \
		go generate ./...

clean:
	rm -rf example/restapi example/models example/client
