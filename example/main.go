package main

import (
	"encoding/json"
	"fmt"
	go_xmlrpc2 "github.com/hheconvit23012002/go-xmlrpc/go-xmlrpc"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// SlogAdapter adapts slog to the LoggerInterface
type SlogAdapter struct {
	logger *slog.Logger
}

func (s *SlogAdapter) Debug(msg string, args ...interface{}) {
	s.logger.Debug(msg, args...)
}

func (s *SlogAdapter) Info(msg string, args ...interface{}) {
	s.logger.Info(msg, args...)
}

func (s *SlogAdapter) Error(msg string, args ...interface{}) {
	s.logger.Error(msg, args...)
}

// CallParams represents the expected request parameters
type CallParams struct {
	CallCenterPhone string `json:"call_center_phone"`
	CustomerPhone   string `json:"customer_phone"`
}

// CallResponse represents the response structure
type CallResponse struct {
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CallID    string    `json:"call_id"`
	Timestamp time.Time `json:"timestamp"`
}

// InitCallInHandler handles the InitCallIn method
type InitCallInHandler struct {
	logger go_xmlrpc2.LoggerInterface
}

func (h *InitCallInHandler) Handle(params []go_xmlrpc2.ParamValue) (interface{}, error) {
	if len(params) == 0 || params[0].Value.StructValue == nil {
		return nil, fmt.Errorf("invalid parameters")
	}

	// Extract JsonData
	var jsonData string
	for _, member := range params[0].Value.StructValue.Members {
		if member.Name == "JsonData" {
			jsonData = member.Value.StringValue
			break
		}
	}

	if jsonData == "" {
		return nil, fmt.Errorf("missing JsonData")
	}

	// Parse parameters
	var callParams CallParams
	if err := json.Unmarshal([]byte(jsonData), &callParams); err != nil {
		return nil, fmt.Errorf("invalid JSON data: %w", err)
	}

	// Create response
	response := CallResponse{
		Status:    "success",
		Message:   "Call initiated successfully",
		CallID:    fmt.Sprintf("CALL-%s-%d", callParams.CustomerPhone[:6], time.Now().Unix()),
		Timestamp: time.Now(),
	}

	return response, nil
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	loggerAdapter := &SlogAdapter{logger: logger}

	// Create and configure XML-RPC server
	server := go_xmlrpc2.NewServer(go_xmlrpc2.ServerConfig{
		Logger: loggerAdapter,
	})

	// Register handlers
	server.RegisterHandler("InitCallIn", &InitCallInHandler{
		logger: loggerAdapter,
	})

	// Setup HTTP server
	httpServer := &http.Server{
		Addr:         ":8054",
		Handler:      server,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	// Start server
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("starting XML-RPC server", "address", httpServer.Addr)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Handle shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err)
	case sig := <-shutdown:
		logger.Info("shutdown signal received", "signal", sig)

		if err := httpServer.Close(); err != nil {
			logger.Error("server close failed", "error", err)
		}
	}
}
