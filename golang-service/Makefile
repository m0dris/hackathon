include .env
export


export PROJECT_ROOT=$(shell pwd)


hackathon-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/hackathon/main.go


proto-gen:
	mkdir -p internal/grpc/pb
	protoc \
		--go_out=. --go_opt=module=github.com/m0dris/hackathon \
		--go-grpc_out=. --go-grpc_opt=module=github.com/m0dris/hackathon \
		api/proto/uk_service.proto
