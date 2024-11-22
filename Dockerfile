FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download

COPY main.go .
COPY cmd/ cmd/
COPY internal/ internal/
COPY models/ models/

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a \
    -o app main.go

#FROM gcr.io/distroless/static:nonroot
FROM alpine:latest


COPY --from=builder /workspace/app /usr/local/bin/app
USER 65532:65532

ENTRYPOINT ["app"]