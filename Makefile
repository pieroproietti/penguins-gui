.PHONY: build package appimage clean run test tidy

build:
	@test "$$(id -u)" != 0 || { echo "Build as a normal user, without sudo."; exit 1; }
	go build -o penguins-gui .

package: build
	go run ./cmd/package

appimage: build
	go run ./cmd/appimage

clean:
	rm -f penguins-gui
	rm -rf dist

run:
	go run .

test:
	go test ./...

tidy:
	go mod tidy
