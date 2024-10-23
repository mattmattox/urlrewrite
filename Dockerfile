# Use golang alpine image as the builder stage
FROM golang:1.22.4-alpine3.20 AS builder

# Install git and other necessary tools
RUN apk update && apk add --no-cache git bash

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .

# Fetch dependencies using go mod if your project uses Go modules
RUN go mod download

# Version and Git Commit build arguments
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# Compile the Go application and embed version information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags "-X github.com/mattmattox/urlrewrite/pkg/version.Version=$VERSION \
    -X github.com/mattmattox/urlrewrite/pkg/version.GitCommit=$GIT_COMMIT \
    -X github.com/mattmattox/urlrewrite/pkg/version.BuildTime=$BUILD_DATE" \
    -o /bin/urlrewrite

# Use ubuntu as the final image
FROM ubuntu:latest

# Install Common Dependencies
RUN apt-get update && \
    apt install -y \
    ca-certificates \
    curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary and Swagger docs from the builder stage
COPY --from=builder /bin/urlrewrite /root/urlrewrite

# Ensure the binary is executable
RUN chmod +x /root/urlrewrite

EXPOSE 8080

# Command to run the executable
CMD ["./urlrewrite"]
