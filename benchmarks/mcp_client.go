// Package benchmarks provides MCP client for stdio-based communication
package benchmarks

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// MCPError provides structured error information with context
type MCPError struct {
	RequestID int           // The request ID that failed
	TraceID   string        // Trace ID for request correlation
	Operation string        // The operation being performed (e.g., "initialize", "tool_call")
	ToolName  string        // The tool being called (for tool_call operations)
	Latency   time.Duration // How long the operation took before failing
	Err       error         // The underlying error
}

// Error implements the error interface
func (e *MCPError) Error() string {
	if e.ToolName != "" {
		return fmt.Sprintf("MCP %s failed (trace=%s, req=%d, tool=%s, latency=%v): %v",
			e.Operation, e.TraceID, e.RequestID, e.ToolName, e.Latency, e.Err)
	}
	return fmt.Sprintf("MCP %s failed (trace=%s, req=%d, latency=%v): %v",
		e.Operation, e.TraceID, e.RequestID, e.Latency, e.Err)
}

// Unwrap returns the underlying error
func (e *MCPError) Unwrap() error {
	return e.Err
}

// generateTraceID creates a unique trace ID for request correlation
func generateTraceID(reqID int) string {
	return fmt.Sprintf("trace-%d-%d", time.Now().UnixNano(), reqID)
}

// Config contains configurable parameters for MCPClient
type Config struct {
	// ReadyTimeout is how long to wait for server to become ready
	ReadyTimeout time.Duration
	// InitializeDelay is how long to wait after sending initialized notification
	InitializeDelay time.Duration
	// ToolCallTimeout is the default timeout for tool calls
	ToolCallTimeout time.Duration
	// CloseTimeout is how long to wait for graceful shutdown
	CloseTimeout time.Duration
	// ProtocolVersion is the MCP protocol version to use
	ProtocolVersion string
}

// DefaultConfig returns default configuration values
func DefaultConfig() Config {
	return Config{
		ReadyTimeout:    10 * time.Second,
		InitializeDelay: 100 * time.Millisecond,
		ToolCallTimeout: 30 * time.Second,
		CloseTimeout:    5 * time.Second,
		ProtocolVersion: "2024-11-05",
	}
}

// MCPClient manages communication with an MCP server via stdio
type MCPClient struct {
	serverPath string
	env        []string
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *bufio.Reader
	stderr     *bufio.Reader
	encoder    *json.Encoder
	requestID  int
	mu         sync.Mutex
	writeMu    sync.Mutex
	readMu     sync.Mutex // Protect stdout reader from concurrent access
	waitOnce   sync.Once  // Ensure cmd.Wait() is only called once
	waitErr    error      // Result of cmd.Wait()
	done       chan struct{}
	config     Config
	ctx        context.Context    // P2 #10: Context for lifecycle management
	cancel     context.CancelFunc // P2 #10: Cancel function for cleanup
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Method  string                 `json:"method"`
	Params  map[string]interface{} `json:"params"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string                 `json:"jsonrpc"`
	ID      int                    `json:"id"`
	Result  map[string]interface{} `json:"result,omitempty"`
	Error   *JSONRPCError          `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ToolResponse represents the response from a tool call
type ToolResponse struct {
	Content map[string]interface{}
}

// NewMCPClient creates a new MCP client with optional configuration
func NewMCPClient(serverPath string, env []string, configs ...Config) (*MCPClient, error) {
	if serverPath == "" {
		return nil, fmt.Errorf("server path cannot be empty")
	}

	// Resolve to absolute path if relative
	absPath, err := filepath.Abs(serverPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	// Check if server binary exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("server binary not found: %s", absPath)
	}

	serverPath = absPath

	// Use provided config or default
	config := DefaultConfig()
	if len(configs) > 0 {
		config = configs[0]
	}

	// P2 #10: Create context with cancel for lifecycle management
	ctx, cancel := context.WithCancel(context.Background())

	return &MCPClient{
		serverPath: serverPath,
		env:        env,
		done:       make(chan struct{}),
		config:     config,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start spawns the server process and establishes stdio communication
func (c *MCPClient) Start() error {
	// Validate server path before execution
	if !filepath.IsAbs(c.serverPath) {
		return fmt.Errorf("server path must be absolute: %s", c.serverPath)
	}
	if _, err := os.Stat(c.serverPath); err != nil {
		return fmt.Errorf("server binary not accessible: %w", err)
	}

	// Create command (serverPath validated: absolute path + file exists)
	c.cmd = exec.Command(c.serverPath) // #nosec G204 -- serverPath validated: must be absolute path and exist as file
	c.cmd.Env = append(os.Environ(), c.env...)

	// Setup pipes
	stdin, err := c.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}
	c.stdin = stdin

	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	// Use larger buffer (64KB) to handle large responses (P2 #7)
	c.stdout = bufio.NewReaderSize(stdout, 64*1024)

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}
	c.stderr = bufio.NewReader(stderr)

	// Create encoder for sending requests
	c.encoder = json.NewEncoder(c.stdin)

	// Start server process
	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Monitor process lifecycle
	go c.monitorProcess()

	// Wait for server to be ready
	if err := c.waitForReady(c.config.ReadyTimeout); err != nil {
		_ = c.Close() // Best effort cleanup
		return fmt.Errorf("server not ready: %w", err)
	}

	// Perform MCP protocol handshake
	if err := c.Initialize(); err != nil {
		_ = c.Close() // Best effort cleanup
		return fmt.Errorf("MCP initialization failed: %w", err)
	}

	return nil
}

// Initialize performs the MCP protocol handshake
func (c *MCPClient) Initialize() error {
	start := time.Now()

	// Get next request ID
	c.mu.Lock()
	c.requestID++
	reqID := c.requestID
	c.mu.Unlock()

	traceID := generateTraceID(reqID)

	// Build initialize request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": c.config.ProtocolVersion,
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "unified-thinking-benchmark",
				"version": "1.0.0",
			},
		},
	}

	// Send initialize request
	c.writeMu.Lock()
	if err := c.encoder.Encode(req); err != nil {
		c.writeMu.Unlock()
		return &MCPError{
			RequestID: reqID,
			TraceID:   traceID,
			Operation: "initialize",
			Latency:   time.Since(start),
			Err:       fmt.Errorf("failed to send initialize: %w", err),
		}
	}
	c.writeMu.Unlock()

	// Read initialize response using unified reader
	resp, err := c.readJSONRPCResponse()
	if err != nil {
		return &MCPError{
			RequestID: reqID,
			TraceID:   traceID,
			Operation: "initialize",
			Latency:   time.Since(start),
			Err:       fmt.Errorf("failed to read initialize response: %w", err),
		}
	}

	if resp.Error != nil {
		return &MCPError{
			RequestID: reqID,
			TraceID:   traceID,
			Operation: "initialize",
			Latency:   time.Since(start),
			Err:       fmt.Errorf("initialize failed: %s", resp.Error.Message),
		}
	}

	// Send initialized notification (no response expected)
	notification := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "initialized",
		"params":  map[string]interface{}{},
	}

	c.writeMu.Lock()
	if err := c.encoder.Encode(notification); err != nil {
		c.writeMu.Unlock()
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	c.writeMu.Unlock()

	// Give server a moment to process
	time.Sleep(c.config.InitializeDelay)

	return nil
}

// monitorProcess watches for process exit (P2 #10: cancels context on exit)
func (c *MCPClient) monitorProcess() {
	go func() {
		// Use sync.Once to prevent race with Close() calling Wait()
		c.waitOnce.Do(func() {
			c.waitErr = c.cmd.Wait()
		})
		c.cancel() // Cancel context when process exits
		close(c.done)
	}()
}

// waitForReady waits for server to be ready (optimized with bytes operations)
func (c *MCPClient) waitForReady(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Pre-allocate byte slices for ready signal matching (P2 #7 optimization)
	readySignals := [][]byte{
		[]byte("listening"),
		[]byte("started"),
		[]byte("ready"),
	}

	// Read stderr in background looking for ready signal
	readyChan := make(chan bool, 1)
	errChan := make(chan error, 1)

	go func() {
		for {
			// Use ReadBytes instead of ReadString for better performance
			line, err := c.stderr.ReadBytes('\n')
			if err != nil {
				errChan <- err
				return
			}

			// Look for ready signals in stderr using bytes.Contains
			for _, signal := range readySignals {
				if bytes.Contains(line, signal) {
					readyChan <- true
					return
				}
			}
		}
	}()

	select {
	case <-readyChan:
		return nil
	case err := <-errChan:
		if err == io.EOF {
			return fmt.Errorf("server exited before becoming ready")
		}
		return err
	case <-ctx.Done():
		// Timeout - server likely ready but didn't log expected message
		// Give it a moment and assume ready
		time.Sleep(500 * time.Millisecond)
		return nil
	case <-c.done:
		return fmt.Errorf("server process exited")
	}
}

// retryConfig defines retry behavior
type retryConfig struct {
	maxRetries   int
	initialDelay time.Duration
}

// defaultRetryConfig returns default retry configuration
func defaultRetryConfig() retryConfig {
	return retryConfig{
		maxRetries:   3,
		initialDelay: 100 * time.Millisecond,
	}
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	// Don't retry on RPC errors (these are permanent)
	if contains(err.Error(), "RPC error") {
		return false
	}
	// Retry on timeouts, connection issues, and temporary failures
	return contains(err.Error(), "timeout") ||
		contains(err.Error(), "cancelled") ||
		contains(err.Error(), "failed to read") ||
		contains(err.Error(), "connection")
}

// CallTool sends a tool call request and returns the response with retry logic
func (c *MCPClient) CallTool(toolName string, args map[string]interface{}) (*ToolResponse, error) {
	config := defaultRetryConfig()
	var lastErr error

	for attempt := 0; attempt <= config.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 100ms, 200ms, 400ms
			// Cap shift to prevent overflow (max 30 prevents int overflow)
			shift := attempt - 1
			if shift > 30 {
				shift = 30
			}
			delay := config.initialDelay * time.Duration(1<<shift)
			time.Sleep(delay)
		}

		resp, err := c.callToolOnce(toolName, args)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		if !isRetryableError(err) {
			return nil, err
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// readJSONRPCResponse reads a single JSON-RPC response from stdout
// This provides unified response parsing for all RPC methods
func (c *MCPClient) readJSONRPCResponse() (*JSONRPCResponse, error) {
	// Protect stdout reader from concurrent access
	c.readMu.Lock()
	defer c.readMu.Unlock()

	line, err := c.stdout.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &resp, nil
}

// callToolOnce sends a single tool call request without retry
func (c *MCPClient) callToolOnce(toolName string, args map[string]interface{}) (*ToolResponse, error) {
	start := time.Now()

	// Check if process is still running
	select {
	case <-c.done:
		return nil, fmt.Errorf("server process has exited")
	default:
	}

	// Get next request ID
	c.mu.Lock()
	c.requestID++
	reqID := c.requestID
	c.mu.Unlock()

	traceID := generateTraceID(reqID)

	// Build request
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      reqID,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      toolName,
			"arguments": args,
		},
	}

	// Send request (with write lock)
	c.writeMu.Lock()
	if err := c.encoder.Encode(req); err != nil {
		c.writeMu.Unlock()
		return nil, &MCPError{
			RequestID: reqID,
			TraceID:   traceID,
			Operation: "tool_call",
			ToolName:  toolName,
			Latency:   time.Since(start),
			Err:       fmt.Errorf("failed to send request: %w", err),
		}
	}
	c.writeMu.Unlock()

	// Read response with context timeout
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ToolCallTimeout)
	defer cancel()

	resp, err := c.readResponse(ctx)
	if err != nil {
		return nil, &MCPError{
			RequestID: reqID,
			TraceID:   traceID,
			Operation: "tool_call",
			ToolName:  toolName,
			Latency:   time.Since(start),
			Err:       err,
		}
	}

	return resp, nil
}

// readResponse reads a JSON-RPC response with context-based cancellation
func (c *MCPClient) readResponse(ctx context.Context) (*ToolResponse, error) {
	respChan := make(chan *JSONRPCResponse, 1)
	errChan := make(chan error, 1)

	// Note: Don't close channels here - let them be garbage collected
	// Closing them causes "send on closed channel" panic in concurrent scenarios

	// Read response in goroutine with context awareness
	go func() {
		// Recover from panics caused by closed reader during teardown
		defer func() {
			if r := recover(); r != nil {
				select {
				case errChan <- fmt.Errorf("reader panic (likely closed connection): %v", r):
				case <-ctx.Done():
				}
			}
		}()

		// Check context before starting
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Use unified response reader
		resp, err := c.readJSONRPCResponse()
		if err != nil {
			select {
			case errChan <- err:
			case <-ctx.Done():
			}
			return
		}

		select {
		case respChan <- resp:
		case <-ctx.Done():
		}
	}()

	// Wait for response with context cancellation
	select {
	case resp := <-respChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("RPC error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return &ToolResponse{Content: resp.Result}, nil
	case err := <-errChan:
		return nil, fmt.Errorf("failed to read response: %w", err)
	case <-ctx.Done():
		return nil, fmt.Errorf("request cancelled: %w", ctx.Err())
	case <-c.done:
		return nil, fmt.Errorf("server process exited")
	}
}

// Close gracefully shuts down the server process (P2 #10: cancels context first)
func (c *MCPClient) Close() error {
	// P2 #10: Cancel context to signal all goroutines to stop
	if c.cancel != nil {
		c.cancel()
	}

	if c.stdin != nil {
		_ = c.stdin.Close() // Best effort cleanup
	}

	if c.cmd == nil || c.cmd.Process == nil {
		return nil
	}

	// Try graceful shutdown first (SIGINT)
	if err := c.cmd.Process.Signal(os.Interrupt); err != nil {
		// If signal fails, try SIGKILL
		return c.cmd.Process.Kill()
	}

	// Wait for process to exit (with timeout)
	// Use sync.Once to prevent race with monitorProcess() calling Wait()
	done := make(chan struct{})
	go func() {
		c.waitOnce.Do(func() {
			c.waitErr = c.cmd.Wait()
		})
		close(done)
	}()

	select {
	case <-done:
		return c.waitErr
	case <-time.After(c.config.CloseTimeout):
		// Force kill after timeout
		_ = c.cmd.Process.Kill()
		return fmt.Errorf("forced kill after timeout")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) >= len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				containsInMiddle(s, substr)))
}

func containsInMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
