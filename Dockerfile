# Development stage
FROM golang:1.21.1-bookworm AS development

# Set working directory
WORKDIR /go/src/app

# Copy source
COPY . .

# Set workdir to desired service
WORKDIR cmd/serverd

# Fetch packages
RUN go get -d -v ./...

# An ENTRYPOINT allows you to configure a container that will run as an executable.
ENTRYPOINT ["sh", "-c", "go run main.go"]

# Builder stage
FROM development AS builder

# Install
RUN go install -v ./...

# Production stage
FROM debian:bookworm-slim

# Create unprivileged user
RUN useradd -rm -s /bin/sh go

# Run as an unprivileged user
USER go

# Copy executables
COPY --from=builder /go/bin/serverd /usr/local/bin/

# Copy default config
COPY --from=builder /go/src/app/.env.example /etc/serverd.env

# Copy CA certificates to prevent x509: certificate signed by unknown authority errors
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# The main purpose of a CMD is to provide defaults for an executing container.
# These defaults can include an executable, or they can omit the executable
CMD ["sh", "-c", "serverd -config /etc/serverd.env"]
