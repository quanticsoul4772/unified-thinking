package main

import (
	"fmt"
	"log"
	"os"

	"unified-thinking/internal/contextbridge"
	"unified-thinking/internal/embeddings"
	"unified-thinking/internal/knowledge"
	"unified-thinking/internal/modes"
	"unified-thinking/internal/orchestration"
	"unified-thinking/internal/server"
	"unified-thinking/internal/similarity"
	"unified-thinking/internal/storage"
	"unified-thinking/internal/types"
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
	Reranker       embeddings.Reranker
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

	// Initialize embedder and reranker - VOYAGE_API_KEY is REQUIRED
	apiKey := os.Getenv("VOYAGE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("VOYAGE_API_KEY not set: embeddings are required")
	}

	model := os.Getenv("EMBEDDINGS_MODEL")
	if model == "" {
		model = "voyage-3-lite"
	}
	components.Embedder = embeddings.NewVoyageEmbedder(apiKey, model)
	components.AutoMode.SetEmbedder(components.Embedder)
	log.Printf("Initialized Voyage AI embedder (model: %s)", model)

	// Initialize reranker - ALWAYS enabled (VOYAGE_API_KEY already validated above)
	rerankModel := os.Getenv("RERANK_MODEL")
	if rerankModel == "" {
		rerankModel = "rerank-2"
	}
	components.Reranker = embeddings.NewVoyageReranker(apiKey, rerankModel)
	log.Printf("Initialized Voyage AI reranker (model: %s)", rerankModel)

	// SQLite storage is REQUIRED for context bridge and knowledge graph
	sqliteStore, ok := store.(*storage.SQLiteStorage)
	if !ok {
		return nil, fmt.Errorf("SQLite storage is required (set STORAGE_TYPE=sqlite)")
	}

	// Initialize Thompson Sampling RL - ALWAYS enabled
	if err := components.AutoMode.SetRLStorage(sqliteStore); err != nil {
		return nil, fmt.Errorf("failed to initialize Thompson Sampling RL: %w", err)
	}

	// Initialize context bridge - ALWAYS enabled
	bridgeConfig := contextbridge.ConfigFromEnv()
	components.ContextBridge = initializeContextBridge(
		sqliteStore,
		components.Embedder,
		bridgeConfig,
	)
	log.Println("Initialized context bridge with SQLite storage")

	// Create unified server
	unifiedServer, err := server.NewUnifiedServer(
		store,
		components.LinearMode,
		components.TreeMode,
		components.DivergentMode,
		components.AutoMode,
		components.Validator,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create unified server: %w", err)
	}
	components.Server = unifiedServer
	components.Server.SetContextBridge(components.ContextBridge)
	log.Println("Created unified server")

	// Initialize knowledge graph - ALWAYS enabled, will FAIL if requirements not met
	kg, kgErr := initializeKnowledgeGraph(store, components.Embedder, components.Reranker)
	if kgErr != nil {
		return nil, kgErr
	}
	components.KnowledgeGraph = kg
	components.Server.SetKnowledgeGraph(components.KnowledgeGraph)
	log.Println("Knowledge graph enabled and configured")

	// Initialize thought similarity searcher - ALWAYS enabled (embedder is required)
	thoughtSearcher := similarity.NewThoughtSearcher(store, components.Embedder)
	thoughtSearcher.SetReranker(components.Reranker)
	log.Println("Thought similarity search enabled with reranking")
	components.Server.SetThoughtSearcher(thoughtSearcher)

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

// initializeKnowledgeGraph creates and configures the knowledge graph.
// Knowledge graph is ALWAYS enabled - will FAIL at runtime if requirements not met.
func initializeKnowledgeGraph(store storage.Storage, embedder embeddings.Embedder, reranker embeddings.Reranker) (*knowledge.KnowledgeGraph, error) {
	// Knowledge graph is ALWAYS enabled - no NEO4J_ENABLED check
	// FAIL FAST: All requirements must be met
	sqliteStore, ok := store.(*storage.SQLiteStorage)
	if !ok {
		return nil, fmt.Errorf("knowledge graph requires SQLite storage (set STORAGE_TYPE=sqlite)")
	}

	if embedder == nil {
		return nil, fmt.Errorf("knowledge graph requires embedder (set VOYAGE_API_KEY)")
	}

	sqliteDB := sqliteStore.DB()

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
	}

	// Create knowledge graph
	kgCfg := knowledge.KnowledgeGraphConfig{
		Neo4jConfig:  neo4jCfg,
		VectorConfig: vectorCfg,
		SQLiteDB:     sqliteDB,
	}

	kg, err := knowledge.NewKnowledgeGraph(kgCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize knowledge graph: %w", err)
	}

	// Set reranker for improved search results
	if reranker != nil {
		kg.SetReranker(reranker)
		log.Println("Knowledge graph reranking enabled")
	}

	return kg, nil
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
				Input: types.Metadata{
					"description":  "{{problem}}",
					"observations": "{{observations}}",
				},
				StoreAs: "causal_graph",
			},
			{
				ID:   "detect-issues",
				Tool: "detect-biases",
				Input: types.Metadata{
					"content": "{{problem}}",
				},
				DependsOn: []string{"build-graph"},
				StoreAs:   "detected_issues",
			},
			{
				ID:   "think-about-results",
				Tool: "think",
				Input: types.Metadata{
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
				Input: types.Metadata{
					"content": "{{content}}",
				},
				StoreAs: "detected_biases",
			},
			{
				ID:   "check-syntax",
				Tool: "check-syntax",
				Input: types.Metadata{
					"statement": "{{content}}",
				},
				DependsOn: []string{"detect-biases"},
				StoreAs:   "syntax_check",
			},
			{
				ID:   "prove",
				Tool: "prove",
				Input: types.Metadata{
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
				Input: types.Metadata{
					"situation": "{{situation}}",
				},
				StoreAs: "perspectives",
			},
			{
				ID:   "sensitivity-analysis",
				Tool: "sensitivity-analysis",
				Input: types.Metadata{
					"target_claim": "{{decision}}",
					"assumptions":  "{{assumptions}}",
				},
				StoreAs: "sensitivity",
			},
			{
				ID:   "make-decision",
				Tool: "make-decision",
				Input: types.Metadata{
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
