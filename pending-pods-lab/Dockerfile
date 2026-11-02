FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/pending-pods-lab .

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/pending-pods-lab /usr/local/bin/pending-pods-lab
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/pending-pods-lab"]
