package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/Silicon-Ally/zapgcp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if err := run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(args []string) error {
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		minLogLevel zapcore.Level = zapcore.InfoLevel

		local = fs.Bool("local", true, "If true, configure logging for running locally.")
		port  = fs.Int("port", 8080, "The port to run the test HTTP server on")
	)
	fs.Var(&minLogLevel, "min_log_level", "If set, retains logs at the given level and above. Options: 'debug', 'info', 'warn', 'error', 'dpanic', 'panic', 'fatal' - default warn.")
	if err := fs.Parse(args[1:]); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	logger, err := zapgcp.New(&zapgcp.Config{
		Local:       *local,
		MinLogLevel: minLogLevel,
	})
	if err != nil {
		return fmt.Errorf("failed to init logger: %w", err)
	}
	defer logger.Sync()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			logger.Info("request on root path")
		} else {
			logger.Info("request on unhandled path", zap.String("path", r.URL.Path))
		}
	})
	http.HandleFunc("/admin/", func(w http.ResponseWriter, r *http.Request) {
		logger.Warn("received admin request", zap.String("path", r.URL.Path))
		if r.Header.Get("Authorization") == "" {
			logger.Error("unauthorized user sending admin requests",
				zap.String("path", r.URL.Path),
				zap.String("user_agent", r.Header.Get("User-Agent")))
			http.Error(w, "permission denied", http.StatusForbidden)
			return
		}
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	errC := make(chan error)
	go func() {
		errC <- http.ListenAndServe(":"+strconv.Itoa(*port), nil)
	}()

	select {
	case sig := <-c:
		logger.Info("received signal, shutting down", zap.String("signal", sig.String()))
	case err := <-errC:
		logger.Error("error while serving HTTP", zap.Error(err))
	}
	return nil
}
