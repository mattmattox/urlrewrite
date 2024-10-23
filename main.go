package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mattmattox/urlrewrite/pkg/config"
	"github.com/mattmattox/urlrewrite/pkg/logging"
	"github.com/mattmattox/urlrewrite/pkg/version"
	"github.com/sirupsen/logrus"
)

// Ensure the logger is accessible globally.
var log = logging.SetupLogging()

func init() {
	// Load configuration first.
	config.LoadConfiguration()

	// Initialize the logger with appropriate level based on config.
	if config.CFG.Debug {
		log.Debug("Debug mode enabled")
		log.SetLevel(logrus.DebugLevel)
	} else {
		log.SetLevel(logrus.InfoLevel)
	}

	log.Debug("Logger initialized")
}

// requestLogger logs incoming requests with method, path, and query parameters.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Infof("Received request: %s %s%s from %s",
			r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr)

		// Wrap the ResponseWriter to capture the status code.
		wrapped := statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(&wrapped, r)

		log.Infof("Completed request: %s %s%s with status %d in %v",
			r.Method, r.URL.Path, r.URL.RawQuery, wrapped.status, time.Since(start))
	})
}

// statusRecorder wraps the ResponseWriter to track the status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader allows us to capture the HTTP status code.
func (rec *statusRecorder) WriteHeader(code int) {
	rec.status = code
	rec.ResponseWriter.WriteHeader(code)
}

func main() {
	// Log the application version information.
	log.Info("Starting URL rewrite proxy for Feral Hosting")
	log.Infof("Version: %s", version.Version)
	log.Infof("Git Commit: %s", version.GitCommit)
	log.Infof("Build Time: %s", version.BuildTime)

	// Construct the target URL dynamically based on configuration.
	rewriteTarget := "https://" + config.CFG.FeralHostingServer + "/" + config.CFG.FeralHostingUsername

	// Parse the base target URL.
	target, err := url.Parse(rewriteTarget)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}

	// Create a reverse proxy with the base target URL.
	proxy := httputil.NewSingleHostReverseProxy(target)

	// Modify the proxy's director function to append original subpaths.
	proxy.Director = func(req *http.Request) {
		req.Host = target.Host
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		// Join the target path with the original request path.
		req.URL.Path = joinPaths(target.Path, req.URL.Path)
	}

	// Set up the HTTP server with request logging and timeouts.
	addr := "0.0.0.0:" + strconv.Itoa(config.CFG.ServerPort)
	log.Infof("Starting server on %s", addr)

	server := &http.Server{
		Addr:              addr,
		Handler:           requestLogger(proxy),
		ReadHeaderTimeout: 10 * time.Second,  // Mitigate Slowloris attacks.
		ReadTimeout:       30 * time.Second,  // Limit reading the request body.
		WriteTimeout:      30 * time.Second,  // Limit time to write a response.
		IdleTimeout:       120 * time.Second, // Keep-alive timeout.
	}

	// Run the server in a goroutine.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown.
	waitForShutdown(server)
}

// joinPaths ensures correct path joining for the target URL and request path.
func joinPaths(basePath, reqPath string) string {
	// Unconditionally trim slashes from the base and request paths.
	basePath = strings.TrimSuffix(basePath, "/")
	reqPath = strings.TrimPrefix(reqPath, "/")
	return basePath + "/" + reqPath
}

// waitForShutdown gracefully handles server shutdown.
func waitForShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info("Shutting down server...")

	if err := server.Close(); err != nil {
		log.Errorf("Failed to gracefully shut down server: %v", err)
	} else {
		log.Info("Server shut down gracefully.")
	}
}
