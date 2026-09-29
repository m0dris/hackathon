FROM golang:1.26-alpine AS build

RUN apk add --no-cache git protobuf protobuf-dev

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.35.2 && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p internal/grpc/pb && \
    protoc \
      --go_out=. --go_opt=module=github.com/m0dris/hackathon \
      --go-grpc_out=. --go-grpc_opt=module=github.com/m0dris/hackathon \
      api/proto/uk_service.proto

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/hackathon ./cmd/hackathon

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S -G app app && \
    mkdir -p /app/logs && chown -R app:app /app

WORKDIR /app
COPY --from=build /out/hackathon /app/hackathon

USER app

ENV LOGGER_FOLDER=/app/logs

EXPOSE 8080

ENTRYPOINT ["/app/hackathon"]
