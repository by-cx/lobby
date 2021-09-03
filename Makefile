.PHONY: all
all: build

.PHONY: clean
clean:
	rm -rf bin

.PHONY: build
build:
	mkdir -p ./bin
	export CGO_ENABLED=0
	go build -o ./bin/lobbyd daemon/*.go
	
