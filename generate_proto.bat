protoc -I ./proto/ --go-grpc_out=%GOPATH%/src --go_out=%GOPATH%/src proto/messaging.proto
