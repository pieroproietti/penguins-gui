.PHONY: build package clean run test tidy

build:
	@test "$$(id -u)" != 0 || { echo "Build as a normal user, without sudo."; exit 1; }
	go build -o penguins-gui .

package: build
	go run ./cmd/package

clean:
	rm -f penguins-gui
	rm -rf dist

run:
	go run .

test:
	go test ./...

tidy:
	go mod tidy
