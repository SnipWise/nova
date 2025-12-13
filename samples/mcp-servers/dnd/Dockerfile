FROM --platform=$BUILDPLATFORM golang:1.25.2-alpine AS builder
ARG TARGETOS
ARG TARGETARCH
WORKDIR /app
COPY . .

RUN <<EOF
go mod tidy
GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build
EOF

FROM alpine:latest
RUN apk --no-cache add ca-certificates wget
WORKDIR /app
COPY --from=builder /app/mcp-dnd-server .
ENTRYPOINT ["./mcp-dnd-server"]
