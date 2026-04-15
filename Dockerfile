# Build stage
FROM quay.io/projectquay/golang:1.26 AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /build

COPY go.mod ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w" -o resolver .

# Runtime stage
FROM --platform=$TARGETPLATFORM quay.io/curl/curl:latest

COPY --from=builder /build/resolver /usr/local/bin/resolver

USER 65532:65532

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/resolver"]
