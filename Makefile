.PHONY: proto build test vet

proto:
	protoc --proto_path=proto \
		--go_out=pkg/pb --go_opt=paths=source_relative \
		--go-grpc_out=pkg/pb --go-grpc_opt=paths=source_relative \
		shipment/v1/shipment.proto

build:
	go build ./...

test:
	go test ./internal/domain/shipment/ ./internal/service/shipment/ -v

vet:
	go vet ./...
