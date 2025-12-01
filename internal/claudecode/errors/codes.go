// Package errors provides structured error handling with recovery suggestions.
//
// Error codes are organized into categories:
//   - 1xxx: Resource errors (not found, already exists)
//   - 2xxx: Validation errors (invalid parameter, missing required)
//   - 3xxx: State errors (session active, graph finalized)
//   - 4xxx: External errors (embedding failed, neo4j connection)
//   - 5xxx: Limit errors (rate limited, context too large)
package errors

// Error codes for resource errors (1xxx)
const (
	// ErrThoughtNotFound indicates a thought ID was not found
	ErrThoughtNotFound = "ERR_1001_THOUGHT_NOT_FOUND"
	// ErrBranchNotFound indicates a branch ID was not found
	ErrBranchNotFound = "ERR_1002_BRANCH_NOT_FOUND"
	// ErrSessionNotFound indicates a session ID was not found
	ErrSessionNotFound = "ERR_1003_SESSION_NOT_FOUND"
	// ErrGraphNotFound indicates a graph ID was not found
	ErrGraphNotFound = "ERR_1004_GRAPH_NOT_FOUND"
	// ErrCheckpointNotFound indicates a checkpoint ID was not found
	ErrCheckpointNotFound = "ERR_1005_CHECKPOINT_NOT_FOUND"
	// ErrDecisionNotFound indicates a decision ID was not found
	ErrDecisionNotFound = "ERR_1006_DECISION_NOT_FOUND"
	// ErrWorkflowNotFound indicates a workflow ID was not found
	ErrWorkflowNotFound = "ERR_1007_WORKFLOW_NOT_FOUND"
	// ErrPresetNotFound indicates a preset ID was not found
	ErrPresetNotFound = "ERR_1008_PRESET_NOT_FOUND"
	// ErrEntityNotFound indicates an entity ID was not found
	ErrEntityNotFound = "ERR_1009_ENTITY_NOT_FOUND"
	// ErrBeliefNotFound indicates a belief ID was not found
	ErrBeliefNotFound = "ERR_1010_BELIEF_NOT_FOUND"
)

// Error codes for validation errors (2xxx)
const (
	// ErrInvalidParameter indicates a parameter has an invalid value
	ErrInvalidParameter = "ERR_2001_INVALID_PARAMETER"
	// ErrMissingRequired indicates a required parameter is missing
	ErrMissingRequired = "ERR_2002_MISSING_REQUIRED"
	// ErrInvalidMode indicates an invalid thinking mode was specified
	ErrInvalidMode = "ERR_2003_INVALID_MODE"
	// ErrInvalidConfidence indicates a confidence value is out of range
	ErrInvalidConfidence = "ERR_2004_INVALID_CONFIDENCE"
	// ErrInvalidFormat indicates an invalid format level was specified
	ErrInvalidFormat = "ERR_2005_INVALID_FORMAT"
	// ErrInvalidOperation indicates the operation is not valid
	ErrInvalidOperation = "ERR_2006_INVALID_OPERATION"
	// ErrInvalidJSON indicates JSON parsing failed
	ErrInvalidJSON = "ERR_2007_INVALID_JSON"
	// ErrInvalidMergeStrategy indicates an invalid merge strategy
	ErrInvalidMergeStrategy = "ERR_2008_INVALID_MERGE_STRATEGY"
	// ErrInvalidVersion indicates an incompatible version
	ErrInvalidVersion = "ERR_2009_INVALID_VERSION"
	// ErrInvalidGraphID indicates an invalid graph ID format
	ErrInvalidGraphID = "ERR_2010_INVALID_GRAPH_ID"
)

// Error codes for state errors (3xxx)
const (
	// ErrSessionActive indicates a session is already active
	ErrSessionActive = "ERR_3001_SESSION_ALREADY_ACTIVE"
	// ErrSessionNotActive indicates no session is active
	ErrSessionNotActive = "ERR_3002_SESSION_NOT_ACTIVE"
	// ErrBranchLocked indicates a branch is locked for modification
	ErrBranchLocked = "ERR_3003_BRANCH_LOCKED"
	// ErrGraphFinalized indicates a graph is already finalized
	ErrGraphFinalized = "ERR_3004_GRAPH_ALREADY_FINALIZED"
	// ErrWorkflowRunning indicates a workflow is already running
	ErrWorkflowRunning = "ERR_3005_WORKFLOW_ALREADY_RUNNING"
	// ErrPresetExecuting indicates a preset is currently executing
	ErrPresetExecuting = "ERR_3006_PRESET_EXECUTING"
	// ErrAlreadyExists indicates the resource already exists
	ErrAlreadyExists = "ERR_3007_ALREADY_EXISTS"
	// ErrInvalidState indicates the resource is in an invalid state
	ErrInvalidState = "ERR_3008_INVALID_STATE"
)

// Error codes for external errors (4xxx)
const (
	// ErrEmbeddingFailed indicates embedding generation failed
	ErrEmbeddingFailed = "ERR_4001_EMBEDDING_FAILED"
	// ErrNeo4jConnection indicates Neo4j connection failed
	ErrNeo4jConnection = "ERR_4002_NEO4J_CONNECTION"
	// ErrLLMFailed indicates an LLM API call failed
	ErrLLMFailed = "ERR_4003_LLM_CALL_FAILED"
	// ErrStorageFailed indicates a storage operation failed
	ErrStorageFailed = "ERR_4004_STORAGE_OPERATION"
	// ErrNetworkFailed indicates a network operation failed
	ErrNetworkFailed = "ERR_4005_NETWORK_FAILED"
	// ErrAPIKeyMissing indicates a required API key is not configured
	ErrAPIKeyMissing = "ERR_4006_API_KEY_MISSING"
	// ErrExternalTimeout indicates an external service timed out
	ErrExternalTimeout = "ERR_4007_EXTERNAL_TIMEOUT"
)

// Error codes for limit errors (5xxx)
const (
	// ErrRateLimited indicates the rate limit has been exceeded
	ErrRateLimited = "ERR_5001_RATE_LIMITED"
	// ErrContextTooLarge indicates the context size exceeds limits
	ErrContextTooLarge = "ERR_5002_CONTEXT_TOO_LARGE"
	// ErrTooManyBranches indicates the branch limit has been reached
	ErrTooManyBranches = "ERR_5003_TOO_MANY_BRANCHES"
	// ErrMaxDepthReached indicates the maximum depth has been reached
	ErrMaxDepthReached = "ERR_5004_MAX_DEPTH_REACHED"
	// ErrMaxIterationsReached indicates the maximum iterations reached
	ErrMaxIterationsReached = "ERR_5005_MAX_ITERATIONS_REACHED"
	// ErrQuotaExceeded indicates a quota has been exceeded
	ErrQuotaExceeded = "ERR_5006_QUOTA_EXCEEDED"
)

// ErrorCategory returns the category name for an error code
func ErrorCategory(code string) string {
	if len(code) < 8 {
		return "unknown"
	}
	switch code[4] {
	case '1':
		return "resource"
	case '2':
		return "validation"
	case '3':
		return "state"
	case '4':
		return "external"
	case '5':
		return "limit"
	default:
		return "unknown"
	}
}

// IsRetryable returns whether an error is potentially retryable
func IsRetryable(code string) bool {
	// External errors and rate limits are often retryable
	category := ErrorCategory(code)
	return category == "external" || code == ErrRateLimited
}
