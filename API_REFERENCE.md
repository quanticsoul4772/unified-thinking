# Unified Thinking Server - API Reference

This document provides a comprehensive reference for all 63 MCP tools available in the Unified Thinking Server.

## Table of Contents

1. [Core Thinking Tools](#1-core-thinking-tools)
2. [Probabilistic Reasoning Tools](#2-probabilistic-reasoning-tools)
3. [Decision & Problem-Solving Tools](#3-decision--problem-solving-tools)
4. [Metacognition Tools](#4-metacognition-tools)
5. [Hallucination & Calibration Tools](#5-hallucination--calibration-tools)
6. [Perspective & Temporal Analysis Tools](#6-perspective--temporal-analysis-tools)
7. [Causal Reasoning Tools](#7-causal-reasoning-tools)
8. [Integration & Orchestration Tools](#8-integration--orchestration-tools)
9. [Dual-Process Reasoning Tools](#9-dual-process-reasoning-tools)
10. [Backtracking Tools](#10-backtracking-tools)
11. [Abductive Reasoning Tools](#11-abductive-reasoning-tools)
12. [Case-Based Reasoning Tools](#12-case-based-reasoning-tools)
13. [Symbolic Reasoning Tools](#13-symbolic-reasoning-tools)
14. [Enhanced Tools](#14-enhanced-tools)
15. [Episodic Memory & Learning Tools](#15-episodic-memory--learning-tools)

---

## 1. Core Thinking Tools

### think

Main thinking tool supporting multiple cognitive modes.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | The thought content to process |
| `mode` | string | No | Thinking mode: "linear", "tree", "divergent", "auto" (default: "auto") |
| `confidence` | float | No | Confidence level 0.0-1.0 (default: 0.8) |
| `key_points` | string[] | No | Array of key observations |
| `branch_id` | string | No | For tree mode continuation |
| `parent_id` | string | No | Parent thought ID |
| `require_validation` | bool | No | Force validation of thought |
| `challenge_assumptions` | bool | No | Enable assumption challenging |
| `force_rebellion` | bool | No | Force divergent/creative mode |
| `cross_refs` | object[] | No | Cross-references to other branches |

**Example Request:**
```json
{
  "content": "Analyze database performance bottlenecks",
  "mode": "linear",
  "confidence": 0.7,
  "key_points": ["Query optimization", "Index analysis"]
}
```

**Example Response:**
```json
{
  "thought_id": "thought_1732234567890_1",
  "mode": "linear",
  "branch_id": "",
  "status": "success",
  "priority": 0.8,
  "confidence": 0.7,
  "insight_count": 2,
  "is_valid": true,
  "metadata": {
    "suggested_next_tools": ["validate", "decompose-problem"],
    "validation_opportunities": ["Low confidence - consider research"],
    "export_formats": {}
  }
}
```

**Common Usage Patterns:**
- Research-Enhanced Thinking: `brave_web_search` -> `think` -> `assess-evidence`
- Validated Chain: `think` -> `validate` -> `think` (iterate)
- Tree Exploration: `think` (mode: "tree") -> `synthesize-insights`

---

### history

View thinking history with optional filters.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `mode` | string | No | Filter by thinking mode |
| `branch_id` | string | No | Filter by branch ID |
| `limit` | int | No | Maximum results (default: 100) |
| `offset` | int | No | Pagination offset |

**Example Request:**
```json
{
  "mode": "linear",
  "limit": 50
}
```

**Example Response:**
```json
{
  "thoughts": [
    {
      "id": "thought_1",
      "content": "Analysis of...",
      "mode": "linear",
      "confidence": 0.85,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

---

### list-branches

List all thinking branches (tree mode).

**Parameters:** None

**Example Request:**
```json
{}
```

**Example Response:**
```json
{
  "branches": [
    {
      "id": "branch_1",
      "name": "Main Analysis",
      "confidence": 0.8,
      "thought_count": 5
    }
  ],
  "active_branch_id": "branch_1"
}
```

---

### focus-branch

Switch the active thinking branch.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `branch_id` | string | Yes | Branch ID to activate |

**Example Request:**
```json
{
  "branch_id": "branch_2"
}
```

**Example Response:**
```json
{
  "status": "success",
  "active_branch_id": "branch_2"
}
```

---

### branch-history

Get detailed history of a specific branch including insights and cross-references.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `branch_id` | string | Yes | Branch ID to retrieve |

**Example Request:**
```json
{
  "branch_id": "branch_1"
}
```

**Example Response:**
```json
{
  "branch_id": "branch_1",
  "thoughts": [],
  "insights": [],
  "cross_refs": [],
  "metrics": {
    "confidence": 0.85,
    "priority": 0.9
  }
}
```

---

### recent-branches

Get recently accessed branches for quick context switching.

**Parameters:** None

**Example Request:**
```json
{}
```

**Example Response:**
```json
{
  "active_branch_id": "branch_1",
  "recent_branches": [],
  "count": 5
}
```

---

### validate

Validate a thought for logical consistency.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | Yes | Thought ID to validate |

**Example Request:**
```json
{
  "thought_id": "thought_123"
}
```

**Example Response:**
```json
{
  "is_valid": true,
  "reason": "No logical inconsistencies detected"
}
```

---

### prove

Attempt to prove a logical conclusion from premises.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `premises` | string[] | Yes | Array of premise statements |
| `conclusion` | string | Yes | Conclusion to prove |

**Example Request:**
```json
{
  "premises": ["All humans are mortal", "Socrates is human"],
  "conclusion": "Socrates is mortal"
}
```

**Example Response:**
```json
{
  "is_provable": true,
  "premises": ["All humans are mortal", "Socrates is human"],
  "conclusion": "Socrates is mortal",
  "steps": ["Step 1: Apply modus ponens..."]
}
```

---

### check-syntax

Validate syntax of logical statements.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `statements` | string[] | Yes | Array of statements to check |

**Example Request:**
```json
{
  "statements": ["P AND Q", "IF P THEN Q"]
}
```

**Example Response:**
```json
{
  "checks": [
    {"statement": "P AND Q", "is_valid": true},
    {"statement": "IF P THEN Q", "is_valid": true}
  ]
}
```

---

### search

Search through all thoughts.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query |
| `mode` | string | No | Filter by thinking mode |
| `limit` | int | No | Maximum results (default: 100) |
| `offset` | int | No | Pagination offset |

**Example Request:**
```json
{
  "query": "database optimization",
  "limit": 20
}
```

**Example Response:**
```json
{
  "thoughts": [
    {
      "id": "thought_1",
      "content": "Database optimization strategies...",
      "mode": "linear",
      "confidence": 0.9
    }
  ]
}
```

---

### get-metrics

Get system performance and usage metrics.

**Parameters:** None

**Example Request:**
```json
{}
```

**Example Response:**
```json
{
  "total_thoughts": 150,
  "total_branches": 12,
  "total_insights": 45,
  "total_validations": 30,
  "thoughts_by_mode": {
    "linear": 80,
    "tree": 50,
    "divergent": 20
  },
  "average_confidence": 0.78,
  "context_bridge": {},
  "probabilistic": {}
}
```

---

## 2. Probabilistic Reasoning Tools

### probabilistic-reasoning

Perform Bayesian inference and update probabilistic beliefs based on evidence.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `operation` | string | Yes | Operation: "create", "update", "get", "combine" |
| `statement` | string | For create | Belief statement |
| `prior_prob` | float | For create | Prior probability 0-1 |
| `belief_id` | string | For update/get | Existing belief ID |
| `evidence_id` | string | For update | Evidence identifier |
| `likelihood` | float | For update | P(E\|H) - likelihood 0-1 |
| `evidence_prob` | float | For update | P(E) - evidence probability 0-1 |
| `belief_ids` | string[] | For combine | Array of belief IDs to combine |
| `combine_op` | string | For combine | "and" or "or" |

**Example Request (Create):**
```json
{
  "operation": "create",
  "statement": "The system has a memory leak",
  "prior_prob": 0.3
}
```

**Example Response:**
```json
{
  "belief": {
    "id": "belief_1",
    "statement": "The system has a memory leak",
    "probability": 0.3,
    "evidence_history": []
  },
  "operation": "create",
  "status": "success"
}
```

**Example Request (Update with Bayesian inference):**
```json
{
  "operation": "update",
  "belief_id": "belief_1",
  "evidence_id": "high_memory_usage",
  "likelihood": 0.8,
  "evidence_prob": 0.4
}
```

---

### assess-evidence

Assess the quality, reliability, and relevance of evidence for claims.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | Evidence content |
| `source` | string | Yes | Evidence source |
| `claim_id` | string | No | Related claim ID |
| `supports_claim` | bool | Yes | Whether evidence supports the claim |

**Example Request:**
```json
{
  "content": "Memory profiler shows 50% increase over 24 hours",
  "source": "Production monitoring",
  "supports_claim": true
}
```

**Example Response:**
```json
{
  "evidence": {
    "id": "evidence_1",
    "quality_score": 0.85,
    "reliability": 0.9,
    "relevance": 0.8
  },
  "status": "success"
}
```

---

### detect-contradictions

Detect contradictions among a set of thoughts or statements.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_ids` | string[] | No | Specific thought IDs to check |
| `branch_id` | string | No | Check all thoughts in a branch |
| `mode` | string | No | Check all thoughts in a mode |

**Example Request:**
```json
{
  "branch_id": "branch_1"
}
```

**Example Response:**
```json
{
  "contradictions": [
    {
      "thought1_id": "thought_1",
      "thought2_id": "thought_5",
      "description": "Conflicting conclusions about memory usage",
      "severity": "high"
    }
  ],
  "count": 1,
  "status": "success"
}
```

---

### sensitivity-analysis

Test robustness of conclusions to changes in underlying assumptions.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `target_claim` | string | Yes | The claim to analyze |
| `assumptions` | string[] | Yes | List of assumptions |
| `base_confidence` | float | Yes | Base confidence level |

**Example Request:**
```json
{
  "target_claim": "System will handle 10x load increase",
  "assumptions": [
    "Database can scale horizontally",
    "Network latency remains stable"
  ],
  "base_confidence": 0.75
}
```

**Example Response:**
```json
{
  "analysis": {
    "target_claim": "System will handle 10x load increase",
    "sensitivity_scores": {
      "Database can scale horizontally": 0.8,
      "Network latency remains stable": 0.6
    },
    "most_sensitive": "Database can scale horizontally",
    "robustness_score": 0.65
  },
  "status": "success"
}
```

---

## 3. Decision & Problem-Solving Tools

### make-decision

Create structured multi-criteria decision framework and recommendations.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `question` | string | Yes | Decision question |
| `options` | object[] | Yes | Array of options with id, name, description, scores, pros, cons |
| `criteria` | object[] | Yes | Array of criteria with id, name, weight, maximize flag |

**Example Request:**
```json
{
  "question": "Which database should we use?",
  "options": [
    {
      "id": "pg",
      "name": "PostgreSQL",
      "description": "Open-source relational database",
      "scores": {"cost": 0.9, "performance": 0.8, "scalability": 0.7},
      "pros": ["Mature", "Open source"],
      "cons": ["Complex scaling"],
      "total_score": 0
    },
    {
      "id": "mongo",
      "name": "MongoDB",
      "description": "Document database",
      "scores": {"cost": 0.7, "performance": 0.8, "scalability": 0.9},
      "pros": ["Easy scaling"],
      "cons": ["Licensing costs"],
      "total_score": 0
    }
  ],
  "criteria": [
    {"id": "cost", "name": "Cost", "description": "Total cost of ownership", "weight": 0.4, "maximize": true},
    {"id": "performance", "name": "Performance", "description": "Query speed", "weight": 0.3, "maximize": true},
    {"id": "scalability", "name": "Scalability", "description": "Scaling capability", "weight": 0.3, "maximize": true}
  ]
}
```

**Example Response:**
```json
{
  "decision": {
    "id": "decision_1",
    "question": "Which database should we use?",
    "recommendation": "pg",
    "confidence": 0.82,
    "scores": {
      "pg": 0.82,
      "mongo": 0.78
    },
    "analysis": "PostgreSQL scores highest..."
  },
  "status": "success",
  "metadata": {
    "export_formats": {
      "obsidian_note": "# Decision: Database Selection..."
    }
  }
}
```

---

### decompose-problem

Break down complex problems into manageable subproblems with dependencies.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `problem` | string | Yes | Complex problem statement |

**Example Request:**
```json
{
  "problem": "How to improve CI/CD pipeline performance?"
}
```

**Example Response:**
```json
{
  "can_decompose": true,
  "problem_type": "decomposable",
  "decomposition": {
    "id": "decomp_1",
    "original_problem": "How to improve CI/CD pipeline performance?",
    "subproblems": [
      {
        "id": "sub_1",
        "description": "Analyze current pipeline bottlenecks",
        "dependencies": []
      },
      {
        "id": "sub_2",
        "description": "Optimize build step",
        "dependencies": ["sub_1"]
      }
    ],
    "solution_path": ["sub_1", "sub_2"]
  },
  "status": "success",
  "metadata": {
    "suggested_next_tools": ["think", "search"],
    "export_formats": {
      "obsidian_note": "# Problem Decomposition..."
    }
  }
}
```

---

## 4. Metacognition Tools

### self-evaluate

Perform metacognitive self-assessment of reasoning quality and completeness.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | No* | Thought ID to evaluate |
| `branch_id` | string | No* | Branch ID to evaluate |

*Either `thought_id` or `branch_id` must be provided.

**Example Request:**
```json
{
  "thought_id": "thought_123"
}
```

**Example Response:**
```json
{
  "evaluation": {
    "quality_score": 0.78,
    "completeness_score": 0.85,
    "coherence_score": 0.9,
    "strengths": ["Well-structured argument", "Good evidence support"],
    "weaknesses": ["Missing alternative perspectives"],
    "recommendations": ["Consider counterarguments"]
  },
  "status": "success"
}
```

---

### detect-biases

Identify cognitive biases AND logical fallacies in reasoning (comprehensive analysis).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | No* | Thought ID to analyze |
| `branch_id` | string | No* | Branch ID to analyze |

*Either `thought_id` or `branch_id` must be provided.

**Example Request:**
```json
{
  "thought_id": "thought_123"
}
```

**Example Response:**
```json
{
  "biases": [
    {
      "bias_type": "confirmation_bias",
      "description": "Evidence selection favors initial hypothesis",
      "detected_in": "thought_123",
      "severity": "medium",
      "mitigation": "Actively seek disconfirming evidence"
    }
  ],
  "fallacies": [
    {
      "type": "hasty_generalization",
      "category": "informal",
      "explanation": "Conclusion drawn from insufficient samples",
      "location": "Line 3",
      "suggestion": "Gather more data points"
    }
  ],
  "combined": [
    {
      "type": "bias",
      "name": "confirmation_bias",
      "category": "cognitive",
      "description": "Evidence selection favors initial hypothesis",
      "confidence": 0.6
    }
  ],
  "count": 2,
  "status": "success"
}
```

---

### detect-blind-spots

Detect unknown unknowns, blind spots, and knowledge gaps using metacognitive analysis.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | Content to analyze |
| `domain` | string | No | Problem domain |
| `context` | string | No | Additional context |
| `assumptions` | string[] | No | Stated assumptions |
| `confidence` | float | No | Current confidence level |

**Example Request:**
```json
{
  "content": "Our API will handle 1M requests per day easily",
  "domain": "system-design",
  "assumptions": ["Linear scaling", "Consistent traffic patterns"]
}
```

**Example Response:**
```json
{
  "blind_spots": [
    "Traffic spike scenarios not considered",
    "Database connection pooling limits",
    "Third-party API rate limits"
  ],
  "missing_considerations": [
    "Geographic distribution of users",
    "Cache invalidation strategies"
  ],
  "unchallenged_assumptions": [
    "Linear scaling - may not hold under high load"
  ],
  "suggested_questions": [
    "What happens during 10x traffic spikes?",
    "How does the system behave with cold caches?"
  ],
  "overall_risk": 0.65,
  "risk_level": "medium"
}
```

---

## 5. Hallucination & Calibration Tools

### verify-thought

Verify a thought for hallucinations using semantic uncertainty measurement.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | Yes | Thought ID to verify |
| `verification_level` | string | No | "fast", "deep", or "hybrid" (default: "hybrid") |

**Example Request:**
```json
{
  "thought_id": "thought_123",
  "verification_level": "hybrid"
}
```

**Example Response:**
```json
{
  "overall_risk": 0.25,
  "semantic_uncertainty": {
    "aleatory": 0.2,
    "epistemic": 0.3,
    "confidence_mismatch": 0.1
  },
  "claims": [
    {
      "claim": "PostgreSQL handles 100k TPS",
      "status": "verified",
      "confidence": 0.85
    }
  ],
  "verified_count": 3,
  "hallucination_count": 0,
  "recommendations": ["Consider adding source citations"]
}
```

---

### get-hallucination-report

Retrieve cached hallucination verification report for a thought.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | Yes | Thought ID |

**Example Request:**
```json
{
  "thought_id": "thought_123"
}
```

---

### record-prediction

Record a confidence prediction for calibration tracking.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | Yes | Thought ID |
| `confidence` | float | Yes | Confidence score 0-1 |
| `mode` | string | Yes | Thinking mode used |
| `metadata` | object | No | Additional metadata |

**Example Request:**
```json
{
  "thought_id": "thought_123",
  "confidence": 0.8,
  "mode": "linear"
}
```

---

### record-outcome

Record the actual outcome of a prediction for calibration.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `thought_id` | string | Yes | Thought ID (must have existing prediction) |
| `was_correct` | bool | Yes | Whether the thought was correct |
| `actual_confidence` | float | Yes | Actual confidence based on validation 0-1 |
| `source` | string | Yes | How outcome was determined: "validation", "verification", "user_feedback" |
| `metadata` | object | No | Additional context |

**Example Request:**
```json
{
  "thought_id": "thought_123",
  "was_correct": true,
  "actual_confidence": 0.9,
  "source": "validation"
}
```

---

### get-calibration-report

Generate comprehensive confidence calibration report.

**Parameters:** None

**Example Request:**
```json
{}
```

**Example Response:**
```json
{
  "total_predictions": 150,
  "total_outcomes": 120,
  "buckets": {
    "0-10": {"count": 5, "accuracy": 0.0},
    "10-20": {"count": 10, "accuracy": 0.1},
    "80-90": {"count": 30, "accuracy": 0.83}
  },
  "overall_accuracy": 0.75,
  "calibration": 0.08,
  "bias": "slight_overconfidence",
  "by_mode": {
    "linear": {"accuracy": 0.78},
    "tree": {"accuracy": 0.72}
  },
  "recommendations": [
    "Consider lowering confidence for tree mode by ~5%"
  ]
}
```

---

## 6. Perspective & Temporal Analysis Tools

### analyze-perspectives

Analyze a situation from multiple stakeholder perspectives.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `situation` | string | Yes | Situation to analyze |
| `stakeholder_hints` | string[] | No | Stakeholder types/names to consider |

**Example Request:**
```json
{
  "situation": "Implementing new authentication system",
  "stakeholder_hints": ["developers", "security team", "end users"]
}
```

**Example Response:**
```json
{
  "perspectives": [
    {
      "stakeholder": "developers",
      "viewpoint": "Focus on implementation simplicity",
      "concerns": ["Migration complexity", "Learning curve"],
      "priorities": ["Good documentation", "Easy testing"]
    },
    {
      "stakeholder": "security team",
      "viewpoint": "Focus on threat mitigation",
      "concerns": ["Token security", "Audit logging"],
      "priorities": ["Compliance", "Penetration testing"]
    }
  ],
  "count": 3,
  "conflicts": ["Speed vs security tradeoffs"],
  "status": "success"
}
```

---

### analyze-temporal

Analyze short-term vs long-term implications of a decision.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `situation` | string | Yes | Decision or situation to analyze |
| `time_horizon` | string | No | "days-weeks", "months", "years" (default: "months") |

**Example Request:**
```json
{
  "situation": "Refactor authentication module now or after release?",
  "time_horizon": "months"
}
```

**Example Response:**
```json
{
  "analysis": {
    "situation": "Refactor authentication module now or after release?",
    "short_term": {
      "impacts": ["Delays release by 2 weeks"],
      "benefits": ["Cleaner codebase"],
      "risks": ["Scope creep"]
    },
    "long_term": {
      "impacts": ["Reduced maintenance burden"],
      "benefits": ["Easier feature additions"],
      "risks": ["Technical debt if delayed"]
    },
    "tradeoffs": ["Immediate delay vs future velocity"],
    "recommendation": "Refactor now to avoid compounding debt"
  },
  "status": "success"
}
```

---

### compare-time-horizons

Compare how a decision looks across different time horizons.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `situation` | string | Yes | Situation to compare |

**Example Request:**
```json
{
  "situation": "Adopting microservices architecture"
}
```

**Example Response:**
```json
{
  "analyses": {
    "days-weeks": {
      "recommendation": "Negative - significant initial overhead"
    },
    "months": {
      "recommendation": "Mixed - costs and benefits balancing"
    },
    "years": {
      "recommendation": "Positive - scalability benefits realized"
    }
  },
  "status": "success"
}
```

---

### identify-optimal-timing

Determine optimal timing for a decision based on situation and constraints.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `situation` | string | Yes | Decision situation |
| `constraints` | string[] | No | Time or resource constraints |

**Example Request:**
```json
{
  "situation": "When to migrate to new database?",
  "constraints": ["Q4 feature freeze", "Team availability in January"]
}
```

**Example Response:**
```json
{
  "recommendation": "Begin migration planning in December, execute in January before Q1 traffic increase",
  "status": "success"
}
```

---

## 7. Causal Reasoning Tools

### build-causal-graph

Construct a causal graph from observations, identifying variables and causal relationships.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `description` | string | Yes | Context for the causal model |
| `observations` | string[] | Yes | Array of causal statements |

**Example Request:**
```json
{
  "description": "E-commerce sales process",
  "observations": [
    "Marketing increases brand awareness",
    "Brand awareness drives website traffic",
    "Website traffic leads to purchases"
  ]
}
```

**Example Response:**
```json
{
  "graph": {
    "id": "graph_123",
    "description": "E-commerce sales process",
    "variables": [
      {"id": "marketing", "name": "Marketing"},
      {"id": "awareness", "name": "Brand Awareness"},
      {"id": "traffic", "name": "Website Traffic"},
      {"id": "purchases", "name": "Purchases"}
    ],
    "links": [
      {"from": "marketing", "to": "awareness", "strength": 0.8},
      {"from": "awareness", "to": "traffic", "strength": 0.7},
      {"from": "traffic", "to": "purchases", "strength": 0.6}
    ]
  },
  "status": "success"
}
```

---

### simulate-intervention

Simulate the effects of intervening on a variable in a causal graph.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Causal graph ID |
| `variable_id` | string | Yes | Variable to intervene on |
| `intervention_type` | string | Yes | "increase", "decrease", "remove", "introduce" |

**Example Request:**
```json
{
  "graph_id": "graph_123",
  "variable_id": "marketing",
  "intervention_type": "increase"
}
```

**Example Response:**
```json
{
  "intervention": {
    "graph_id": "graph_123",
    "target_variable": "marketing",
    "type": "increase",
    "effects": [
      {"variable": "awareness", "change": "+25%"},
      {"variable": "traffic", "change": "+18%"},
      {"variable": "purchases", "change": "+11%"}
    ],
    "confidence": 0.75
  },
  "status": "success"
}
```

---

### generate-counterfactual

Generate a counterfactual scenario ("what if") by changing variables in a causal model.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Causal graph ID |
| `scenario` | string | Yes | Scenario description |
| `changes` | object | Yes | Variable changes as key-value pairs |

**Example Request:**
```json
{
  "graph_id": "graph_123",
  "scenario": "What if we had doubled marketing spend?",
  "changes": {
    "marketing": "2x"
  }
}
```

**Example Response:**
```json
{
  "counterfactual": {
    "scenario": "What if we had doubled marketing spend?",
    "original_outcome": {"purchases": 1000},
    "counterfactual_outcome": {"purchases": 1650},
    "explanation": "Doubling marketing would increase purchases by ~65%"
  },
  "status": "success"
}
```

---

### analyze-correlation-vs-causation

Analyze whether an observed relationship is likely correlation or causation.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `observation` | string | Yes | Observed relationship to analyze |

**Example Request:**
```json
{
  "observation": "Ice cream sales and drowning deaths both increase in summer"
}
```

**Example Response:**
```json
{
  "analysis": "This is correlation, not causation. Both variables are influenced by a common cause (warm weather). Ice cream consumption does not cause drowning.",
  "status": "success"
}
```

---

### get-causal-graph

Retrieve a previously built causal graph by ID.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Causal graph ID |

**Example Request:**
```json
{
  "graph_id": "graph_123"
}
```

---

## 8. Integration & Orchestration Tools

### synthesize-insights

Synthesize insights from multiple reasoning modes, identifying synergies and conflicts.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `context` | string | Yes | Context for synthesis |
| `inputs` | object[] | Yes | Array of inputs with ID, Mode, Content, Confidence, Metadata |

**Example Request:**
```json
{
  "context": "System architecture decision",
  "inputs": [
    {
      "ID": "linear_1",
      "Mode": "linear",
      "Content": "Systematic analysis shows microservices benefit",
      "Confidence": 0.8,
      "Metadata": {}
    },
    {
      "ID": "divergent_1",
      "Mode": "divergent",
      "Content": "Consider serverless as alternative",
      "Confidence": 0.6,
      "Metadata": {}
    }
  ]
}
```

**Example Response:**
```json
{
  "synthesis": {
    "unified_insight": "Hybrid approach: microservices core with serverless for variable workloads",
    "synergies": ["Both approaches support scaling"],
    "conflicts": ["Operational complexity vs simplicity"],
    "confidence": 0.75
  },
  "status": "success"
}
```

---

### detect-emergent-patterns

Detect emergent patterns that become visible when combining multiple reasoning modes.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `inputs` | object[] | Yes | Array of inputs with ID, Mode, Content, Confidence, Metadata |

**Example Request:**
```json
{
  "inputs": [
    {"ID": "1", "Mode": "linear", "Content": "Users complain about speed", "Confidence": 0.9, "Metadata": {}},
    {"ID": "2", "Mode": "tree", "Content": "Database queries are slow", "Confidence": 0.8, "Metadata": {}},
    {"ID": "3", "Mode": "divergent", "Content": "Consider caching strategy", "Confidence": 0.7, "Metadata": {}}
  ]
}
```

**Example Response:**
```json
{
  "patterns": [
    "Performance issues cluster around data access layer",
    "User experience directly tied to backend optimization"
  ],
  "count": 2,
  "status": "success"
}
```

---

### execute-workflow

Execute a predefined reasoning workflow with automatic tool chaining.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workflow_id` | string | Yes | Workflow identifier |
| `input` | object | Yes | Workflow parameters (must include "problem" field) |

**Example Request:**
```json
{
  "workflow_id": "comprehensive-analysis",
  "input": {
    "problem": "Optimize database query performance"
  }
}
```

**Example Response:**
```json
{
  "result": {
    "workflow_id": "comprehensive-analysis",
    "steps_completed": 5,
    "final_output": {
      "analysis": "...",
      "recommendations": []
    }
  },
  "status": "success"
}
```

---

### list-workflows

List all available automated workflows for multi-tool reasoning pipelines.

**Parameters:** None

**Example Request:**
```json
{}
```

**Example Response:**
```json
{
  "workflows": [
    {
      "id": "comprehensive-analysis",
      "name": "Comprehensive Analysis",
      "description": "Full analysis pipeline",
      "steps": ["think", "validate", "assess-evidence", "synthesize"]
    }
  ],
  "count": 3
}
```

---

### register-workflow

Register a new custom workflow for automated tool coordination.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `workflow` | object | Yes | Workflow definition with id, name, description, type, steps |

**Example Request:**
```json
{
  "workflow": {
    "id": "my-custom-workflow",
    "name": "Custom Analysis",
    "description": "My custom reasoning pipeline",
    "type": "sequential",
    "steps": [
      {"id": "step1", "tool": "think", "input": {}},
      {"id": "step2", "tool": "validate", "input": {}, "depends_on": ["step1"]}
    ],
    "created_at": "2024-01-15T10:00:00Z"
  }
}
```

---

### list-integration-patterns

List common multi-server workflow patterns for orchestrating tools across the MCP ecosystem.

**Parameters:** None

**Example Request:**
```json
{}
```

**Example Response:**
```json
{
  "patterns": [
    {
      "name": "Research-Enhanced Thinking",
      "description": "Combine web search with reasoning",
      "steps": ["brave_web_search", "think", "assess-evidence"],
      "use_case": "When you need external information",
      "servers": ["brave-search", "unified-thinking"]
    }
  ],
  "count": 10,
  "status": "success"
}
```

---

## 9. Dual-Process Reasoning Tools

### dual-process-think

Execute dual-process reasoning (System 1: fast/intuitive, System 2: slow/analytical).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | Thought content |
| `mode` | string | No | Thinking mode (default: "linear") |
| `branch_id` | string | No | Branch for tree mode |
| `force_system` | string | No | "system1", "system2", or empty for auto |
| `key_points` | string[] | No | Key observations |
| `metadata` | object | No | Additional metadata |

**Example Request:**
```json
{
  "content": "Should we refactor this function?",
  "force_system": "system2"
}
```

**Example Response:**
```json
{
  "thought_id": "thought_456",
  "system_used": "system2",
  "complexity": 0.65,
  "escalated": false,
  "system1_time": "5ms",
  "system2_time": "250ms",
  "confidence": 0.85,
  "content": "Detailed analysis suggests refactoring...",
  "metadata": {}
}
```

---

## 10. Backtracking Tools

### create-checkpoint

Create a backtracking checkpoint in tree mode for later restoration.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `branch_id` | string | Yes | Branch to checkpoint |
| `name` | string | Yes | Checkpoint name |
| `description` | string | No | Checkpoint description |

**Example Request:**
```json
{
  "branch_id": "branch_1",
  "name": "Before risky exploration",
  "description": "Checkpoint before testing radical approach"
}
```

**Example Response:**
```json
{
  "checkpoint_id": "checkpoint_123",
  "name": "Before risky exploration",
  "description": "Checkpoint before testing radical approach",
  "branch_id": "branch_1",
  "thought_count": 5,
  "insight_count": 2,
  "created_at": "2024-01-15T10:30:00Z"
}
```

---

### restore-checkpoint

Restore branch from a checkpoint, enabling backtracking in tree exploration.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `checkpoint_id` | string | Yes | Checkpoint to restore |

**Example Request:**
```json
{
  "checkpoint_id": "checkpoint_123"
}
```

**Example Response:**
```json
{
  "branch_id": "branch_1",
  "thought_count": 5,
  "insight_count": 2,
  "message": "Checkpoint restored successfully"
}
```

---

### list-checkpoints

List available checkpoints for backtracking.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `branch_id` | string | No | Filter by branch ID |

**Example Request:**
```json
{
  "branch_id": "branch_1"
}
```

**Example Response:**
```json
{
  "checkpoints": [
    {
      "id": "checkpoint_123",
      "name": "Before risky exploration",
      "description": "...",
      "branch_id": "branch_1",
      "thought_count": 5,
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "count": 1
}
```

---

## 11. Abductive Reasoning Tools

### generate-hypotheses

Generate hypotheses from observations using abductive reasoning (inference to best explanation).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `observations` | object[] | Yes | Array of observations with description, confidence |
| `max_hypotheses` | int | No | Maximum hypotheses to generate (default: 10) |
| `min_parsimony` | float | No | Minimum parsimony threshold |

**Example Request:**
```json
{
  "observations": [
    {"description": "Server response times increased", "confidence": 0.9},
    {"description": "Memory usage spiking", "confidence": 0.85},
    {"description": "No code deployments recently", "confidence": 0.95}
  ],
  "max_hypotheses": 5
}
```

**Example Response:**
```json
{
  "hypotheses": [
    {
      "id": "hyp_1",
      "description": "Memory leak in background process",
      "observations": ["Memory usage spiking", "Server response times increased"],
      "parsimony": 0.8,
      "prior_probability": 0.6,
      "assumptions": ["Background processes are running"]
    },
    {
      "id": "hyp_2",
      "description": "Database connection pool exhaustion",
      "observations": ["Server response times increased"],
      "parsimony": 0.7,
      "prior_probability": 0.4,
      "assumptions": []
    }
  ],
  "count": 2
}
```

---

### evaluate-hypotheses

Evaluate and rank hypotheses using Bayesian inference, parsimony, and explanatory power.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `observations` | object[] | Yes | Array of observations |
| `hypotheses` | object[] | Yes | Array of hypotheses to evaluate |
| `method` | string | No | "bayesian", "parsimony", "combined" (default: "combined") |

**Example Request:**
```json
{
  "observations": [
    {"description": "Server response times increased", "confidence": 0.9}
  ],
  "hypotheses": [
    {"description": "Memory leak", "observations": ["Server response times increased"], "prior_probability": 0.5},
    {"description": "Network congestion", "observations": ["Server response times increased"], "prior_probability": 0.3}
  ],
  "method": "combined"
}
```

**Example Response:**
```json
{
  "ranked_hypotheses": [
    {
      "description": "Memory leak",
      "posterior_probability": 0.65,
      "explanatory_power": 0.8,
      "parsimony": 0.7,
      "rank": 1
    },
    {
      "description": "Network congestion",
      "posterior_probability": 0.35,
      "explanatory_power": 0.5,
      "parsimony": 0.6,
      "rank": 2
    }
  ],
  "best_hypothesis": {
    "description": "Memory leak",
    "posterior_probability": 0.65
  },
  "method": "combined"
}
```

---

## 12. Case-Based Reasoning Tools

### retrieve-similar-cases

Retrieve similar cases from case library using CBR (case-based reasoning).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `problem` | object | Yes | Problem with description, context, goals, constraints |
| `domain` | string | No | Problem domain |
| `max_cases` | int | No | Maximum cases to retrieve (default: 5) |
| `min_similarity` | float | No | Minimum similarity threshold (default: 0.3) |

**Example Request:**
```json
{
  "problem": {
    "description": "API response times exceeding SLA",
    "context": "Production environment",
    "goals": ["Reduce latency to <200ms"],
    "constraints": ["No infrastructure changes"]
  },
  "domain": "performance",
  "max_cases": 3
}
```

**Example Response:**
```json
{
  "cases": [
    {
      "case_id": "case_45",
      "problem": {
        "description": "Similar API latency issue",
        "context": "Staging environment",
        "goals": ["Reduce latency"]
      },
      "solution": {
        "description": "Added caching layer",
        "approach": "Redis caching for frequent queries",
        "steps": ["Identify slow endpoints", "Add cache"]
      },
      "similarity": 0.85,
      "success_rate": 0.9,
      "domain": "performance"
    }
  ],
  "retrieved": 1
}
```

---

### perform-cbr-cycle

Perform full CBR cycle: Retrieve similar cases, Reuse/adapt solution, provide recommendations.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `problem` | object | Yes | Problem with description, context, goals, constraints |
| `domain` | string | No | Problem domain |

**Example Request:**
```json
{
  "problem": {
    "description": "Database queries taking too long",
    "context": "E-commerce platform",
    "goals": ["Queries under 100ms"]
  },
  "domain": "database"
}
```

**Example Response:**
```json
{
  "retrieved_count": 3,
  "best_case": {
    "case_id": "case_78",
    "similarity": 0.88
  },
  "adapted_solution": {
    "description": "Add indexes and optimize queries",
    "steps": [
      "Analyze slow query log",
      "Add composite indexes",
      "Rewrite N+1 queries"
    ]
  },
  "strategy": "Adaptation based on 3 similar cases",
  "confidence": 0.82
}
```

---

## 13. Symbolic Reasoning Tools

### prove-theorem

Attempt to prove a theorem using symbolic reasoning and logical inference rules.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | No | Theorem name |
| `premises` | string[] | Yes | Array of premise statements |
| `conclusion` | string | Yes | Conclusion to prove |

**Example Request:**
```json
{
  "name": "Transitivity",
  "premises": ["A implies B", "B implies C"],
  "conclusion": "A implies C"
}
```

**Example Response:**
```json
{
  "name": "Transitivity",
  "status": "proven",
  "is_valid": true,
  "confidence": 0.95,
  "proof": {
    "steps": [
      {
        "step_number": 1,
        "statement": "A implies B",
        "justification": "Premise",
        "rule": "given",
        "dependencies": []
      },
      {
        "step_number": 2,
        "statement": "B implies C",
        "justification": "Premise",
        "rule": "given",
        "dependencies": []
      },
      {
        "step_number": 3,
        "statement": "A implies C",
        "justification": "Hypothetical syllogism",
        "rule": "hypothetical_syllogism",
        "dependencies": [1, 2]
      }
    ],
    "method": "forward_chaining",
    "explanation": "Applied hypothetical syllogism to derive conclusion"
  }
}
```

---

### check-constraints

Check consistency of symbolic constraints. Detect conflicts and contradictions.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `symbols` | object[] | Yes | Array of symbols with name, type, domain |
| `constraints` | object[] | Yes | Array of constraints with type, expression, symbols |

**Example Request:**
```json
{
  "symbols": [
    {"name": "x", "type": "variable", "domain": "integer"},
    {"name": "y", "type": "variable", "domain": "integer"}
  ],
  "constraints": [
    {"type": "inequality", "expression": "x > 10", "symbols": ["x"]},
    {"type": "inequality", "expression": "x < 5", "symbols": ["x"]}
  ]
}
```

**Example Response:**
```json
{
  "is_consistent": false,
  "conflicts": [
    {
      "constraint1": "x > 10",
      "constraint2": "x < 5",
      "conflict_type": "contradiction",
      "explanation": "No value of x can be both greater than 10 and less than 5"
    }
  ],
  "explanation": "Constraints are inconsistent due to contradictory requirements on x"
}
```

---

## 14. Enhanced Tools

### find-analogy

Find analogies between source and target domains for cross-domain reasoning.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `source_domain` | string | Yes | Source domain for analogy |
| `target_problem` | string | Yes | Target problem to solve |
| `constraints` | string[] | No | Constraints on the analogy |

**Example Request:**
```json
{
  "source_domain": "biology: immune system",
  "target_problem": "How to protect a computer network from attacks?"
}
```

**Example Response:**
```json
{
  "analogy": {
    "id": "analogy_123",
    "source_domain": "biology: immune system",
    "target_problem": "How to protect a computer network from attacks?",
    "mappings": [
      {"source": "antibodies", "target": "firewall rules"},
      {"source": "white blood cells", "target": "intrusion detection"},
      {"source": "immune memory", "target": "threat database"}
    ],
    "insights": [
      "Layered defense like skin + immune system",
      "Adaptive response to new threats"
    ],
    "confidence": 0.78
  },
  "status": "success"
}
```

---

### apply-analogy

Apply an existing analogy to a new context.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `analogy_id` | string | Yes | Analogy ID from find-analogy |
| `target_context` | string | Yes | New context to apply analogy |

**Example Request:**
```json
{
  "analogy_id": "analogy_123",
  "target_context": "Securing an IoT network"
}
```

---

### decompose-argument

Break down an argument into premises, claims, assumptions, and inference chains.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `argument` | string | Yes | Argument to decompose |

**Example Request:**
```json
{
  "argument": "We should adopt microservices because studies show they reduce deployment time by 40%, and faster deployments lead to better customer satisfaction."
}
```

**Example Response:**
```json
{
  "decomposition": {
    "id": "arg_123",
    "main_claim": "We should adopt microservices",
    "premises": [
      "Studies show microservices reduce deployment time by 40%",
      "Faster deployments lead to better customer satisfaction"
    ],
    "assumptions": [
      "The studies are applicable to our context",
      "Customer satisfaction is a priority"
    ],
    "inference_chain": [
      "Microservices -> faster deployment -> better satisfaction -> should adopt"
    ]
  },
  "status": "success"
}
```

---

### generate-counter-arguments

Generate counter-arguments for a given argument using multiple strategies.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `argument_id` | string | Yes | Argument ID from decompose-argument |

**Example Request:**
```json
{
  "argument_id": "arg_123"
}
```

**Example Response:**
```json
{
  "counter_arguments": [
    {
      "type": "premise_attack",
      "content": "The 40% improvement may not apply to all organizations",
      "targets": "Studies show microservices reduce deployment time by 40%"
    },
    {
      "type": "assumption_challenge",
      "content": "Correlation between deployment speed and satisfaction may be weak",
      "targets": "Faster deployments lead to better customer satisfaction"
    }
  ],
  "count": 2,
  "status": "success"
}
```

---

### detect-fallacies

Detect formal and informal logical fallacies in reasoning.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | Content to analyze |
| `check_formal` | bool | No | Check formal fallacies (default: true) |
| `check_informal` | bool | No | Check informal fallacies (default: true) |

**Example Request:**
```json
{
  "content": "Everyone on the team agrees we should use this framework, so it must be the best choice.",
  "check_formal": true,
  "check_informal": true
}
```

**Example Response:**
```json
{
  "fallacies": [
    {
      "type": "appeal_to_popularity",
      "category": "informal",
      "explanation": "The popularity of an opinion does not determine its truth",
      "location": "Everyone on the team agrees...",
      "suggestion": "Evaluate the framework based on technical merits"
    }
  ],
  "count": 1,
  "status": "success"
}
```

---

### process-evidence-pipeline

Process evidence and auto-update beliefs, causal graphs, and decisions.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `content` | string | Yes | Evidence content |
| `source` | string | Yes | Evidence source |
| `claim_id` | string | No | Related claim ID |
| `supports_claim` | bool | Yes | Whether evidence supports the claim |

**Example Request:**
```json
{
  "content": "New study shows microservices increase operational complexity by 60%",
  "source": "IEEE Software 2024",
  "supports_claim": false
}
```

---

### analyze-temporal-causal-effects

Analyze how causal effects evolve across different time horizons.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Causal graph ID from build-causal-graph |
| `variable_id` | string | Yes | Variable to analyze |
| `intervention_type` | string | Yes | "increase", "decrease", "remove", "introduce" |

**Example Request:**
```json
{
  "graph_id": "graph_123",
  "variable_id": "marketing_spend",
  "intervention_type": "increase"
}
```

---

### analyze-decision-timing

Determine optimal timing for decisions based on causal and temporal analysis.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `situation` | string | Yes | Decision situation |
| `causal_graph_id` | string | No | Related causal graph ID |

**Example Request:**
```json
{
  "situation": "When to launch new product?",
  "causal_graph_id": "graph_123"
}
```

---

## 15. Episodic Memory & Learning Tools

### start-reasoning-session

Start tracking a reasoning session to build episodic memory and learn from experience.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_id` | string | Yes | Unique session identifier |
| `description` | string | Yes | Problem description |
| `goals` | string[] | No | Goals to achieve |
| `domain` | string | No | Problem domain (e.g., "software-engineering") |
| `context` | string | No | Additional context |
| `complexity` | float | No | Estimated complexity 0.0-1.0 |
| `metadata` | object | No | Additional metadata |

**Example Request:**
```json
{
  "session_id": "debug_2024_001",
  "description": "Optimize database query performance",
  "goals": ["Reduce query time", "Improve user experience"],
  "domain": "software-engineering",
  "complexity": 0.6
}
```

**Example Response:**
```json
{
  "session_id": "debug_2024_001",
  "problem_id": "prob_abc123",
  "status": "active",
  "suggestions": [
    {
      "type": "tool_sequence",
      "suggestion": "Similar problems benefited from: think -> decompose-problem -> think",
      "success_rate": 0.85,
      "reasoning": "Based on 3 similar past sessions"
    }
  ]
}
```

---

### complete-reasoning-session

Complete a reasoning session and store the trajectory for learning.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_id` | string | Yes | Session to complete |
| `status` | string | Yes | "success", "partial", or "failure" |
| `goals_achieved` | string[] | No | Array of achieved goals |
| `goals_failed` | string[] | No | Array of failed goals |
| `solution` | string | No | Description of solution |
| `confidence` | float | No | Confidence in solution 0.0-1.0 |
| `unexpected_outcomes` | string[] | No | Unexpected results |

**Example Request:**
```json
{
  "session_id": "debug_2024_001",
  "status": "success",
  "goals_achieved": ["Reduce query time"],
  "solution": "Added composite indexes and rewrote N+1 queries",
  "confidence": 0.85
}
```

**Example Response:**
```json
{
  "trajectory_id": "traj_debug_2024_001_abc123",
  "session_id": "debug_2024_001",
  "success_score": 0.9,
  "quality_score": 0.85,
  "patterns_found": 2,
  "status": "completed"
}
```

---

### get-recommendations

Get adaptive recommendations based on episodic memory of similar past problems.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `description` | string | Yes | Problem description |
| `goals` | string[] | No | Problem goals |
| `domain` | string | No | Problem domain |
| `context` | string | No | Additional context |
| `complexity` | float | No | Estimated complexity 0.0-1.0 |
| `limit` | int | No | Max recommendations (default: 5) |

**Example Request:**
```json
{
  "description": "Need to implement user authentication",
  "domain": "security",
  "goals": ["Secure login", "Session management"],
  "limit": 3
}
```

**Example Response:**
```json
{
  "recommendations": [
    {
      "type": "tool_sequence",
      "priority": 0.9,
      "suggestion": "Use: decompose-problem -> think (security) -> detect-blind-spots",
      "reasoning": "This sequence had 85% success rate for auth implementations",
      "success_rate": 0.85
    },
    {
      "type": "warning",
      "priority": 0.7,
      "suggestion": "Avoid skipping threat modeling - led to failures in 40% of cases",
      "reasoning": "Pattern detected from failed sessions",
      "success_rate": 0.6
    }
  ],
  "similar_cases": 5,
  "learned_patterns": [
    {
      "pattern": "Security problems benefit from multi-perspective analysis",
      "confidence": 0.8
    }
  ],
  "count": 2
}
```

---

### search-trajectories

Search for past reasoning trajectories to learn from experience.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `domain` | string | No | Filter by domain |
| `tags` | string[] | No | Filter by tags |
| `min_success` | float | No | Minimum success score 0.0-1.0 |
| `problem_type` | string | No | Filter by problem type |
| `limit` | int | No | Max results (default: 10) |

**Example Request:**
```json
{
  "domain": "software-engineering",
  "min_success": 0.7,
  "limit": 5
}
```

**Example Response:**
```json
{
  "trajectories": [
    {
      "id": "traj_123",
      "session_id": "session_456",
      "problem": "Optimize CI/CD pipeline",
      "domain": "software-engineering",
      "strategy": "decomposition-first",
      "tools_used": ["decompose-problem", "think", "make-decision"],
      "success_score": 0.9,
      "duration": "45m30s",
      "tags": ["devops", "optimization"]
    }
  ],
  "count": 1
}
```

---

### analyze-trajectory

Perform retrospective analysis of a completed reasoning session.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `trajectory_id` | string | Yes | Trajectory ID from complete-reasoning-session |

**Example Request:**
```json
{
  "trajectory_id": "traj_debug_2024_001_abc123"
}
```

**Example Response:**
```json
{
  "summary": {
    "assessment": "good",
    "success_score": 0.85,
    "quality_score": 0.8,
    "duration": "32m15s",
    "strategy": "systematic-analysis"
  },
  "strengths": [
    "Efficient problem decomposition",
    "Good evidence gathering",
    "Clear decision rationale"
  ],
  "weaknesses": [
    "Could have explored more alternatives",
    "Some assumptions not validated"
  ],
  "improvements": [
    {
      "category": "approach",
      "suggestion": "Consider using detect-blind-spots earlier",
      "expected_impact": "15% better coverage",
      "priority": "medium"
    }
  ],
  "lessons_learned": [
    "Database problems often benefit from systematic query analysis",
    "Index recommendations should be validated with EXPLAIN"
  ],
  "comparative_analysis": {
    "percentile_rank": 75,
    "comparison": "Better than 75% of similar sessions"
  }
}
```

---

## Knowledge Graph Tools

### store-entity

Store an entity in the knowledge graph with semantic indexing.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `entity_id` | string | Yes | Unique entity identifier |
| `label` | string | Yes | Human-readable entity label |
| `type` | string | Yes | Entity type (Concept, Person, Tool, File, Decision, Strategy, Problem) |
| `content` | string | Yes | Content for semantic search embedding |
| `description` | string | No | Detailed description |
| `metadata` | object | No | Additional metadata |

---

### search-knowledge-graph

Search for entities using semantic similarity or graph traversal.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Search query |
| `search_type` | string | Yes | "semantic" or "hybrid" |
| `limit` | integer | No | Max results (default: 10) |
| `max_hops` | integer | No | For hybrid search, max graph hops (default: 1) |
| `min_similarity` | float | No | Minimum similarity threshold (default: 0.7) |

---

### create-relationship

Create a typed relationship between entities in the knowledge graph.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `relationship_id` | string | Yes | Unique relationship identifier |
| `from_id` | string | Yes | Source entity ID |
| `to_id` | string | Yes | Target entity ID |
| `type` | string | Yes | Relationship type (CAUSES, ENABLES, CONTRADICTS, BUILDS_UPON, RELATES_TO) |
| `strength` | float | Yes | Relationship strength 0.0-1.0 |
| `confidence` | float | Yes | Confidence in relationship 0.0-1.0 |

---

## Similarity Tools

### search-similar-thoughts

Search for thoughts similar to a query using semantic embeddings.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | Yes | Text to find similar thoughts |
| `limit` | integer | No | Maximum results (default: 5) |
| `min_similarity` | float | No | Threshold 0-1 (default: 0.5) |

---

## Graph-of-Thoughts Tools

### got-initialize

Initialize a new Graph-of-Thoughts graph with an initial thought.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Unique identifier for this graph |
| `initial_thought` | string | Yes | Starting thought content |
| `config` | object | No | GraphConfig with limits |

---

### got-generate

Generate k diverse continuations from active or specified vertices.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |
| `k` | integer | Yes | Number of continuations per source (1-10) |
| `problem` | string | Yes | Original problem context |
| `source_ids` | array | No | Specific vertices to expand from (default: active) |

---

### got-aggregate

Merge multiple parallel reasoning paths into a unified insight.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |
| `vertex_ids` | array | Yes | Array of vertices to merge (min: 2) |
| `problem` | string | Yes | Original problem context |

---

### got-refine

Iteratively improve a thought through self-critique.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |
| `vertex_id` | string | Yes | Vertex to refine |
| `problem` | string | Yes | Original problem context |

---

### got-score

Evaluate thought quality with multi-criteria breakdown.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |
| `vertex_id` | string | Yes | Vertex to score |
| `problem` | string | Yes | Original problem context |

---

### got-prune

Remove low-quality vertices below threshold (preserves roots and terminals).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |
| `threshold` | float | No | Minimum score to keep (default: config.PruneThreshold) |

---

### got-get-state

Get current graph state with all vertices and metadata.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |

---

### got-finalize

Mark terminal vertices and retrieve final conclusions.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `graph_id` | string | Yes | Graph identifier |
| `terminal_ids` | array | Yes | Array of final conclusion vertex IDs |

---

## Claude Code Optimization Tools

### export-session

Export current reasoning session to a portable JSON format for backup, sharing, or later restoration.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `session_id` | string | No | Session identifier (default: "default") |
| `include_decisions` | boolean | No | Include decision records (default: true) |
| `include_causal_graphs` | boolean | No | Include causal graph data (default: true) |
| `compress` | boolean | No | Gzip compress the output (default: false) |

---

### import-session

Import a previously exported reasoning session with merge strategy control.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `export_data` | string | Yes | JSON string from export-session |
| `merge_strategy` | string | No | "replace" (clear existing), "merge" (update/add), "append" (keep existing, add new) |
| `validate_only` | boolean | No | Check validity without importing (default: false) |
| `preserve_timestamps` | boolean | No | Keep original timestamps (default: true) |

---

### list-presets

List available workflow presets for common development tasks.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `category` | string | No | Filter by category (code, architecture, research, testing, documentation, operations) |

---

### run-preset

Execute a workflow preset with provided inputs.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `preset_id` | string | Yes | ID of the preset to run |
| `input` | object | Yes | Input values matching preset's input_schema |
| `dry_run` | boolean | No | Preview steps without executing (default: false) |
| `step_by_step` | boolean | No | Pause after each step (default: false) |

---

### format-response

Apply format optimization to reduce response size.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `response` | any | Yes | The response object to format |
| `level` | string | No | Format level - "full" (default), "compact" (40-60% reduction), "minimal" (80%+ reduction) |

---

## Error Handling

All tools return errors in the following format:

```json
{
  "error": "Error message describing what went wrong",
  "status": "error"
}
```

Common error types:
- **Validation errors**: Missing required parameters or invalid values
- **Not found errors**: Referenced ID (thought, branch, graph, etc.) does not exist
- **Processing errors**: Internal processing failures

---

## Version Information

- **Server Version**: 1.0.0
- **MCP SDK Version**: 0.8.0
- **Total Tools**: 80
- **Last Updated**: 2025-12-01
