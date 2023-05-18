export CGO_ENABLED=1
export GO111MODULE=on

proto_path="api/kpimon:${GOPATH}/pkg/src/github.com/gogo/protobuf/protobuf:${GOPATH}/pkg/src/github.com/gogo/protobuf:${GOPATH}/pkg/src"

.PHONY: gen client clean

gen:
	protoc --proto_path=${proto_path} \
		--go_out=:api \
		--go-grpc_out=:api \
		api/kpimon/*.proto

client:
	go run cmd/client/client.go

clean:
	rm api/onos.kpimon/*.go




