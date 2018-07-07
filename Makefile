
default:	test

fmt:
	go fmt ./...

test:	build/gopgvenv
	./$<

build/gopgvenv:	cmd/gopgvenv/main.go
	go build -o $@ ./...

clean:
	rm -rf ./build

.PHONY:	clean default fmt test
