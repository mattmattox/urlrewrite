# URL Rewrite Proxy for Feral Hosting

This project is a URL rewrite proxy designed to work with [Feral Hosting](https://www.feralhosting.com/) servers. It proxies requests to specific paths on your Feral Hosting server by rewriting URLs dynamically. This is particularly useful for services like Radarr, Sonarr, and Lidarr, which don't support custom base URLs.

## Overview
The project consists of the following key components:

### `pkg/version/version.go`
- Contains version information for the application.
- Provides the `GetVersion` function to return the version in JSON format.

### `pkg/config/config.go`
- Manages the configuration of the application.
- Defines the `AppConfig` structure to load configuration from environment variables and command line flags.

### `pkg/logging/logging.go`
- Handles logging using [Logrus](https://github.com/sirupsen/logrus).
- Includes the `SetupLogging` function to initialize the logger based on configuration.

### `main.go`
- The main entry point of the application.
- Sets up the URL rewrite proxy using the provided configuration.
- Defines request logging middleware and manages graceful server shutdown.

## Getting Started

1. **Set up Configuration:**
   - Use environment variables or command line flags to configure the application.
     - Example environment variables:
       ```bash
       export FERAL_HOSTING_SERVER="server.feralhosting.com"
       export FERAL_HOSTING_USERNAME="yourusername"
       export SERVER_PORT=8080
       ```

2. **Run the Application:**
   ```bash
   go run main.go
   ```

3. **Check Version:**
   - Use the `-version` flag to display the version information:
     ```bash
     go run main.go -version
     ```

## Usage

- Configure your Feral Hosting server and username in the environment variables.
- Run the application using Docker or directly with Go. The proxy will forward requests to the Feral Hosting server based on the original paths.

### **Run with Docker:**
```bash
docker run -d -p 8080:8080 \
  -e FERAL_HOSTING_SERVER=server.feralhosting.com \
  -e FERAL_HOSTING_USERNAME=username \
  -e SERVER_PORT=8080 \
  cube8021/urlrewrite
```

## Dependencies
- [logrus](https://github.com/sirupsen/logrus): A structured logging library for Go.

## License
This project is licensed under the Apache 2.0 License. See the [LICENSE](LICENSE) file for more details.
