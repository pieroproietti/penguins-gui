.PHONY: build run test tidy

build:
	go build -o penguins-gui .

run:
	go run .

test:
	go test ./...

tidy:
	go mod tidy
