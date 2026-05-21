.PHONY: build test clean install

build:
	go build -o cr-cli .

test:
	go test ./... -v

clean:
	rm -f cr-cli

install: build
	cp cr-cli /usr/local/bin/
