.PHONY: generate
generate:
	protoc -I proto proto/cloudv1/cloudv1.proto \
	--go_out=./pkg/ --go_opt=paths=source_relative \
	--go-grpc_out=./pkg/ --go-grpc_opt=paths=source_relative
