FROM golang:1.26.5-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

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
