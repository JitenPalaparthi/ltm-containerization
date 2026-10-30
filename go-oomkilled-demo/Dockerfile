FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /oom-demo .
FROM alpine:3.21
COPY --from=build /oom-demo /oom-demo
ENTRYPOINT ["/oom-demo"]
