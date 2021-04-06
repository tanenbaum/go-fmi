.PHONY: build-examples
build-examples:
	mkdir -p out/build
	go build -buildmode c-shared -o ./out/build ./examples/...

.PHONY: test
test:
	go test -tags=none -race -cover ./...