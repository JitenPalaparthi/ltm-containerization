FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /image-server .

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app -u 10001 app && mkdir -p /data/images && chown -R app:app /data
COPY --from=build /image-server /image-server
USER 10001:10001
EXPOSE 8080
ENTRYPOINT ["/image-server"]
