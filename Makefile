BUILD = fountain-finder

run:
	@type air > /dev/null 2>&1 && air || go run main.go

build:
	go build -o ${BUILD} main.go

lint:
	golangci-lint run ./...
