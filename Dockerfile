ARG GO_VERSION=1.20
ARG ALPINE_VERSION=3.17
ARG BASE_IMAGE=scratch

### Build binary
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} as build-binary
ENV GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0
COPY . /go/src/github.com/deepmap/oapi-codegen
WORKDIR /go/src/github.com/deepmap/oapi-codegen
RUN go build -o ./bin/oapi-codegen ./cmd/oapi-codegen/oapi-codegen.go

### User/passwd layer
FROM alpine:${ALPINE_VERSION} as user-passwd
RUN addgroup -S oapi-codegen && adduser -S oapi-codegen -G oapi-codegen

### Image
FROM ${BASE_IMAGE}
COPY --from=user-passwd /etc/passwd /etc/passwd
COPY --from=build-binary /go/src/github.com/deepmap/oapi-codegen/bin/oapi-codegen /usr/local/bin/oapi-codegen
USER oapi-codegen
ENTRYPOINT ["/usr/local/bin/oapi-codegen"]
