BUILD = fountain-finder

run:
	air

build:
	go build -o ${BUILD} main.go

lint:
	golangci-lint run ./...
