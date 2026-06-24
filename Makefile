.PHONY: run build test tidy fmt vet docker clean

BIN := bin/url-shortener

run:
	go run ./cmd/server

build:
	go build -o $(BIN) ./cmd/server

test:
	go test ./... -v

tidy:
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...

docker:
	docker build -t url-shortener .

clean:
	rm -rf bin urls.db urls.db-wal urls.db-shm
