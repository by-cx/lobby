VERSION=1.1


.PHONY: all
all: build

.PHONY: clean
clean:
	rm -rf bin

.PHONY:test
test:
	go test -v server/*.go

init:
	mkdir -p ./bin

.PHONY: build
build: test clean init linux-amd64 linux-arm linux-arm64 darwin-amd64 darwin-arm64
	

.PHONY: linux-amd64
linux-amd64: clean init
	# linux amd64
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/lobbyd-${VERSION}-linux-amd64 daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/lobbyctl-${VERSION}-linux-amd64 ctl/*.go

.PHONY: linux-arm
linux-arm: clean init
	# linux arm
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./bin/lobbyd-${VERSION}-linux-arm daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./bin/lobbyctl-${VERSION}-linux-arm ctl/*.go
	
.PHONY: linux-arm64
linux-arm64: clean init
	# linux arm64
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/lobbyd-${VERSION}-linux-arm64 daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/lobbyctl-${VERSION}-linux-arm64 ctl/*.go

.PHONY: darwin-amd64
darwin-amd64: clean init
	# darwin amd64
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/lobbyd-${VERSION}-darwin-amd64 daemon/*.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ./bin/lobbyctl-${VERSION}-darwin-amd64 ctl/*.go

.PHONY: darwin-arm64
darwin-arm64: clean init
	# darwin arm64
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/lobbyd-${VERSION}-darwin-arm64 daemon/*.go
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o ./bin/lobbyctl-${VERSION}-darwin-arm64 ctl/*.go
