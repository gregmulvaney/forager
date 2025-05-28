run:
	go run cmd/forager/main.go

build-plugins:
	go build -buildmode=plugin -o tmp/plugins/example.so ./plugin-example/main.go
