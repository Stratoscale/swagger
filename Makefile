all: clean testdata test

image = go-swagger:strato

swagger = docker run --rm \
	-e GOPATH=$(GOPATH):/go \
	-v $(PWD):$(PWD) \
	-w $(PWD)/testdata \
	-u $(shell id -u):$(shell id -g) \
	$(image)

build:
	docker build . -t $(image)

test:
	cd testdata && go build ./main.go
	cd testdata && go test ./...

testdata: build clean
	cd testdata ; \
		$(swagger) generate server ; \
		$(swagger) generate client ; \
		go generate ./...

clean:
	rm -rf testdata/restapi testdata/models testdata/client
