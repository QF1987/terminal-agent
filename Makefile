.PHONY: build run serve test clean proto

# 编译
build:
	go build -o bin/device-ctl ./cmd/device-ctl

# 启动 gRPC 服务（默认端口 9090）
serve: build
	./bin/device-ctl serve

# 指定端口启动
serve-port: build
	./bin/device-ctl serve --port $(PORT)

# protoc 重新生成 Go 代码
proto:
	mkdir -p gen
	export PATH="$$PATH:$$(go env GOPATH)/bin" && protoc \
		--proto_path=proto \
		--go_out=gen --go_opt=paths=source_relative \
		--go-grpc_out=gen --go-grpc_opt=paths=source_relative \
		terminal_agent/v1/device.proto \
		terminal_agent/v1/service.proto

# 运行测试
test:
	go test ./...

# 清理编译产物
clean:
	rm -rf bin/ gen/

# go mod tidy
tidy:
	go mod tidy
