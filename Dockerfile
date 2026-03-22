FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /ames-unifi-mcp ./cmd/ames-unifi-mcp/

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /ames-unifi-mcp /ames-unifi-mcp
ENTRYPOINT ["/ames-unifi-mcp"]
