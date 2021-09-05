VERSION=1.0


.PHONY: all
all: build

.PHONY: clean
clean:
	rm -rf bin

.PHONY:test
test:
	go test -v server/*.go

.PHONY: build
build: test clean
	mkdir -p ./bin
	# linux amd64
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/lobbyd-${VERSION}-linux-amd64 daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/lobbyctl-${VERSION}-linux-amd64 ctl/*.go

	# linux arm
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./bin/lobbyd-${VERSION}-linux-arm daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -o ./bin/lobbyctl-${VERSION}-linux-arm ctl/*.go
	
	# linux arm64
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/lobbyd-${VERSION}-linux-arm64 daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/lobbyctl-${VERSION}-linux-arm64 ctl/*.go

	# darwin amd64
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/lobbyd-${VERSION}-darwin-amd64 daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/lobbyctl-${VERSION}-darwin-amd64 ctl/*.go

	# darwin arm64
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/lobbyd-${VERSION}-darwin-arm64 daemon/*.go
	env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o ./bin/lobbyctl-${VERSION}-darwin-arm64 ctl/*.go
