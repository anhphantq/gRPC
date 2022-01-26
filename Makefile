clean:
	del /s pb\*.go

server:
	go run ./cmd/server/main.go -port 9080

client:
	go run ./cmd/client/main.go -address 0.0.0.0:9080

test:
	go test -cover -race ./...