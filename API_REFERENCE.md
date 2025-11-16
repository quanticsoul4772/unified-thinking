# Unified Thinking Server - API Reference

Complete reference for all 50 MCP tools with parameters, examples, and response formats.

---

## Table of Contents

1. [Core Thinking Tools (11)](#core-thinking-tools)
2. [Probabilistic Reasoning Tools (4)](#probabilistic-reasoning-tools)
3. [Decision & Problem-Solving Tools (3)](#decision--problem-solving-tools)
4. [Metacognition Tools (3)](#metacognition-tools)
5. [Hallucination & Calibration Tools (4)](#hallucination--calibration-tools)
6. [Perspective & Temporal Analysis Tools (4)](#perspective--temporal-analysis-tools)
7. [Causal Reasoning Tools (5)](#causal-reasoning-tools)
8. [Integration & Orchestration Tools (6)](#integration--orchestration-tools)
9. [Dual-Process Reasoning Tools (1)](#dual-process-reasoning-tools)
10. [Backtracking Tools (3)](#backtracking-tools)
11. [Abductive Reasoning Tools (2)](#abductive-reasoning-tools)
12. [Case-Based Reasoning Tools (2)](#case-based-reasoning-tools)
13. [Symbolic Reasoning Tools (2)](#symbolic-reasoning-tools)

---

## Core Thinking Tools

### 1. think

Main thinking tool supporting multiple cognitive modes.

**Parameters:**
```typescript
{
  content: string;              // Required: The thought content
  mode: string;                 // Required: "auto" | "linear" | "tree" | "divergent" | "reflection" | "backtracking"
  confidence?: number;          // Optional: 0.0-1.0 (default: 0.8)
  key_points?: string[];        // Optional: Key observations
  branch_id?: string;           // Optional: For tree mode continuation
  parent_id?: string;           // Optional: Previous thought ID
  force_rebellion?: boolean;    // Optional: For divergent mode
  challenge_assumptions?: boolean; // Optional: Challenge existing assumptions
  require_validation?: boolean; // Optional: Force validation
}
```

**Request Example:**
```json
{
  "content": "Analyze the benefits of automated testing",
  "mode": "linear",
  "confidence": 0.8,
  "key_points": ["cost reduction", "bug detection", "maintainability"]
}
```

**Response Format:**
```json
{
  "status": "success",
  "thought_id": "thought-1763243373-1",
  "mode": "linear",
  "confidence": 0.8,
  "is_valid": true,
  "metadata": {
    "suggested_next_tools": [
      {
        "server_tool": "unified-thinking:think",
        "reason": "Continue linear reasoning chain",
        "priority": "recommended",
        "input_hint": "Set previous_thought_id to 'thought-1763243373-1'"
      }
    ],
    "validation_opportunities": ["Consider using web_search to gather evidence"],
    "action_recommendations": [
      {
        "type": "persist",
        "description": "Store this high-quality thought",
        "tool_chain": ["memory:create_entities", "obsidian:create-note"]
      }
    ]
  }
}
```

**Use Cases:**
- Linear mode: Step-by-step systematic reasoning
- Tree mode: Explore multiple solution paths
- Divergent mode: Creative brainstorming
- Auto mode: Automatic mode selection based on content

**Related Tools:** validate, history, search, synthesize-insights

---

### 2. history

View thinking history with optional filtering.

**Parameters:**
```typescript
{
  limit?: number;     // Optional: Max results (default: 10)
  offset?: number;    // Optional: Skip N results (default: 0)
  mode?: string;      // Optional: Filter by mode
  branch_id?: string; // Optional: Filter by branch
}
```

**Request Example:**
```json
{
  "limit": 5,
  "mode": "linear"
}
```

**Response Format:**
```json
{
  "thoughts": [
    {
      "id": "thought-1763243373-1",
      "content": "Analyze the benefits of automated testing",
      "mode": "linear",
      "confidence": 0.8,
      "timestamp": "2025-11-15T13:49:33.6011004-08:00",
      "is_rebellion": false,
      "challenges_assumption": false,
      "metadata": {
        "auto_validation_triggered": true,
        "auto_validation_scores": {
          "quality": 0.6,
          "completeness": 0.7,
          "coherence": 0.7
        }
      }
    }
  ]
}
```

**Use Cases:**
- Review reasoning progression
- Analyze thought patterns
- Track confidence evolution

**Related Tools:** search, list-branches

---

### 3. list-branches

List all thinking branches in tree mode.

**Parameters:**
```typescript
{} // No parameters required
```

**Response Format:**
```json
{
  "branches": [
    {
      "id": "branch-1763243379-1",
      "thought_count": 3,
      "insight_count": 2,
      "confidence": 0.75,
      "priority": 1.2,
      "created_at": "2025-11-15T13:49:39.6410025-08:00",
      "last_accessed": "2025-11-15T13:49:45.7749489-08:00"
    }
  ]
}
```

**Use Cases:**
- Overview of parallel exploration paths
- Identify high-priority branches
- Track branch exploration status

**Related Tools:** focus-branch, branch-history, recent-branches

---

### 4. focus-branch

Switch active branch in tree mode.

**Parameters:**
```typescript
{
  branch_id: string; // Required: Branch to activate
}
```

**Request Example:**
```json
{
  "branch_id": "branch-1763243379-1"
}
```

**Response Format:**
```json
{
  "status": "success",
  "active_branch": "branch-1763243379-1",
  "thought_count": 3,
  "insight_count": 2
}
```

**Use Cases:**
- Switch between exploration paths
- Resume work on different branches

**Related Tools:** list-branches, branch-history

---

### 5. branch-history

Get detailed history of a specific branch.

**Parameters:**
```typescript
{
  branch_id: string; // Required: Branch ID
}
```

**Request Example:**
```json
{
  "branch_id": "branch-1763243379-1"
}
```

**Response Format:**
```json
{
  "branch": {
    "id": "branch-1763243379-1",
    "thoughts": [...],
    "insights": [...],
    "cross_refs": [...],
    "confidence": 0.75,
    "priority": 1.2
  }
}
```

**Use Cases:**
- Deep dive into branch exploration
- Review insights and cross-references
- Analyze branch evolution

**Related Tools:** list-branches, focus-branch

---

### 6. recent-branches

Get recently accessed branches for quick context switching.

**Parameters:**
```typescript
{
  limit?: number; // Optional: Max results (default: 5)
}
```

**Response Format:**
```json
{
  "branches": [
    {
      "id": "branch-1763243379-1",
      "last_accessed": "2025-11-15T13:49:45.7749489-08:00",
      "thought_count": 3,
      "insight_count": 2
    }
  ]
}
```

**Use Cases:**
- Quick navigation to recent work
- Resume interrupted exploration

**Related Tools:** list-branches, focus-branch

---

### 7. validate

Validate thought for logical consistency.

**Parameters:**
```typescript
{
  thought_id: string; // Required: Thought to validate
}
```

**Request Example:**
```json
{
  "thought_id": "thought-1763243373-1"
}
```

**Response Format:**
```json
{
  "is_valid": true,
  "reason": "Thought is logically consistent",
  "issues": []
}
```

**Use Cases:**
- Check logical consistency
- Detect contradictions
- Quality assurance

**Related Tools:** detect-contradictions, detect-biases

---

### 8. prove

Attempt to prove a logical conclusion from premises.

**Parameters:**
```typescript
{
  premises: string[];  // Required: Array of premise statements
  conclusion: string;  // Required: Conclusion to prove
}
```

**Request Example:**
```json
{
  "premises": ["All humans are mortal", "Socrates is human"],
  "conclusion": "Socrates is mortal"
}
```

**Response Format:**
```json
{
  "is_valid": true,
  "proof_steps": [
    {
      "step": 1,
      "statement": "All humans are mortal",
      "rule": "premise"
    },
    {
      "step": 2,
      "statement": "Socrates is human",
      "rule": "premise"
    },
    {
      "step": 3,
      "statement": "Socrates is mortal",
      "rule": "modus_ponens",
      "from": [1, 2]
    }
  ]
}
```

**Use Cases:**
- Formal logical proofs
- Argument validation
- Deductive reasoning

**Related Tools:** check-syntax, validate, prove-theorem

---

### 9. check-syntax

Validate logical statement syntax.

**Parameters:**
```typescript
{
  statements: string[]; // Required: Statements to validate
}
```

**Request Example:**
```json
{
  "statements": ["P -> Q", "P", "Q"]
}
```

**Response Format:**
```json
{
  "is_valid": true,
  "errors": [],
  "warnings": []
}
```

**Use Cases:**
- Syntax validation before proof attempts
- Statement well-formedness checking

**Related Tools:** prove, validate

---

### 10. search

Search thoughts by query with optional mode filtering.

**Parameters:**
```typescript
{
  query: string;   // Required: Search query
  mode?: string;   // Optional: Filter by mode
  limit?: number;  // Optional: Max results (default: 10)
  offset?: number; // Optional: Skip N results (default: 0)
}
```

**Request Example:**
```json
{
  "query": "test coverage software quality",
  "limit": 10
}
```

**Response Format:**
```json
{
  "thoughts": [
    {
      "id": "thought-1763243373-1",
      "content": "Analyze the benefits of comprehensive test coverage...",
      "mode": "linear",
      "confidence": 0.8,
      "relevance_score": 0.95
    }
  ]
}
```

**Use Cases:**
- Find related thoughts
- Context retrieval
- Pattern discovery

**Related Tools:** history, list-branches

---

### 11. get-metrics

Get system performance and usage metrics.

**Parameters:**
```typescript
{} // No parameters required
```

**Response Format:**
```json
{
  "total_thoughts": 150,
  "total_branches": 12,
  "total_insights": 25,
  "storage_type": "memory",
  "average_confidence": 0.75,
  "mode_distribution": {
    "linear": 80,
    "tree": 45,
    "divergent": 25
  }
}
```

**Use Cases:**
- System monitoring
- Usage analytics
- Performance tracking

**Related Tools:** history, list-branches

---

## Probabilistic Reasoning Tools

### 12. probabilistic-reasoning

Bayesian inference and belief updates.

**Parameters:**
```typescript
{
  operation: string;      // Required: "create" | "update" | "get" | "combine"
  // For create:
  statement?: string;     // Statement to track
  prior_prob?: number;    // Prior probability (0-1)
  // For update:
  belief_id?: string;     // Belief to update
  evidence_id?: string;   // Evidence to incorporate
  likelihood?: number;    // P(E|H) - likelihood (0-1)
  evidence_prob?: number; // P(E) - evidence probability (0-1)
  // For combine:
  belief_ids?: string[];  // Beliefs to combine
  combine_op?: string;    // "and" | "or"
}
```

**Request Example (Create):**
```json
{
  "operation": "create",
  "statement": "Test coverage improvements will reduce bugs by 30%",
  "prior_prob": 0.6
}
```

**Response Format (Create):**
```json
{
  "status": "success",
  "operation": "create",
  "belief": {
    "id": "belief-1",
    "statement": "Test coverage improvements will reduce bugs by 30%",
    "prior_prob": 0.6,
    "probability": 0.6,
    "evidence": [],
    "updated_at": "2025-11-15T13:50:14.3454043-08:00"
  }
}
```

**Request Example (Update):**
```json
{
  "operation": "update",
  "belief_id": "belief-1",
  "evidence_id": "evidence-1",
  "likelihood": 0.8,
  "evidence_prob": 0.7
}
```

**Response Format (Update):**
```json
{
  "status": "success",
  "operation": "update",
  "belief": {
    "id": "belief-1",
    "statement": "Test coverage improvements will reduce bugs by 30%",
    "prior_prob": 0.6,
    "probability": 0.7059,
    "evidence": ["evidence-1"],
    "updated_at": "2025-11-15T13:50:19.6364272-08:00"
  }
}
```

**Use Cases:**
- Bayesian belief updating
- Probabilistic reasoning
- Confidence tracking with evidence

**Related Tools:** assess-evidence, evaluate-hypotheses

---

### 13. assess-evidence

Assess evidence quality and reliability.

**Parameters:**
```typescript
{
  content: string;          // Required: Evidence content
  source: string;           // Required: Evidence source
  supports_claim: boolean;  // Required: Does it support the claim?
  claim_id?: string;        // Optional: Associated claim ID
}
```

**Request Example:**
```json
{
  "content": "Research shows 25-40% reduction in defects with 80% coverage",
  "source": "Academic research papers on software testing",
  "supports_claim": true
}
```

**Response Format:**
```json
{
  "status": "success",
  "evidence": {
    "id": "evidence-1",
    "content": "Research shows 25-40% reduction in defects with 80% coverage",
    "source": "Academic research papers on software testing",
    "supports_claim": true,
    "relevance": 0.8,
    "reliability": 0.6,
    "overall_score": 0.69,
    "quality": "moderate",
    "created_at": "2025-11-15T13:50:26.1320176-08:00"
  }
}
```

**Use Cases:**
- Evidence quality assessment
- Source reliability evaluation
- Multi-dimensional scoring

**Related Tools:** probabilistic-reasoning, detect-contradictions

---

### 14. detect-contradictions

Find contradictions among thoughts or within branches.

**Parameters:**
```typescript
{
  thought_ids?: string[];  // Optional: Specific thoughts to check
  branch_id?: string;      // Optional: Check all thoughts in branch
  mode?: string;           // Optional: "explicit" | "implicit" | "both"
}
```

**Request Example:**
```json
{
  "thought_ids": ["thought-1", "thought-2", "thought-3"],
  "mode": "both"
}
```

**Response Format:**
```json
{
  "contradictions": [
    {
      "thought_1": "thought-1",
      "thought_2": "thought-2",
      "type": "explicit",
      "explanation": "Direct logical contradiction detected",
      "severity": "high"
    }
  ],
  "count": 1
}
```

**Use Cases:**
- Consistency checking
- Conflict detection
- Logical validation

**Related Tools:** validate, detect-biases

---

### 15. sensitivity-analysis

Test robustness of conclusions to assumption changes.

**Parameters:**
```typescript
{
  target_claim: string;      // Required: Claim to test
  assumptions: string[];     // Required: Underlying assumptions
  base_confidence: number;   // Required: Initial confidence (0-1)
}
```

**Request Example:**
```json
{
  "target_claim": "Test coverage above 80% significantly reduces bugs",
  "assumptions": [
    "Tests are well-designed",
    "Code review process is maintained",
    "Team follows best practices"
  ],
  "base_confidence": 0.8
}
```

**Response Format:**
```json
{
  "status": "success",
  "analysis": {
    "target_claim": "Test coverage above 80% significantly reduces bugs",
    "base_confidence": 0.8,
    "robustness_score": 0.65,
    "vulnerable_assumptions": [
      {
        "assumption": "Tests are well-designed",
        "impact": "high",
        "confidence_if_false": 0.3
      }
    ],
    "recommendation": "Claim is moderately robust but depends heavily on test quality"
  }
}
```

**Use Cases:**
- Robustness testing
- Assumption validation
- Risk assessment

**Related Tools:** detect-blind-spots, analyze-temporal

---

## Decision & Problem-Solving Tools

### 16. make-decision

Multi-criteria decision analysis with weighted scoring.

**Parameters:**
```typescript
{
  question: string;         // Required: Decision question
  options: Array<{          // Required: Decision options
    id: string;
    name: string;
    description: string;
    scores: Record<string, number>;  // Criterion scores (0-1)
    pros: string[];
    cons: string[];
    total_score: number;
  }>;
  criteria: Array<{         // Required: Decision criteria
    id: string;
    name: string;
    description: string;
    weight: number;         // Weight (0-1, sum to 1)
    maximize: boolean;      // Higher is better?
  }>;
}
```

**Request Example:**
```json
{
  "question": "Which database should we use?",
  "options": [
    {
      "id": "postgres",
      "name": "PostgreSQL",
      "description": "Open-source relational database",
      "scores": {"cost": 0.9, "performance": 0.8, "scalability": 0.7},
      "pros": ["Free", "Reliable", "Good ecosystem"],
      "cons": ["Complex setup", "Resource intensive"],
      "total_score": 0
    }
  ],
  "criteria": [
    {
      "id": "cost",
      "name": "Cost",
      "description": "Total cost of ownership",
      "weight": 0.4,
      "maximize": true
    },
    {
      "id": "performance",
      "name": "Performance",
      "description": "Query speed and throughput",
      "weight": 0.35,
      "maximize": true
    },
    {
      "id": "scalability",
      "name": "Scalability",
      "description": "Ability to handle growth",
      "weight": 0.25,
      "maximize": true
    }
  ]
}
```

**Response Format:**
```json
{
  "status": "success",
  "decision": {
    "question": "Which database should we use?",
    "recommendation": {
      "option_id": "postgres",
      "name": "PostgreSQL",
      "total_score": 0.815,
      "confidence": 0.85
    },
    "ranked_options": [
      {
        "option_id": "postgres",
        "name": "PostgreSQL",
        "total_score": 0.815,
        "breakdown": {
          "cost": 0.36,
          "performance": 0.28,
          "scalability": 0.175
        }
      }
    ],
    "metadata": {
      "export_formats": {
        "obsidian_note": "# Decision: Which database should we use?\n\n## Recommendation\n**PostgreSQL** (score: 0.815)\n\n..."
      }
    }
  }
}
```

**Use Cases:**
- Multi-criteria decision making
- Option comparison
- Weighted scoring analysis

**Related Tools:** decompose-problem, analyze-temporal, analyze-perspectives

---

### 17. decompose-problem

Break down complex problems into manageable subproblems.

**Parameters:**
```typescript
{
  problem: string; // Required: Problem statement
}
```

**Request Example:**
```json
{
  "problem": "How to improve CI/CD pipeline performance by 50%?"
}
```

**Response Format:**
```json
{
  "status": "success",
  "decomposition": {
    "problem": "How to improve CI/CD pipeline performance by 50%?",
    "subproblems": [
      {
        "id": "sub-1",
        "description": "Optimize build process",
        "complexity": "medium",
        "dependencies": []
      },
      {
        "id": "sub-2",
        "description": "Parallelize test execution",
        "complexity": "high",
        "dependencies": ["sub-1"]
      },
      {
        "id": "sub-3",
        "description": "Implement caching strategy",
        "complexity": "medium",
        "dependencies": []
      }
    ],
    "solution_path": ["sub-1", "sub-3", "sub-2"],
    "metadata": {
      "export_formats": {
        "obsidian_note": "# Problem: How to improve CI/CD pipeline performance\n\n## Subproblems\n- [ ] Optimize build process\n- [ ] Implement caching strategy\n- [ ] Parallelize test execution\n..."
      }
    }
  }
}
```

**Use Cases:**
- Complex problem solving
- Task breakdown
- Dependency analysis

**Related Tools:** make-decision, think (tree mode), synthesize-insights

---

### 18. verify-thought

Verify thought validity and structure.

**Parameters:**
```typescript
{
  thought_id: string; // Required: Thought to verify
}
```

**Response Format:**
```json
{
  "is_valid": true,
  "structure_check": "passed",
  "content_check": "passed",
  "metadata_check": "passed"
}
```

**Use Cases:**
- Thought structure validation
- Data integrity checks

**Related Tools:** validate, self-evaluate

---

## Metacognition Tools

### 19. self-evaluate

Metacognitive self-assessment of reasoning quality.

**Parameters:**
```typescript
{
  thought_id?: string;  // Optional: Specific thought
  branch_id?: string;   // Optional: Entire branch
}
```

**Request Example:**
```json
{
  "thought_id": "thought-1763243373-1"
}
```

**Response Format:**
```json
{
  "status": "success",
  "evaluation": {
    "id": "eval-3",
    "thought_id": "thought-1763243373-1",
    "quality_score": 0.5,
    "completeness_score": 0.5,
    "coherence_score": 0.7,
    "strengths": [
      "Clear logical structure",
      "Well-defined premises"
    ],
    "weaknesses": [
      "Lacks supporting evidence",
      "Could explore counter-arguments"
    ],
    "improvement_suggestions": [
      "Gather additional evidence",
      "Consider alternative perspectives"
    ],
    "created_at": "2025-11-15T13:51:08.4743765-08:00"
  }
}
```

**Use Cases:**
- Quality assessment
- Self-improvement
- Meta-reasoning

**Related Tools:** detect-biases, detect-blind-spots

---

### 20. detect-biases

Identify cognitive biases and logical fallacies.

**Parameters:**
```typescript
{
  thought_id?: string;  // Optional: Specific thought
  branch_id?: string;   // Optional: Entire branch
}
```

**Request Example:**
```json
{
  "thought_id": "thought-1763243373-1"
}
```

**Response Format:**
```json
{
  "status": "success",
  "biases": [
    {
      "type": "confirmation_bias",
      "severity": "medium",
      "explanation": "Selectively focusing on evidence that supports the claim",
      "location": "thought-1763243373-1"
    }
  ],
  "fallacies": [
    {
      "type": "hasty_generalization",
      "severity": "low",
      "explanation": "Drawing broad conclusion from limited evidence",
      "location": "thought-1763243373-1"
    }
  ],
  "combined": [
    /* Biases and fallacies merged */
  ],
  "count": 2
}
```

**Detected Biases (15+):**
- Confirmation bias
- Anchoring bias
- Availability heuristic
- Sunk cost fallacy
- Overconfidence bias
- Recency bias
- Groupthink

**Detected Fallacies (20+):**
- Ad hominem
- Straw man
- False dilemma
- Slippery slope
- Circular reasoning
- Hasty generalization
- Appeal to authority

**Use Cases:**
- Reasoning quality check
- Bias detection
- Argument validation

**Related Tools:** self-evaluate, validate, detect-blind-spots

---

### 21. detect-blind-spots

Identify unknown unknowns and knowledge gaps.

**Parameters:**
```typescript
{
  content: string;         // Required: Content to analyze
  domain?: string;         // Optional: Domain context
  context?: string;        // Optional: Additional context
  assumptions?: string[];  // Optional: Known assumptions
  confidence?: number;     // Optional: Confidence level (0-1)
}
```

**Request Example:**
```json
{
  "content": "Test coverage improves software quality by catching bugs early",
  "domain": "software engineering",
  "assumptions": ["Tests are run regularly", "Tests are well-designed"]
}
```

**Response Format:**
```json
{
  "status": "success",
  "overall_risk": 0.1,
  "risk_level": "LOW",
  "blind_spots": [],
  "missing_considerations": [],
  "unchallenged_assumptions": [],
  "suggested_questions": [
    "What assumptions are you making that might not hold?",
    "What factors haven't been considered?",
    "Who might disagree with this analysis and why?",
    "What would need to be true for this to be wrong?"
  ],
  "analysis": "Detected 0 potential blind spots.\n\nOverall risk level: 0.10\n✓ LOW RISK - Few blind spots detected, but remain vigilant."
}
```

**Use Cases:**
- Unknown unknowns detection
- Knowledge gap identification
- Assumption validation

**Related Tools:** detect-biases, self-evaluate, sensitivity-analysis

---

## Hallucination & Calibration Tools

### 22. get-hallucination-report

Retrieve hallucination detection reports for a thought.

**Parameters:**
```typescript
{
  thought_id: string; // Required: Thought ID
}
```

**Response Format:**
```json
{
  "status": "success",
  "report": {
    "thought_id": "thought-123",
    "overall_risk": 0.25,
    "semantic_uncertainty": {
      "aleatory": 0.15,
      "epistemic": 0.20,
      "confidence_mismatch": 0.10
    },
    "claims": [
      {
        "claim": "Test coverage above 80% guarantees zero bugs",
        "verified": false,
        "hallucination_likelihood": 0.8,
        "reason": "Unrealistic absolute claim"
      }
    ],
    "verified_count": 2,
    "hallucination_count": 1,
    "recommendations": [
      "Review absolute claims",
      "Gather supporting evidence"
    ]
  }
}
```

**Use Cases:**
- Hallucination detection
- Claim verification
- Quality assurance

**Related Tools:** verify-thought, assess-evidence

---

### 23. record-prediction

Record a confidence prediction for calibration tracking.

**Parameters:**
```typescript
{
  thought_id: string;    // Required: Thought ID
  confidence: number;    // Required: Confidence (0-1)
  mode: string;          // Required: Thinking mode
  metadata?: object;     // Optional: Additional metadata
}
```

**Request Example:**
```json
{
  "thought_id": "thought-123",
  "confidence": 0.8,
  "mode": "linear"
}
```

**Response Format:**
```json
{
  "status": "success",
  "prediction_id": "pred-1",
  "recorded_at": "2025-11-15T14:00:00Z"
}
```

**Use Cases:**
- Confidence tracking
- Calibration analysis
- Quality metrics

**Related Tools:** record-outcome, get-calibration-report

---

### 24. record-outcome

Record the actual outcome of a prediction.

**Parameters:**
```typescript
{
  thought_id: string;          // Required: Thought ID
  was_correct: boolean;        // Required: Was the thought correct?
  actual_confidence: number;   // Required: Actual confidence (0-1)
  source: string;              // Required: "validation" | "verification" | "user_feedback"
  metadata?: object;           // Optional: Additional context
}
```

**Request Example:**
```json
{
  "thought_id": "thought-123",
  "was_correct": true,
  "actual_confidence": 0.9,
  "source": "validation"
}
```

**Response Format:**
```json
{
  "status": "success",
  "outcome_id": "outcome-1",
  "recorded_at": "2025-11-15T14:00:00Z"
}
```

**Use Cases:**
- Outcome tracking
- Calibration validation
- Performance metrics

**Related Tools:** record-prediction, get-calibration-report

---

### 25. get-calibration-report

Retrieve comprehensive confidence calibration analysis.

**Parameters:**
```typescript
{} // No parameters required
```

**Response Format:**
```json
{
  "status": "success",
  "total_predictions": 150,
  "total_outcomes": 120,
  "buckets": [
    {
      "range": "0.8-0.9",
      "predicted_confidence": 0.85,
      "actual_accuracy": 0.78,
      "count": 45
    }
  ],
  "overall_accuracy": 0.75,
  "calibration": 0.08,
  "bias": "slight_overconfidence",
  "by_mode": {
    "linear": {"accuracy": 0.80, "calibration": 0.05},
    "tree": {"accuracy": 0.72, "calibration": 0.10}
  },
  "recommendations": [
    "Reduce confidence estimates by 5-10% in tree mode",
    "Linear mode is well-calibrated"
  ]
}
```

**Use Cases:**
- Calibration analysis
- Confidence adjustment
- Performance monitoring

**Related Tools:** record-prediction, record-outcome

---

## Perspective & Temporal Analysis Tools

### 26. analyze-perspectives

Multi-stakeholder perspective analysis.

**Parameters:**
```typescript
{
  situation: string;           // Required: Situation to analyze
  stakeholder_hints?: string[]; // Optional: Stakeholder types
}
```

**Request Example:**
```json
{
  "situation": "Implementing new authentication system",
  "stakeholder_hints": ["developers", "security team", "users", "management"]
}
```

**Response Format:**
```json
{
  "status": "success",
  "perspectives": [
    {
      "stakeholder": "developers",
      "viewpoint": "Focus on implementation complexity and maintainability",
      "concerns": ["Development time", "Code complexity", "Testing burden"],
      "priorities": ["Clean architecture", "Good documentation", "Automated tests"]
    },
    {
      "stakeholder": "security team",
      "viewpoint": "Prioritize security and compliance",
      "concerns": ["Vulnerability risks", "Compliance requirements", "Audit trails"],
      "priorities": ["Strong encryption", "MFA support", "Security auditing"]
    }
  ],
  "metadata": {
    "suggested_next_tools": [
      {
        "server_tool": "unified-thinking:synthesize-insights",
        "reason": "Find common ground across perspectives"
      }
    ],
    "export_formats": {
      "memory_entities": [
        {
          "entity_type": "stakeholder",
          "name": "developers",
          "observations": ["Focus on implementation complexity"]
        }
      ]
    }
  }
}
```

**Use Cases:**
- Stakeholder analysis
- Conflict identification
- Consensus building

**Related Tools:** make-decision, synthesize-insights

---

### 27. analyze-temporal

Analyze short-term vs long-term implications.

**Parameters:**
```typescript
{
  situation: string;      // Required: Decision or situation
  time_horizon?: string;  // Optional: "days-weeks" | "months" | "years"
}
```

**Request Example:**
```json
{
  "situation": "Refactor codebase now or after release?",
  "time_horizon": "months"
}
```

**Response Format:**
```json
{
  "status": "success",
  "analysis": {
    "short_term_view": "Refactoring now delays release but reduces technical debt",
    "medium_term_view": "Clean codebase enables faster feature development",
    "long_term_view": "Maintainability benefits compound over time",
    "tradeoffs": [
      {
        "decision": "refactor_now",
        "short_term": "Release delay, team disruption",
        "long_term": "Lower maintenance costs, happier developers"
      }
    ],
    "recommendation": "Refactor now if technical debt is high and release can be delayed by 2-3 weeks",
    "confidence": 0.75
  },
  "metadata": {
    "export_formats": {
      "obsidian_note": "# Temporal Analysis: Refactor now or later?\n\n## Short-term (days-weeks)\n...\n\n## Long-term (years)\n..."
    }
  }
}
```

**Use Cases:**
- Temporal tradeoff analysis
- Time horizon comparison
- Strategic planning

**Related Tools:** compare-time-horizons, identify-optimal-timing, make-decision

---

### 28. compare-time-horizons

Compare decision across multiple time horizons.

**Parameters:**
```typescript
{
  situation: string; // Required: Situation to analyze
}
```

**Response Format:**
```json
{
  "status": "success",
  "comparison": {
    "days_weeks": {
      "view": "Immediate impact assessment",
      "priority": "high"
    },
    "months": {
      "view": "Medium-term strategic considerations",
      "priority": "medium"
    },
    "years": {
      "view": "Long-term architectural implications",
      "priority": "low"
    }
  }
}
```

**Use Cases:**
- Multi-horizon analysis
- Strategic planning
- Priority assessment

**Related Tools:** analyze-temporal, identify-optimal-timing

---

### 29. identify-optimal-timing

Determine optimal timing for a decision.

**Parameters:**
```typescript
{
  situation: string;       // Required: Situation
  constraints?: string[];  // Optional: Timing constraints
}
```

**Request Example:**
```json
{
  "situation": "When to launch new product feature?",
  "constraints": ["Market window closes in Q2", "Development complete in 6 weeks"]
}
```

**Response Format:**
```json
{
  "status": "success",
  "timing": {
    "recommended_timeframe": "6-8 weeks from now",
    "rationale": "Balances development readiness with market opportunity",
    "risks": [
      {
        "if_earlier": "Incomplete feature, poor user experience",
        "if_later": "Missed market window, competitor advantage"
      }
    ],
    "confidence": 0.7
  }
}
```

**Use Cases:**
- Timing optimization
- Risk analysis
- Strategic scheduling

**Related Tools:** analyze-temporal, compare-time-horizons

---

## Causal Reasoning Tools

### 30. build-causal-graph

Construct causal graphs from observations.

**Parameters:**
```typescript
{
  description: string;    // Required: Graph description
  observations: string[]; // Required: Causal observations (format: "X causes Y")
}
```

**Request Example:**
```json
{
  "description": "Software quality and testing relationship",
  "observations": [
    "Increased test coverage causes better code quality",
    "Better code quality causes fewer production bugs",
    "Fewer production bugs causes higher customer satisfaction"
  ]
}
```

**Response Format:**
```json
{
  "status": "success",
  "graph": {
    "id": "causal-graph-1",
    "description": "Software quality and testing relationship",
    "variables": [
      {
        "id": "var-1",
        "name": "increased test coverage",
        "type": "continuous",
        "observable": true
      },
      {
        "id": "var-2",
        "name": "better code quality",
        "type": "continuous",
        "observable": true
      },
      {
        "id": "var-3",
        "name": "fewer production bugs",
        "type": "continuous",
        "observable": true
      }
    ],
    "links": [
      {
        "id": "link-1",
        "from": "var-1",
        "to": "var-2",
        "type": "positive",
        "strength": 0.7,
        "confidence": 0.7,
        "evidence": ["increased test coverage causes better code quality"]
      }
    ],
    "created_at": "2025-11-15T13:50:41.4287959-08:00"
  },
  "metadata": {
    "suggested_next_tools": [
      {
        "server_tool": "unified-thinking:simulate-intervention",
        "input_hint": "Use graph_id 'causal-graph-1' to simulate interventions",
        "reason": "Test interventions on this causal model"
      },
      {
        "server_tool": "memory:create_entities",
        "input_hint": "Use export_formats.memory_entities from this response",
        "reason": "Persist causal model in knowledge graph"
      }
    ],
    "export_formats": {
      "memory_entities": [
        {
          "entity_type": "causal_variable",
          "name": "increased test coverage",
          "observations": ["Type: continuous"]
        }
      ],
      "memory_relations": [
        {
          "from": "increased test coverage",
          "to": "better code quality",
          "relation_type": "causes_positive"
        }
      ]
    }
  }
}
```

**Use Cases:**
- Causal modeling
- System understanding
- Intervention planning

**Related Tools:** simulate-intervention, generate-counterfactual, analyze-correlation-vs-causation

---

### 31. simulate-intervention

Simulate interventions using do-calculus.

**Parameters:**
```typescript
{
  graph_id: string;          // Required: Causal graph ID
  variable_id: string;       // Required: Variable to intervene on
  intervention_type: string; // Required: "increase" | "decrease" | "set"
  value?: number;            // Optional: For "set" intervention
}
```

**Request Example:**
```json
{
  "graph_id": "causal-graph-1",
  "variable_id": "var-1",
  "intervention_type": "increase"
}
```

**Response Format:**
```json
{
  "status": "success",
  "intervention": {
    "id": "intervention-2",
    "graph_id": "causal-graph-1",
    "variable": "increased test coverage",
    "intervention_type": "increase",
    "confidence": 0.299,
    "metadata": {
      "graph_surgery_applied": true,
      "intervention_note": "Applied Pearl's do-calculus: removed incoming edges to intervention variable"
    },
    "predicted_effects": [
      {
        "variable": "better code quality",
        "effect": "increase",
        "magnitude": 0.7,
        "path_length": 1,
        "probability": 0.49,
        "explanation": "Via positive causal link from intervention variable"
      },
      {
        "variable": "fewer production bugs",
        "effect": "increase",
        "magnitude": 0.7,
        "path_length": 2,
        "probability": 0.49,
        "explanation": "Via positive causal link from intervention variable"
      }
    ],
    "created_at": "2025-11-15T13:50:48.0860705-08:00"
  }
}
```

**Use Cases:**
- Intervention simulation
- Causal effect prediction
- What-if analysis

**Related Tools:** build-causal-graph, generate-counterfactual

---

### 32. generate-counterfactual

Generate "what if" counterfactual scenarios.

**Parameters:**
```typescript
{
  graph_id: string;               // Required: Causal graph ID
  scenario: string;               // Required: Scenario description
  changes: Record<string, string>; // Required: Variable changes
}
```

**Request Example:**
```json
{
  "graph_id": "causal-graph-1",
  "scenario": "What if test coverage had not been increased?",
  "changes": {
    "var-1": "decreased"
  }
}
```

**Response Format:**
```json
{
  "status": "success",
  "counterfactual": {
    "id": "counterfactual-3",
    "graph_id": "causal-graph-1",
    "scenario": "What if test coverage had not been increased?",
    "changes": {"var-1": "decreased"},
    "outcomes": {},
    "plausibility": 0.7,
    "created_at": "2025-11-15T13:50:54.6339452-08:00"
  }
}
```

**Use Cases:**
- Counterfactual reasoning
- Alternative scenario analysis
- Causal understanding

**Related Tools:** build-causal-graph, simulate-intervention

---

### 33. analyze-correlation-vs-causation

Distinguish correlation from causation.

**Parameters:**
```typescript
{
  observation: string; // Required: Observed relationship
}
```

**Request Example:**
```json
{
  "observation": "Ice cream sales and drowning deaths both increase in summer"
}
```

**Response Format:**
```json
{
  "status": "success",
  "analysis": {
    "observation": "Ice cream sales and drowning deaths both increase in summer",
    "likely_relationship": "correlation",
    "explanation": "Both caused by a common factor (warm weather)",
    "confounding_factors": ["temperature", "outdoor activity"],
    "causal_evidence": "none",
    "recommendation": "Do not infer causation without controlling for confounders"
  }
}
```

**Use Cases:**
- Correlation vs causation analysis
- Confounding factor identification
- Causal inference validation

**Related Tools:** build-causal-graph, assess-evidence

---

### 34. get-causal-graph

Retrieve previously built causal graph.

**Parameters:**
```typescript
{
  graph_id: string; // Required: Graph ID
}
```

**Response Format:**
```json
{
  "status": "success",
  "graph": {
    "id": "causal-graph-1",
    "description": "Software quality and testing relationship",
    "variables": [...],
    "links": [...],
    "created_at": "2025-11-15T13:50:41.4287959-08:00"
  }
}
```

**Use Cases:**
- Graph retrieval
- State inspection
- Analysis continuation

**Related Tools:** build-causal-graph, simulate-intervention

---

## Integration & Orchestration Tools

### 35. synthesize-insights

Synthesize insights from multiple reasoning modes.

**Parameters:**
```typescript
{
  context: string;        // Required: Integration context
  inputs: Array<{         // Required: Inputs from different modes
    ID: string;
    Mode: string;
    Content: string;
    Confidence: number;
    Metadata: object;
  }>;
}
```

**Request Example:**
```json
{
  "context": "Software quality improvement through testing",
  "inputs": [
    {
      "ID": "thought-1",
      "Mode": "linear",
      "Content": "Test coverage improvements benefit software quality",
      "Confidence": 0.8,
      "Metadata": {}
    },
    {
      "ID": "thought-2",
      "Mode": "tree",
      "Content": "Multiple approaches to quality: testing, reviews, static analysis",
      "Confidence": 0.8,
      "Metadata": {}
    }
  ]
}
```

**Response Format:**
```json
{
  "status": "success",
  "synthesis": {
    "id": "synthesis-1",
    "confidence": 0.85,
    "sources": ["thought-1", "thought-2"],
    "integrated_view": "Integrated analysis of: Software quality improvement through testing\n\nKey Insights:\n1. [linear mode] Test coverage improvements benefit software quality (confidence: 0.80)\n2. [tree mode] Multiple approaches to quality: testing, reviews, static analysis (confidence: 0.80)\n\nComplementary Insights:\n- Multiple reasoning modes (linear, tree) provide complementary lenses on the situation\n\nSynthesized Conclusion:\nHigh confidence synthesis: Analysis across 2 modes (linear, tree) provides complementary insights that strengthen understanding.",
    "synergies": [
      "Multiple reasoning modes (linear, tree) provide complementary lenses on the situation"
    ],
    "conflicts": [],
    "created_at": "2025-11-15T13:52:15.8065558-08:00",
    "metadata": {
      "context": "Software quality improvement through testing",
      "modes": ["linear", "tree"]
    }
  }
}
```

**Use Cases:**
- Cross-mode integration
- Insight synthesis
- Holistic analysis

**Related Tools:** detect-emergent-patterns, think (multiple modes)

---

### 36. detect-emergent-patterns

Detect emergent patterns across reasoning modes.

**Parameters:**
```typescript
{
  inputs: Array<{         // Required: Inputs from different modes
    ID: string;
    Mode: string;
    Content: string;
    Confidence: number;
    Metadata: object;
  }>;
}
```

**Request Example:**
```json
{
  "inputs": [
    {
      "ID": "thought-1",
      "Mode": "linear",
      "Content": "Sequential analysis of test benefits",
      "Confidence": 0.8,
      "Metadata": {}
    },
    {
      "ID": "thought-2",
      "Mode": "probabilistic",
      "Content": "Bayesian update shows 70% confidence in benefit",
      "Confidence": 0.7,
      "Metadata": {}
    },
    {
      "ID": "thought-3",
      "Mode": "causal",
      "Content": "Causal chain from testing to quality",
      "Confidence": 0.75,
      "Metadata": {}
    }
  ]
}
```

**Response Format:**
```json
{
  "status": "success",
  "count": 1,
  "patterns": [
    "Cross-mode analysis reveals interconnected factors requiring holistic consideration"
  ]
}
```

**Use Cases:**
- Pattern discovery
- Emergent insight detection
- Multi-modal analysis

**Related Tools:** synthesize-insights, think (multiple modes)

---

### 37. execute-workflow

Execute predefined multi-tool workflows.

**Parameters:**
```typescript
{
  workflow_id: string;  // Required: Workflow identifier
  input: object;        // Required: Workflow-specific input
}
```

**Request Example:**
```json
{
  "workflow_id": "causal-analysis",
  "input": {
    "problem": "Understanding test coverage impact",
    "observations": [
      "Test coverage increased",
      "Bug rate decreased"
    ]
  }
}
```

**Response Format:**
```json
{
  "status": "success",
  "result": {
    "causal_graph": {...},
    "detected_issues": {...},
    "analysis": {...}
  }
}
```

**Use Cases:**
- Automated workflows
- Multi-tool coordination
- Complex analysis pipelines

**Related Tools:** list-workflows, register-workflow

---

### 38. list-workflows

List available automated workflows.

**Parameters:**
```typescript
{} // No parameters required
```

**Response Format:**
```json
{
  "count": 3,
  "workflows": [
    {
      "id": "causal-analysis",
      "name": "Causal Analysis Pipeline",
      "description": "Complete causal analysis with fallacy detection",
      "type": "sequential",
      "steps": [
        {
          "id": "build-graph",
          "tool": "build-causal-graph",
          "input": {"description": "{{problem}}", "observations": "{{observations}}"},
          "store_as": "causal_graph"
        }
      ]
    }
  ]
}
```

**Use Cases:**
- Workflow discovery
- Available pipeline listing
- Automation options

**Related Tools:** execute-workflow, register-workflow

---

### 39. register-workflow

Register custom workflows for automation.

**Parameters:**
```typescript
{
  workflow: {
    id: string;
    name: string;
    description: string;
    type: "sequential" | "parallel";
    steps: Array<{
      id: string;
      tool: string;
      input: object;
      depends_on?: string[];
      store_as?: string;
      condition?: object;
    }>;
  };
}
```

**Use Cases:**
- Custom workflow creation
- Automation setup
- Reusable pipelines

**Related Tools:** execute-workflow, list-workflows

---

### 40. list-integration-patterns

Discover multi-server integration patterns.

**Parameters:**
```typescript
{} // No parameters required
```

**Response Format:**
```json
{
  "status": "success",
  "count": 10,
  "patterns": [
    {
      "name": "Research-Enhanced Thinking",
      "description": "Gather external evidence before reasoning, then validate conclusions",
      "servers": ["brave-search", "unified-thinking"],
      "steps": [
        "1. brave-search:brave_web_search - Search for relevant information",
        "2. unified-thinking:think - Reason with gathered context",
        "3. unified-thinking:assess-evidence - Validate evidence quality",
        "4. unified-thinking:synthesize-insights - Combine findings"
      ],
      "use_case": "When reasoning about topics requiring external validation or current information"
    },
    {
      "name": "Knowledge-Backed Decision Making",
      "description": "Query existing knowledge before making decisions, then document results",
      "servers": ["memory", "conversation", "unified-thinking", "obsidian"],
      "steps": [
        "1. memory:traverse_graph - Find related concepts from knowledge base",
        "2. conversation:conversation_search - Check past discussions",
        "3. unified-thinking:make-decision - Decide with full context",
        "4. memory:create_entities - Store decision rationale",
        "5. obsidian:create-note - Document decision for future reference"
      ],
      "use_case": "Important decisions that benefit from organizational memory and history"
    }
  ]
}
```

**Available Patterns:**
1. Research-Enhanced Thinking
2. Knowledge-Backed Decision Making
3. Causal Model to Knowledge Graph
4. Problem Decomposition Workflow
5. Temporal Decision Analysis
6. Stakeholder-Aware Planning
7. Validated File Operations
8. Evidence-Based Causal Reasoning
9. Iterative Problem Refinement
10. Knowledge Discovery Pipeline

**Use Cases:**
- Integration pattern discovery
- Multi-server workflow guidance
- Best practices reference

**Related Tools:** execute-workflow, synthesize-insights

---

## Dual-Process Reasoning Tools

### 41. dual-process-think

System 1 (fast intuitive) vs System 2 (slow deliberate) reasoning.

**Parameters:**
```typescript
{
  content: string;          // Required: Problem to analyze
  mode?: string;            // Optional: "auto" | "linear" | "tree"
  force_system?: string;    // Optional: "system1" | "system2"
  key_points?: string[];    // Optional: Key observations
  metadata?: object;        // Optional: Additional metadata
}
```

**Request Example:**
```json
{
  "content": "Should we invest in increasing test coverage from 75% to 90%?",
  "mode": "auto"
}
```

**Response Format:**
```json
{
  "status": "success",
  "thought_id": "thought-1763243489-4",
  "content": "Should we invest in increasing test coverage from 75% to 90%?",
  "system_used": "system1",
  "complexity": 0.1,
  "confidence": 0.8,
  "escalated": false,
  "system1_time": "0s",
  "metadata": {
    "escalation_available": true,
    "processing_mode": "fast_heuristic",
    "processing_system": "System1"
  }
}
```

**System Selection:**
- **System 1** (Fast): Low complexity, pattern recognition, quick decisions
- **System 2** (Slow): High complexity, analytical reasoning, detailed analysis
- **Auto-escalation**: System 1 can escalate to System 2 when needed

**Use Cases:**
- Quick intuitive decisions
- Complex analytical reasoning
- Adaptive processing

**Related Tools:** think, self-evaluate

---

## Backtracking Tools

### 42. create-checkpoint

Create reasoning checkpoint for backtracking.

**Parameters:**
```typescript
{
  branch_id: string;     // Required: Branch to checkpoint
  name: string;          // Required: Checkpoint name
  description?: string;  // Optional: Checkpoint description
}
```

**Request Example:**
```json
{
  "branch_id": "branch-123",
  "name": "Before hypothesis testing",
  "description": "Checkpoint before exploring alternative hypotheses"
}
```

**Response Format:**
```json
{
  "status": "success",
  "checkpoint_id": "checkpoint-1",
  "branch_id": "branch-123",
  "thought_count": 5,
  "insight_count": 2,
  "created_at": "2025-11-15T14:00:00Z"
}
```

**Use Cases:**
- Save reasoning state
- Enable backtracking
- Exploration safety net

**Related Tools:** restore-checkpoint, list-checkpoints

---

### 43. restore-checkpoint

Restore branch from a checkpoint.

**Parameters:**
```typescript
{
  checkpoint_id: string; // Required: Checkpoint to restore
}
```

**Request Example:**
```json
{
  "checkpoint_id": "checkpoint-1"
}
```

**Response Format:**
```json
{
  "status": "success",
  "branch_id": "branch-123",
  "thought_count": 5,
  "insight_count": 2,
  "message": "Branch restored to checkpoint 'Before hypothesis testing'"
}
```

**Use Cases:**
- Undo reasoning path
- Try alternative approaches
- Safe exploration

**Related Tools:** create-checkpoint, list-checkpoints

---

### 44. list-checkpoints

List available checkpoints.

**Parameters:**
```typescript
{
  branch_id?: string; // Optional: Filter by branch
}
```

**Response Format:**
```json
{
  "checkpoints": [
    {
      "id": "checkpoint-1",
      "name": "Before hypothesis testing",
      "description": "Checkpoint before exploring alternative hypotheses",
      "branch_id": "branch-123",
      "thought_count": 5,
      "created_at": "2025-11-15T14:00:00Z"
    }
  ]
}
```

**Use Cases:**
- Checkpoint discovery
- Backtracking options
- State management

**Related Tools:** create-checkpoint, restore-checkpoint

---

## Abductive Reasoning Tools

### 45. generate-hypotheses

Generate explanatory hypotheses from observations.

**Parameters:**
```typescript
{
  observations: Array<{     // Required: Observations to explain
    description: string;
    confidence?: number;    // Optional: 0-1
  }>;
  max_hypotheses?: number;  // Optional: Max to generate (default: 5)
  min_parsimony?: number;   // Optional: Min parsimony score (default: 0.5)
}
```

**Request Example:**
```json
{
  "observations": [
    {"description": "Production bug rate decreased by 35%", "confidence": 0.9},
    {"description": "Test coverage increased from 51% to 81%", "confidence": 1.0},
    {"description": "Code review process was also improved", "confidence": 0.8}
  ],
  "max_hypotheses": 5
}
```

**Response Format:**
```json
{
  "status": "success",
  "count": 3,
  "hypotheses": [
    {
      "id": "hyp-1",
      "description": "Increased test coverage caused bug reduction",
      "parsimony": 0.9,
      "prior_probability": 0.7,
      "observations": ["obs-1", "obs-2"],
      "assumptions": ["Tests are effective at finding bugs"]
    },
    {
      "id": "hyp-2",
      "description": "Improved code review process caused bug reduction",
      "parsimony": 0.85,
      "prior_probability": 0.6,
      "observations": ["obs-1", "obs-3"],
      "assumptions": ["Code reviews catch bugs before production"]
    },
    {
      "id": "hyp-3",
      "description": "Combined effect of testing and reviews",
      "parsimony": 0.7,
      "prior_probability": 0.8,
      "observations": ["obs-1", "obs-2", "obs-3"],
      "assumptions": ["Both testing and reviews contribute"]
    }
  ]
}
```

**Use Cases:**
- Hypothesis generation
- Explanatory reasoning
- Root cause analysis

**Related Tools:** evaluate-hypotheses, assess-evidence

---

### 46. evaluate-hypotheses

Evaluate hypothesis plausibility using Bayesian inference or parsimony.

**Parameters:**
```typescript
{
  observations: Array<{     // Required: Observations
    description: string;
    confidence?: number;
  }>;
  hypotheses: Array<{       // Required: Hypotheses to evaluate
    description: string;
    observations: string[];
    prior_probability?: number;
    assumptions?: string[];
  }>;
  method?: string;          // Optional: "bayesian" | "parsimony" | "combined"
}
```

**Request Example:**
```json
{
  "observations": [
    {"description": "Production bug rate decreased by 35%", "confidence": 0.9},
    {"description": "Test coverage increased from 51% to 81%", "confidence": 1.0}
  ],
  "hypotheses": [
    {
      "description": "Increased test coverage caused bug reduction",
      "observations": ["Production bug rate decreased by 35%", "Test coverage increased from 51% to 81%"],
      "prior_probability": 0.7
    }
  ],
  "method": "bayesian"
}
```

**Response Format:**
```json
{
  "status": "success",
  "method": "bayesian",
  "best_hypothesis": {
    "description": "Increased test coverage caused bug reduction",
    "rank": 1,
    "posterior_probability": 0.82,
    "explanatory_power": 0.9,
    "parsimony": 0.85
  },
  "ranked_hypotheses": [
    {
      "description": "Increased test coverage caused bug reduction",
      "rank": 1,
      "posterior_probability": 0.82,
      "explanatory_power": 0.9,
      "parsimony": 0.85
    }
  ]
}
```

**Use Cases:**
- Hypothesis evaluation
- Best explanation selection
- Abductive inference

**Related Tools:** generate-hypotheses, probabilistic-reasoning

---

## Case-Based Reasoning Tools

### 47. retrieve-similar-cases

Retrieve similar cases from case library using similarity matching.

**Parameters:**
```typescript
{
  problem: {                  // Required: Problem to match
    description: string;
    context?: string;
    goals?: string[];
    constraints?: string[];
    features?: object;
  };
  domain?: string;            // Optional: Domain filter
  max_cases?: number;         // Optional: Max results (default: 5)
  min_similarity?: number;    // Optional: Min similarity (default: 0.5)
}
```

**Request Example:**
```json
{
  "problem": {
    "description": "Need to improve test coverage from 50% to 80%",
    "context": "Go-based MCP server project",
    "goals": ["Achieve 80%+ coverage", "Maintain test performance"],
    "constraints": ["Limited dev time", "No breaking changes"]
  },
  "domain": "software testing",
  "max_cases": 5,
  "min_similarity": 0.6
}
```

**Response Format:**
```json
{
  "status": "success",
  "count": 2,
  "cases": [
    {
      "case_id": "case-123",
      "similarity": 0.85,
      "problem": {
        "description": "Improve test coverage in microservices project",
        "context": "Go microservices",
        "goals": ["Reach 85% coverage"],
        "constraints": ["2 week timeline"]
      },
      "solution": {
        "description": "Table-driven tests and mock interfaces",
        "steps": ["Identify untested paths", "Create table-driven tests", "Mock dependencies"],
        "success_rate": 0.9
      },
      "outcome": "Achieved 87% coverage in 10 days"
    }
  ]
}
```

**Use Cases:**
- Case retrieval
- Solution reuse
- Experience-based reasoning

**Related Tools:** perform-cbr-cycle

---

### 48. perform-cbr-cycle

Execute full CBR cycle (Retrieve, Reuse, Revise, Retain).

**Parameters:**
```typescript
{
  problem: {                  // Required: Current problem
    description: string;
    context?: string;
    goals?: string[];
    constraints?: string[];
    features?: object;
  };
  domain?: string;            // Optional: Domain context
}
```

**Request Example:**
```json
{
  "problem": {
    "description": "Need to improve test coverage from 50% to 80%",
    "context": "Go-based MCP server project",
    "goals": ["Achieve 80%+ coverage", "Maintain test suite performance"],
    "constraints": ["Limited development time", "Must not break existing tests"]
  },
  "domain": "software testing"
}
```

**Response Format:**
```json
{
  "status": "success",
  "retrieved_count": 3,
  "best_case": {
    "case_id": "case-123",
    "similarity": 0.85
  },
  "adapted_solution": {
    "description": "Table-driven tests adapted for MCP handlers",
    "steps": [
      "Identify handler functions with low coverage",
      "Create table-driven tests for each handler",
      "Mock storage layer for isolation",
      "Add edge case tests"
    ],
    "confidence": 0.8
  },
  "strategy": "reuse_with_adaptation",
  "recommendations": [
    "Focus on handlers package first (51.6% → 80%+ target)",
    "Use existing test patterns from validation package (91.2%)",
    "Leverage MockStorage for consistent testing"
  ]
}
```

**CBR Cycle Steps:**
1. **Retrieve**: Find similar cases from library
2. **Reuse**: Adapt solution from best matching case
3. **Revise**: Adjust solution for current context
4. **Retain**: Store successful solutions (when outcome provided)

**Use Cases:**
- Complete case-based reasoning
- Solution adaptation
- Experiential learning

**Related Tools:** retrieve-similar-cases, make-decision

---

## Symbolic Reasoning Tools

### 49. prove-theorem

Formal theorem proving using natural deduction.

**Parameters:**
```typescript
{
  name: string;           // Required: Theorem name
  premises: string[];     // Required: Premise statements
  conclusion: string;     // Required: Conclusion to prove
}
```

**Request Example:**
```json
{
  "name": "Modus Ponens Example",
  "premises": ["P", "P -> Q"],
  "conclusion": "Q"
}
```

**Response Format:**
```json
{
  "status": "proven",
  "name": "Modus Ponens Example",
  "is_valid": true,
  "confidence": 0.95,
  "proof": {
    "method": "natural_deduction",
    "steps": [
      {
        "step_number": 1,
        "statement": "P",
        "rule": "assumption",
        "justification": "Premise 1",
        "dependencies": []
      },
      {
        "step_number": 2,
        "statement": "P -> Q",
        "rule": "assumption",
        "justification": "Premise 2",
        "dependencies": []
      },
      {
        "step_number": 3,
        "statement": "Q",
        "rule": "modus_ponens",
        "justification": "From steps 1 and 2",
        "dependencies": [1, 2]
      }
    ],
    "explanation": "Conclusion Q follows from P and P -> Q by modus ponens"
  }
}
```

**Inference Rules:**
- Modus Ponens: P, P → Q ⊢ Q
- Modus Tollens: P → Q, ¬Q ⊢ ¬P
- Simplification: P ∧ Q ⊢ P
- Conjunction: P, Q ⊢ P ∧ Q
- Universal Instantiation: ∀x.P(x) ⊢ P(a)

**Use Cases:**
- Formal theorem proving
- Logical derivation
- Argument validation

**Related Tools:** check-syntax, prove, check-constraints

---

### 50. check-constraints

Check symbolic constraint satisfaction and consistency.

**Parameters:**
```typescript
{
  symbols: Array<{          // Required: Symbol definitions
    name: string;
    type: string;           // "integer" | "real" | "boolean"
    domain: string;         // e.g., "0..100", "0..inf", "true|false"
  }>;
  constraints: Array<{      // Required: Constraints
    type: string;           // "equality" | "inequality" | "logical"
    expression: string;     // Constraint expression
    symbols: string[];      // Symbols involved
  }>;
}
```

**Request Example:**
```json
{
  "symbols": [
    {"name": "coverage", "type": "real", "domain": "0..100"},
    {"name": "bugs", "type": "real", "domain": "0..inf"}
  ],
  "constraints": [
    {"type": "inequality", "expression": "coverage > 80", "symbols": ["coverage"]},
    {"type": "inequality", "expression": "bugs < 10", "symbols": ["bugs"]}
  ]
}
```

**Response Format:**
```json
{
  "is_consistent": true,
  "conflicts": [],
  "explanation": "All constraints are mutually consistent"
}
```

**If Inconsistent:**
```json
{
  "is_consistent": false,
  "conflicts": [
    {
      "constraint_1": "coverage > 80",
      "constraint_2": "coverage < 70",
      "explanation": "These constraints cannot both be satisfied"
    }
  ],
  "explanation": "Detected 1 constraint conflict"
}
```

**Use Cases:**
- Constraint satisfaction
- Consistency checking
- Symbolic problem solving

**Related Tools:** prove-theorem, validate

---

## Metadata & Export Formats

### Suggested Next Tools

Most tools return `metadata.suggested_next_tools` with intelligent recommendations:

```json
{
  "suggested_next_tools": [
    {
      "server_tool": "unified-thinking:validate",
      "reason": "Validate logical consistency of reasoning",
      "priority": "recommended",
      "input_hint": "Use thought_id from this response"
    }
  ]
}
```

### Export Formats

Tools provide ready-to-use export formats for other MCP servers:

```json
{
  "metadata": {
    "export_formats": {
      "memory_entities": [...],      // For Memory MCP
      "memory_relations": [...],     // For Memory MCP
      "obsidian_note": "...",        // For Obsidian MCP
      "brave_search_query": "..."   // For Brave Search MCP
    }
  }
}
```

### Validation Opportunities

Low-confidence thoughts include validation suggestions:

```json
{
  "metadata": {
    "validation_opportunities": [
      "Consider using web_search to gather additional evidence",
      "Query conversation:conversation_search for historical context"
    ]
  }
}
```

---

## Error Handling

All tools return consistent error format:

```json
{
  "error": "Error description",
  "status": "failed"
}
```

Common error codes:
- `"missing_parameter"` - Required parameter not provided
- `"invalid_parameter"` - Parameter value invalid
- `"not_found"` - Resource (thought/branch/graph) not found
- `"no_results"` - Search/retrieval returned no results
- `"tool_not_supported"` - Tool not available in current context

---

## Rate Limits & Performance

- **No rate limits** for MCP protocol communication
- **In-memory storage**: O(1) access, fast operations
- **SQLite storage**: O(log n) indexed access, FTS5 search
- **Typical response time**: 10-200ms depending on complexity
- **Resource limits**: MaxSearchResults=1000, MaxIndexSize=100000

---

## Version Information

**API Version**: 1.0
**MCP SDK Version**: 0.8.0
**Server Version**: See `go.mod` for current version
**Last Updated**: November 2025

---

**For More Information**:
- Project README: `README.md`
- Integration Test Report: `MCP_INTEGRATION_TEST_REPORT.md`
- Project Index: `PROJECT_INDEX.md`
- Architecture Details: `CLAUDE.md`
