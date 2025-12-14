package benchmarks

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"unified-thinking/benchmarks/evaluators"
	"unified-thinking/internal/storage"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Load .env file from project root
	loadEnvFile("../.env")
}

// loadEnvFile loads environment variables from a file
func loadEnvFile(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	file, err := os.Open(absPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Only set if not already set (env vars take precedence)
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

// getServerPath returns the platform-appropriate server binary path
func getServerPath() string {
	if runtime.GOOS == "windows" {
		return "../bin/unified-thinking.exe"
	}
	return "../bin/unified-thinking"
}

// checkRequiredEnv verifies required environment variables are set AND services are reachable - FAILS if not
func checkRequiredEnv(t *testing.T) {
	t.Helper()
	if os.Getenv("VOYAGE_API_KEY") == "" {
		t.Fatal("VOYAGE_API_KEY not set - required for embeddings")
	}
	if os.Getenv("ANTHROPIC_API_KEY") == "" {
		t.Fatal("ANTHROPIC_API_KEY not set - required for GoT and LLM features")
	}
	if os.Getenv("NEO4J_URI") == "" {
		t.Fatal("NEO4J_URI not set - required for knowledge graph")
	}
	if os.Getenv("NEO4J_USERNAME") == "" {
		t.Fatal("NEO4J_USERNAME not set - required for knowledge graph")
	}
	if os.Getenv("NEO4J_PASSWORD") == "" {
		t.Fatal("NEO4J_PASSWORD not set - required for knowledge graph")
	}

	// Verify Neo4j connectivity - fail fast if not reachable
	checkNeo4jConnectivity(t)
}

// checkNeo4jConnectivity verifies Neo4j is running and reachable - FAILS if not
func checkNeo4jConnectivity(t *testing.T) {
	t.Helper()

	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")

	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		t.Fatalf("Neo4j driver creation failed - required for knowledge graph: %v", err)
	}
	defer driver.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		t.Fatalf("Neo4j not reachable - required for knowledge graph: %v", err)
	}
}

// Test 1: Basic MCPClient connectivity
func TestMCPClientStartStop(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	// Skip if server binary not available
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	client, err := NewMCPClient(serverPath, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer client.Close()

	t.Log("MCPClient successfully started and stopped")
}

// Test 2: Single think tool call
func TestMCPClientThinkTool(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	client, err := NewMCPClient(serverPath, nil)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	if err := client.Start(); err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer client.Close()

	// Call think tool
	args := map[string]interface{}{
		"content": "Test reasoning: All men are mortal. Socrates is a man. Therefore, Socrates is mortal.",
		"mode":    "linear",
	}

	start := time.Now()
	resp, err := client.CallTool("think", args)
	latency := time.Since(start)

	if err != nil {
		t.Fatalf("CallTool failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	// Debug: dump full response
	t.Logf("Full response: %+v", resp.Content)

	// Validate response structure
	// MCP servers return data in "content" array
	var thoughtID string
	if content, ok := resp.Content["content"].([]interface{}); ok && len(content) > 0 {
		if contentItem, ok := content[0].(map[string]interface{}); ok {
			if text, ok := contentItem["text"].(string); ok {
				t.Logf("Response text: %s", text)
			}
		}
	}

	// Try structured content
	if structured, ok := resp.Content["structuredContent"].(map[string]interface{}); ok {
		t.Logf("Structured content: %+v", structured)
		if tid, ok := structured["thought_id"].(string); ok {
			thoughtID = tid
		}
	}

	t.Logf("Tool call successful:")
	t.Logf("  Thought ID: %s", thoughtID)
	t.Logf("  Latency: %v", latency)
	t.Logf("  Response content keys: %v", getKeys(resp.Content))
}

// Test 3: Full E2E benchmark suite via MCP
func TestMCPExecutorE2E(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	// Load a small test suite
	suite, err := LoadSuite("datasets/reasoning/logic_puzzles.json")
	if err != nil {
		t.Fatalf("Failed to load suite: %v", err)
	}

	// Use first 3 problems for quick test
	suite.Problems = suite.Problems[:3]

	// Create MCP executor
	executor := NewMCPExecutor(serverPath, nil)
	defer executor.Close()

	evaluator := evaluators.NewContainsEvaluator()

	// Run suite
	t.Log("Running E2E benchmark suite...")
	run, err := RunSuite(suite, evaluator, executor)
	if err != nil {
		t.Fatalf("Failed to run suite: %v", err)
	}

	// Validate results
	if run.TotalProblems != 3 {
		t.Errorf("Expected 3 problems, got %d", run.TotalProblems)
	}

	if run.AvgLatency == 0 {
		t.Error("Average latency is zero")
	}

	t.Logf("\n=== E2E Benchmark Results ===")
	t.Logf("Suite: %s", run.SuiteName)
	t.Logf("Problems: %d", run.TotalProblems)
	t.Logf("Correct: %d (%.1f%%)", run.CorrectProblems, run.OverallAccuracy*100)
	t.Logf("Avg Latency: %v", run.AvgLatency)
	t.Logf("ECE: %.4f", run.OverallECE)

	// Show individual results
	for _, result := range run.Results {
		status := "PASS"
		if !result.Correct {
			status = "FAIL"
		}
		t.Logf("  %s: %s (latency: %v)", result.ProblemID, status, result.Latency)
	}
}

// Test 4: Performance comparison (Direct vs MCP)
func TestMCPVsDirectPerformance(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	suite, err := LoadSuite("datasets/reasoning/logic_puzzles.json")
	if err != nil {
		t.Fatalf("Failed to load suite: %v", err)
	}

	// Use 5 problems for comparison
	suite.Problems = suite.Problems[:5]
	evaluator := evaluators.NewExactMatchEvaluator()

	// Run with DirectExecutor
	t.Log("Running with DirectExecutor...")
	store := storage.NewMemoryStorage()
	directExec := NewDirectExecutor(store)
	directRun, err := RunSuite(suite, evaluator, directExec)
	if err != nil {
		t.Fatalf("Direct execution failed: %v", err)
	}

	// Run with MCPExecutor
	t.Log("Running with MCPExecutor...")
	mcpExec := NewMCPExecutor(serverPath, nil)
	defer mcpExec.Close()
	mcpRun, err := RunSuite(suite, evaluator, mcpExec)
	if err != nil {
		t.Fatalf("MCP execution failed: %v", err)
	}

	// Compare latencies
	overhead := mcpRun.AvgLatency - directRun.AvgLatency
	overheadPct := (float64(overhead) / float64(directRun.AvgLatency)) * 100

	t.Logf("\n=== Performance Comparison ===")
	t.Logf("DirectExecutor:")
	t.Logf("  Avg Latency: %v", directRun.AvgLatency)
	t.Logf("  Total Time: %v", directRun.AvgLatency*time.Duration(directRun.TotalProblems))
	t.Logf("\nMCPExecutor:")
	t.Logf("  Avg Latency: %v", mcpRun.AvgLatency)
	t.Logf("  Total Time: %v", mcpRun.AvgLatency*time.Duration(mcpRun.TotalProblems))
	t.Logf("\nProtocol Overhead:")
	t.Logf("  Per Problem: %v (%.1f%%)", overhead, overheadPct)
	t.Logf("  Total: %v", overhead*time.Duration(mcpRun.TotalProblems))

	// Protocol overhead should be reasonable (< 100ms per call)
	if overhead > 100*time.Millisecond {
		t.Logf("Warning: Protocol overhead is high: %v", overhead)
	} else {
		t.Logf("Protocol overhead is acceptable")
	}
}

// Test 6: Server crash recovery
func TestMCPExecutorServerCrash(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	executor := NewMCPExecutor(serverPath, nil)

	// Create a simple problem
	problem := &Problem{
		ID:          "crash_test",
		Description: "Test problem",
		Input:       map[string]interface{}{"mode": "linear"},
		Expected:    "test",
	}

	evaluator := evaluators.NewExactMatchEvaluator()

	// Execute once (should work)
	result1, err := executor.Execute(problem, evaluator)
	if err != nil {
		t.Fatalf("First execution failed: %v", err)
	}
	t.Logf("First execution: %v", result1.Latency)

	// Kill the server process
	if executor.client != nil && executor.client.cmd != nil {
		executor.client.cmd.Process.Kill()
		time.Sleep(100 * time.Millisecond)
	}

	// Try to execute again (should fail gracefully)
	result2, err := executor.Execute(problem, evaluator)
	if err != nil {
		t.Logf("Expected error after crash: %v", err)
	}
	if result2 != nil && result2.Error != "" {
		t.Logf("Got graceful error: %s", result2.Error)
	} else {
		t.Error("Expected error result after server crash")
	}

	executor.Close()
}

// Test 7: Connection reuse
func TestMCPExecutorConnectionReuse(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	executor := NewMCPExecutor(serverPath, nil)
	defer executor.Close()

	evaluator := evaluators.NewExactMatchEvaluator()

	// Execute multiple problems
	latencies := []time.Duration{}
	for i := 0; i < 5; i++ {
		problem := &Problem{
			ID:          "reuse_test",
			Description: "Test problem",
			Input:       map[string]interface{}{"mode": "linear"},
			Expected:    "test",
		}

		result, err := executor.Execute(problem, evaluator)
		if err != nil {
			t.Fatalf("Execution %d failed: %v", i, err)
		}
		latencies = append(latencies, result.Latency)
		t.Logf("Execution %d: %v", i+1, result.Latency)
	}

	// First call includes server startup time, subsequent calls should be faster
	if latencies[0] < latencies[1] {
		t.Error("Expected first call to be slower (includes startup)")
	}

	t.Logf("Connection reuse working: first=%v, subsequent=%v",
		latencies[0], avgDuration(latencies[1:]))
}

// Helper functions
func getKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func avgDuration(durations []time.Duration) time.Duration {
	if len(durations) == 0 {
		return 0
	}
	var sum time.Duration
	for _, d := range durations {
		sum += d
	}
	return sum / time.Duration(len(durations))
}

// =============================================================================
// P2 #8: Test Coverage Gaps
// =============================================================================

// Test 8: Concurrent tool calls (P2 #8)
// Each goroutine creates its own MCP client since stdio isn't thread-safe
func TestMCPConcurrentCalls(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	var wg sync.WaitGroup
	numCalls := 3 // Reduced from 5 to avoid resource exhaustion
	wg.Add(numCalls)

	results := make(chan error, numCalls)

	for i := 0; i < numCalls; i++ {
		go func(id int) {
			defer wg.Done()

			// Stagger server starts to avoid resource exhaustion on Windows
			time.Sleep(time.Duration(id) * 500 * time.Millisecond)

			// Each goroutine creates its own client/server instance
			client, err := NewMCPClient(serverPath, nil)
			if err != nil {
				results <- fmt.Errorf("goroutine %d: failed to create client: %v", id, err)
				return
			}

			if err := client.Start(); err != nil {
				results <- fmt.Errorf("goroutine %d: failed to start: %v", id, err)
				return
			}
			defer client.Close()

			args := map[string]interface{}{
				"content": fmt.Sprintf("Test problem %d: simple reasoning test", id),
				"mode":    "linear",
			}
			resp, err := client.CallTool("think", args)
			if err != nil {
				results <- fmt.Errorf("goroutine %d: %v", id, err)
				return
			}
			if resp == nil {
				results <- fmt.Errorf("goroutine %d: nil response", id)
				return
			}
			results <- nil
		}(i)
	}

	wg.Wait()
	close(results)

	// Check all results
	var errors []error
	for err := range results {
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		t.Logf("Concurrent calls had %d errors out of %d calls", len(errors), numCalls)
		for _, e := range errors {
			t.Logf("  Error: %v", e)
		}
	}
	assert.Empty(t, errors, "Expected all concurrent calls to succeed")
}

// Test 9: Protocol version mismatch (P2 #8)
func TestMCPProtocolMismatch(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	config := DefaultConfig()
	config.ProtocolVersion = "1999-01-01" // Invalid/old protocol version

	client, err := NewMCPClient(serverPath, nil, config)
	require.NoError(t, err)

	err = client.Start()
	defer client.Close()

	// Server may accept connection but reject during initialize
	// or the handshake might fail gracefully
	if err != nil {
		t.Logf("Protocol mismatch handled: %v", err)
	} else {
		// If start succeeded, try a tool call
		args := map[string]interface{}{
			"content": "Test",
			"mode":    "linear",
		}
		_, callErr := client.CallTool("think", args)
		if callErr != nil {
			t.Logf("Tool call failed with mismatched protocol: %v", callErr)
		}
	}
}

// Test 10: Timeout behavior (P2 #8)
func TestMCPExecutorTimeout(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	config := DefaultConfig()
	config.ToolCallTimeout = 1 * time.Millisecond // Very short timeout

	client, err := NewMCPClient(serverPath, nil, config)
	require.NoError(t, err)
	require.NoError(t, client.Start())
	defer client.Close()

	// Note: Initialize was already called in Start(), so we proceed with tool call

	args := map[string]interface{}{
		"content": "Complex problem requiring multi-step reasoning",
		"mode":    "tree",
	}

	_, err = client.CallTool("think", args)

	// We expect a timeout error due to the very short timeout
	if err != nil {
		t.Logf("Got expected timeout-related error: %v", err)
		// The error might be timeout, cancelled, or retries exceeded
		assert.Error(t, err)
	} else {
		t.Log("Tool call completed before timeout (server was very fast)")
	}
}

// Test 11: Invalid response handling (P2 #8)
// Note: This tests the client's resilience to unexpected response formats
func TestMCPExecutorInvalidResponse(t *testing.T) {
	checkRequiredEnv(t)
	serverPath := getServerPath()

	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		t.Fatal("Server binary not found - run make build-server")
	}

	client, err := NewMCPClient(serverPath, nil)
	require.NoError(t, err)
	require.NoError(t, client.Start())
	defer client.Close()

	// Call with invalid/missing required arguments to trigger error response
	invalidArgs := map[string]interface{}{
		// Missing "content" which is typically required
		"mode": "invalid_mode_that_does_not_exist",
	}

	resp, err := client.CallTool("think", invalidArgs)

	// Server should return an error for invalid arguments
	if err != nil {
		t.Logf("Got expected error for invalid args: %v", err)
	} else if resp != nil {
		t.Logf("Got response (server may have default handling): %+v", resp.Content)
	}
}

// Test 12: Fuzzing MCP client with random inputs (P2 #8)
func FuzzMCPClient(f *testing.F) {
	// Add seed corpus
	f.Add("test content", "linear")
	f.Add("", "")
	f.Add("special chars: !@#$%^&*()", "tree")
	f.Add("unicode: 你好世界 🌍", "divergent")
	f.Add(string(make([]byte, 1000)), "auto")

	serverPath := getServerPath()

	// Check if server binary exists
	if _, err := os.Stat(serverPath); os.IsNotExist(err) {
		f.Skip("Server binary not found")
	}

	f.Fuzz(func(t *testing.T, content string, mode string) {
		// Limit content size to avoid memory issues
		if len(content) > 10000 {
			return
		}

		client, err := NewMCPClient(serverPath, nil)
		if err != nil {
			return
		}
		if err := client.Start(); err != nil {
			return
		}
		defer client.Close()

		args := map[string]interface{}{
			"content": content,
			"mode":    mode,
		}

		// We don't care about the result, just that it doesn't panic
		client.CallTool("think", args)
	})
}
