# Build a small static Linux binary.
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/cpu-stress-lab .

# Non-root runtime image.
FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/cpu-stress-lab /usr/local/bin/cpu-stress-lab
USER app
ENTRYPOINT ["/usr/local/bin/cpu-stress-lab"]
