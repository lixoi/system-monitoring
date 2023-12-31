gen:
	rm -f ./internal/server/grpc/api/api.pb.go
	rm -f ./internal/server/grpc/api/api_grpc.pb.go
	protoc --go-grpc_out=./internal/server/grpc/api api/*.proto
	protoc --go_out=./internal/server/grpc/api api/*.proto

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.50.1

lint: install-lint-deps
	golangci-lint run ./...

build_linux: gen
	env GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o sysstatssvc main.go

build_windows: gen
	env GOOS=windows GOARCH=amd64 go build -a -installsuffix cgo -o sysstatssvc main.go
