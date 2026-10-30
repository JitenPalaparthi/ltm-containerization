FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app main.go

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
USER app
COPY --from=builder /app /app
EXPOSE 8080
ENTRYPOINT ["/app"]
