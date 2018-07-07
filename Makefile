
default:	build/gopgvenv

fmt:
	go fmt ./...

build/gopgvenv:	cmd/gopgvenv/main.go
	go build -o $@ ./...

clean:
	rm -rf ./build

.PHONY:	clean default fmt
