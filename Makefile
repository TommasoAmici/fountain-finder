run:
	@type air > /dev/null 2>&1 && air || go run main.go

build: main.go cli/seed_search.go
	go build -o build/seed_search cli/seed_search.go
	go build -o build/fountain_finder main.go

clean:
	rm -r build

lint:
	golangci-lint run ./...
