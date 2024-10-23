# Builder Stage: Use bci-base to install dependencies and build the binary
FROM registry.suse.com/bci/bci-base:15.6 AS builder

# Install required packages with zypper
RUN zypper --non-interactive refresh && \
    zypper --non-interactive install curl tar git bind-utils && \
    zypper clean --all && \
    rm -rf /var/cache/zypp/* /var/log/zypp/* /tmp/* /var/tmp/*

# Set the working directory for the build
WORKDIR /src

# Copy the source code into the working directory
COPY . .

# Install Go and fetch dependencies
RUN curl -L https://go.dev/dl/go1.22.5.linux-amd64.tar.gz | tar -C /usr/local -xzf - \
    && /usr/local/go/bin/go mod download

# Ensure Go binaries are installed in the correct path
ENV PATH="/root/go/bin:/usr/local/go/bin:$PATH"

# Build arguments for versioning
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# Compile the Go application and embed version information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo \
    -ldflags "-X github.com/mattmattox/urlrewrite/pkg/version.Version=$VERSION \
    -X github.com/mattmattox/urlrewrite/pkg/version.GitCommit=$GIT_COMMIT \
    -X github.com/mattmattox/urlrewrite/pkg/version.BuildTime=$BUILD_DATE" \
    -o /bin/urlrewrite

# Final Stage: Use bci-micro to minimize the image size
FROM registry.suse.com/bci/bci-micro:15.6 AS final

# Set the working directory inside the container
WORKDIR /root/

# Copy the built binary and Swagger docs from the builder stage
COPY --from=builder /bin/urlrewrite /root/urlrewrite

# Ensure the binary is executable
RUN chmod +x /root/urlrewrite

EXPOSE 8080

# Command to run the executable
CMD ["./urlrewrite"]
