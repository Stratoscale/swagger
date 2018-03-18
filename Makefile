image = go-swagger:strato

id = $(shell id -u):$(shell id -g)

example_wd = $(PWD)/example
swagger = docker run --rm -it \
	-e GOPATH=$(GOPATH):/go \
	-v $(HOME):$(HOME) \
	-w $(example_wd) \
	-u $(id) \
	$(image)

build:
	docker build . -t $(image)

test:
	go test ./...

example: build clean
	cd example && \
		$(swagger) generate server && \
		$(swagger) generate client && \
		go generate ./...

clean:
	rm -rf example/restapi example/models example/client
