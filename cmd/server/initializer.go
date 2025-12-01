package main

import (
	"database/sql"
	"log"
	"os"

	"unified-thinking/internal/contextbridge"
	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/knowledge"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/server"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/validation"
)

// ServerComponents holds all initialized server components
type ServerComponents struct {
	Storage        storage.Storage
	LinearMode     *modes.LinearMode
	TreeMode       *modes.TreeMode
	DivergentMode  *modes.DivergentMode
	AutoMode       *modes.AutoMode
	Validator      *validation.LogicValidator
	Embedder       embeddings.Embedder
	ContextBridge  *contextbridge.ContextBridge
	KnowledgeGraph *knowledge.KnowledgeGraph
	Server         *server.UnifiedServer
	Orchestrator   *orchestration.Orchestrator
}

// InitializeServer creates and initializes all server components.
// This function is extracted from main() to enable testing.
func InitializeServer() (*ServerComponents, error) {
	components := &ServerComponents{}

	// Initialize storage
	store, err := storage.NewStorageFromEnv()
	if err != nil {
		return nil, err
	}
	components.Storage = store

	// Initialize thinking modes
	components.LinearMode = modes.NewLinearMode(store)
	components.TreeMode = modes.NewTreeMode(store)
	components.DivergentMode = modes.NewDivergentMode(store)
	components.AutoMode = modes.NewAutoMode(
		components.LinearMode,
		components.TreeMode,
		components.DivergentMode,
	)

	log.Println("Initialized thinking modes: linear, tree, divergent, auto")

	// Initialize validator
	components.Validator = validation.NewLogicValidator()
	log.Println("Initialized logic validator")

	// Initialize embedder if API key is available
	if apiKey := os.Getenv("VOYAGE_API_KEY"); apiKey != "" {
		model := os.Getenv("EMBEDDINGS_MODEL")
		if model == "" {
			model = "voyage-3-lite"
		}
		components.Embedder = embeddings.NewVoyageEmbedder(apiKey, model)
		components.AutoMode.SetEmbedder(components.Embedder)
		log.Printf("Initialized Voyage AI embedder (model: %s)", model)
	} else {
		log.Println("VOYAGE_API_KEY not set, semantic features disabled")
	}

	// Initialize Thompson Sampling RL if SQLite storage is available
	if sqliteStore, ok := store.(*storage.SQLiteStorage); ok {
		if err := components.AutoMode.SetRLStorage(sqliteStore); err != nil {
			log.Printf("Warning: failed to initialize Thompson Sampling RL: %v", err)
		}
	} else {
		log.Println("Thompson Sampling RL requires SQLite storage (using in-memory storage)")
	}

	// Initialize context bridge if SQLite storage is available
	bridgeConfig := contextbridge.ConfigFromEnv()
	if sqliteStore, ok := store.(*storage.SQLiteStorage); ok {
		if bridgeConfig.Enabled {
			components.ContextBridge = initializeContextBridge(
				sqliteStore,
				components.Embedder,
				bridgeConfig,
			)
			log.Println("Initialized context bridge with SQLite storage")
		} else {
			log.Println("Context bridge disabled via CONTEXT_BRIDGE_ENABLED=false")
		}
	} else {
		log.Println("Context bridge requires SQLite storage (using in-memory storage)")
	}

	// Create unified server
	components.Server = server.NewUnifiedServer(
		store,
		components.LinearMode,
		components.TreeMode,
		components.DivergentMode,
		components.AutoMode,
		components.Validator,
	)
	if components.ContextBridge != nil {
		components.Server.SetContextBridge(components.ContextBridge)
	}
	log.Println("Created unified server")

	// Initialize knowledge graph (if enabled)
	components.KnowledgeGraph = initializeKnowledgeGraph(store, components.Embedder)
	if components.KnowledgeGraph != nil && components.KnowledgeGraph.IsEnabled() {
		components.Server.SetKnowledgeGraph(components.KnowledgeGraph)
		log.Println("Knowledge graph enabled and configured")
	}

	// Initialize orchestrator
	executor := server.NewServerToolExecutor(components.Server)
	components.Orchestrator = orchestration.NewOrchestratorWithExecutor(executor)
	components.Server.SetOrchestrator(components.Orchestrator)
	log.Println("Initialized workflow orchestrator")

	// Register predefined workflows
	registerPredefinedWorkflows(components.Orchestrator)

	return components, nil
}

// initializeContextBridge creates and configures the context bridge
func initializeContextBridge(
	sqliteStore *storage.SQLiteStorage,
	embedder embeddings.Embedder,
	config *contextbridge.Config,
) *contextbridge.ContextBridge {
	adapter := contextbridge.NewStorageAdapter(sqliteStore)
	extractor := contextbridge.NewSimpleExtractor()

	// Use embedding-based similarity if embedder available, otherwise concept matching
	var similarity contextbridge.SimilarityCalculator
	if embedder != nil {
		fallback := contextbridge.NewDefaultSimilarity()
		similarity = contextbridge.NewEmbeddingSimilarity(embedder, fallback, true) // hybrid mode
		log.Println("Context bridge using embedding-based semantic similarity")
	} else {
		similarity = contextbridge.NewDefaultSimilarity()
		log.Println("Context bridge using concept-based similarity (no embedder)")
	}

	matcher := contextbridge.NewMatcher(adapter, similarity, extractor)
	return contextbridge.New(config, matcher, extractor, embedder)
}

// initializeKnowledgeGraph creates and configures the knowledge graph
func initializeKnowledgeGraph(store storage.Storage, embedder embeddings.Embedder) *knowledge.KnowledgeGraph {
	// Check if knowledge graph is enabled
	enabled := os.Getenv("NEO4J_ENABLED")
	if enabled != "true" {
		log.Println("Knowledge graph disabled (NEO4J_ENABLED != true)")
		return &knowledge.KnowledgeGraph{} // Return disabled instance
	}

	// Get SQLite database for embedding cache
	var sqliteDB *sql.DB
	if sqliteStore, ok := store.(*storage.SQLiteStorage); ok {
		sqliteDB = sqliteStore.DB()
	} else {
		log.Println("Knowledge graph requires SQLite storage (using in-memory storage)")
		return &knowledge.KnowledgeGraph{}
	}

	// Validate embedder
	if embedder == nil {
		log.Println("Knowledge graph requires embedder (VOYAGE_API_KEY not set)")
		return &knowledge.KnowledgeGraph{}
	}

	// Configure Neo4j
	neo4jCfg := knowledge.DefaultConfig()

	// Configure vector store with persistence
	vectorPersistPath := os.Getenv("VECTOR_STORE_PATH")
	if vectorPersistPath == "" {
		// Default: Use same directory as SQLite database
		if sqlitePath := os.Getenv("SQLITE_PATH"); sqlitePath != "" {
			// Replace .db extension with _vectors
			vectorPersistPath = sqlitePath[:len(sqlitePath)-3] + "_vectors"
		}
	}

	vectorCfg := knowledge.VectorStoreConfig{
		PersistPath: vectorPersistPath,
		Embedder:    embedder,
	}

	if vectorPersistPath != "" {
		log.Printf("Knowledge graph vector store persistence: %s", vectorPersistPath)
	} else {
		log.Println("Knowledge graph vector store using in-memory only (will not persist)")
	}

	// Create knowledge graph
	kgCfg := knowledge.KnowledgeGraphConfig{
		Neo4jConfig:  neo4jCfg,
		VectorConfig: vectorCfg,
		SQLiteDB:     sqliteDB,
		Enabled:      true,
	}

	kg, err := knowledge.NewKnowledgeGraph(kgCfg)
	if err != nil {
		log.Printf("Failed to initialize knowledge graph: %v", err)
		return &knowledge.KnowledgeGraph{}
	}

	return kg
}

// Cleanup closes all server resources
func (c *ServerComponents) Cleanup() error {
	if c.Storage != nil {
		return storage.CloseStorage(c.Storage)
	}
	return nil
}

//   - {{causal_graph.id}} or $causal_graph.id - reference nested field in previous result
//
// Resolution Order:
//  1. Reasoning context results (from previous steps)
//  2. Workflow input parameters
//  3. If unresolved, returns original template string for debugging
func registerPredefinedWorkflows(orchestrator *orchestration.Orchestrator) {
	// Guard against nil orchestrator
	if orchestrator == nil {
		log.Println("Warning: Cannot register workflows with nil orchestrator")
		return
	}

	// Register causal analysis workflow
	causalAnalysisWorkflow := &orchestration.Workflow{
		ID:          "causal-analysis",
		Name:        "Causal Analysis Pipeline",
		Description: "Complete causal analysis with fallacy detection",
		Type:        orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:   "build-graph",
				Tool: "build-causal-graph",
				Input: map[string]interface{}{
					"description":  "{{problem}}",
					"observations": "{{observations}}",
				},
				StoreAs: "causal_graph",
			},
			{
				ID:   "detect-issues",
				Tool: "detect-biases",
				Input: map[string]interface{}{
					"content": "{{problem}}",
				},
				DependsOn: []string{"build-graph"},
				StoreAs:   "detected_issues",
			},
			{
				ID:   "think-about-results",
				Tool: "think",
				Input: map[string]interface{}{
					"content": "Analyze the causal graph and detected issues for: {{problem}}",
					"mode":    "linear",
				},
				DependsOn: []string{"detect-issues"},
				StoreAs:   "analysis",
			},
		},
	}

	// Register critical thinking workflow
	criticalThinkingWorkflow := &orchestration.Workflow{
		ID:          "critical-thinking",
		Name:        "Critical Thinking Analysis",
		Description: "Comprehensive critical thinking pipeline with bias detection and formal validation",
		Type:        orchestration.WorkflowSequential,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:   "detect-biases",
				Tool: "detect-biases",
				Input: map[string]interface{}{
					"content": "{{content}}",
				},
				StoreAs: "detected_biases",
			},
			{
				ID:   "check-syntax",
				Tool: "check-syntax",
				Input: map[string]interface{}{
					"statement": "{{content}}",
				},
				DependsOn: []string{"detect-biases"},
				StoreAs:   "syntax_check",
			},
			{
				ID:   "prove",
				Tool: "prove",
				Input: map[string]interface{}{
					"premises":   "{{premises}}",
					"conclusion": "{{conclusion}}",
				},
				DependsOn: []string{"check-syntax"},
				Condition: &orchestration.StepCondition{
					Type:  "result_match",
					Field: "syntax_check.is_valid",
					Value: true,
				},
				StoreAs: "proof",
			},
		},
	}

	// Register multi-perspective decision workflow
	decisionMakingWorkflow := &orchestration.Workflow{
		ID:          "multi-perspective-decision",
		Name:        "Multi-Perspective Decision Making",
		Description: "Analyze decision from multiple perspectives and make balanced choice",
		Type:        orchestration.WorkflowParallel,
		Steps: []*orchestration.WorkflowStep{
			{
				ID:   "analyze-perspectives",
				Tool: "analyze-perspectives",
				Input: map[string]interface{}{
					"situation": "{{situation}}",
				},
				StoreAs: "perspectives",
			},
			{
				ID:   "sensitivity-analysis",
				Tool: "sensitivity-analysis",
				Input: map[string]interface{}{
					"target_claim": "{{decision}}",
					"assumptions":  "{{assumptions}}",
				},
				StoreAs: "sensitivity",
			},
			{
				ID:   "make-decision",
				Tool: "make-decision",
				Input: map[string]interface{}{
					"situation":    "{{situation}}",
					"criteria":     "{{criteria}}",
					"perspectives": "{{perspectives}}",
					"sensitivity":  "{{sensitivity}}",
				},
				DependsOn: []string{"analyze-perspectives", "sensitivity-analysis"},
			},
		},
	}

	// Register the predefined workflows
	if err := orchestrator.RegisterWorkflow(causalAnalysisWorkflow); err != nil {
		log.Printf("Failed to register causal-analysis workflow: %v", err)
	} else {
		log.Println("Registered workflow: causal-analysis")
	}

	if err := orchestrator.RegisterWorkflow(criticalThinkingWorkflow); err != nil {
		log.Printf("Failed to register critical-thinking workflow: %v", err)
	} else {
		log.Println("Registered workflow: critical-thinking")
	}

	if err := orchestrator.RegisterWorkflow(decisionMakingWorkflow); err != nil {
		log.Printf("Failed to register multi-perspective-decision workflow: %v", err)
	} else {
		log.Println("Registered workflow: multi-perspective-decision")
	}
}
