.PHONY: all
all: build

.PHONY: clean
clean:
	rm -rf bin

.PHONY:test
test:
	go test -v server/*.go

.PHONY: build
build: test
	mkdir -p ./bin
	export CGO_ENABLED=0 && go build -o ./bin/lobbyd daemon/*.go
	export CGO_ENABLED=0 && go build -o ./bin/lobbyctl ctl/*.go
	
