# Stage 1: build the binary with the full Go toolchain
FROM golang:1.26 AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /nodecheck .

# Stage 2: copy only the binary into a small runtime image
FROM ubuntu:24.04
COPY --from=build /nodecheck /usr/bin/nodecheck
ENTRYPOINT ["/usr/bin/nodecheck"]