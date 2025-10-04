# Unified Thinking MCP Server - Comprehensive Test Plan

**Version:** 1.0
**Date:** 2025-10-03
**Status:** Draft

## Executive Summary

This document outlines a comprehensive test plan to validate the accuracy and effectiveness of all 31 tools in the Unified Thinking MCP server. The plan focuses on measuring correctness, identifying weaknesses, and establishing improvement priorities.

**Current Test Coverage:** 73.6% overall
- Storage: 80.5%
- Validation: 94.2%
- Analysis (contradiction, evidence): High
- Analysis (argument): 0.0% ⚠️
- Reasoning (causal, probabilistic, temporal): Partial
- Integration: Partial
- Orchestration: 0.0% ⚠️

---

## 1. Tool Inventory & Categorization

### 1.1 Core Thinking Tools (4 tools)
**Priority: Critical**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `think` | Multi-mode thinking (linear, tree, divergent, auto) | Core | High |
| `history` | View thinking history with filtering | Query | High |
| `search` | Full-text search across thoughts | Query | High |
| `get-metrics` | System performance metrics | Monitoring | High |

### 1.2 Branch Management Tools (3 tools)
**Priority: Critical**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `list-branches` | List all thinking branches | Query | High |
| `focus-branch` | Switch active branch | State Management | High |
| `branch-history` | Detailed branch history with insights | Query | High |
| `recent-branches` | Recently accessed branches | Query | High |

### 1.3 Logical Validation Tools (3 tools)
**Priority: Critical**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `validate` | Logical consistency validation | Validation | 94.2% |
| `prove` | Formal logical proof attempts | Validation | 94.2% |
| `check-syntax` | Syntax validation for logical statements | Validation | 94.2% |

### 1.4 Probabilistic Reasoning Tools (2 tools)
**Priority: High**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `probabilistic-reasoning` | Bayesian inference and belief updates | Reasoning | Partial |
| `assess-evidence` | Evidence quality and reliability assessment | Analysis | High |

### 1.5 Decision Analysis Tools (3 tools)
**Priority: High**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `make-decision` | Multi-criteria decision framework | Reasoning | Partial |
| `decompose-problem` | Break complex problems into subproblems | Reasoning | Partial |
| `sensitivity-analysis` | Test robustness to assumption changes | Analysis | Partial |

### 1.6 Metacognition Tools (3 tools)
**Priority: High**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `self-evaluate` | Metacognitive self-assessment | Metacognition | Partial |
| `detect-biases` | Cognitive bias detection | Metacognition | Partial |
| `detect-contradictions` | Find contradictions in reasoning | Analysis | High |

### 1.7 Multi-Perspective Analysis Tools (1 tool)
**Priority: Medium**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `analyze-perspectives` | Stakeholder perspective analysis | Analysis | Partial |

### 1.8 Temporal Reasoning Tools (3 tools)
**Priority: Medium**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `analyze-temporal` | Short-term vs long-term implications | Reasoning | Partial |
| `compare-time-horizons` | Compare across time horizons | Reasoning | Partial |
| `identify-optimal-timing` | Determine optimal timing | Reasoning | Partial |

### 1.9 Causal Reasoning Tools (5 tools)
**Priority: High**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `build-causal-graph` | Construct causal graphs from observations | Reasoning | Partial |
| `simulate-intervention` | Simulate intervention effects | Reasoning | Partial |
| `generate-counterfactual` | Generate "what if" scenarios | Reasoning | Partial |
| `analyze-correlation-vs-causation` | Distinguish correlation from causation | Reasoning | Partial |
| `get-causal-graph` | Retrieve causal graph by ID | Query | Partial |

### 1.10 Cross-Mode Synthesis Tools (2 tools)
**Priority: High**

| Tool | Purpose | Category | Current Coverage |
|------|---------|----------|------------------|
| `synthesize-insights` | Synthesize insights from multiple modes | Integration | Partial |
| `detect-emergent-patterns` | Detect emergent patterns | Integration | Partial |

---

## 2. Accuracy Validation Approach

### 2.1 Measurement Framework

#### 2.1.1 Quantitative Metrics

1. **Logical Reasoning Tools**
   - **Correctness Rate:** % of correct logical inferences
   - **False Positive Rate:** % of invalid arguments accepted
   - **False Negative Rate:** % of valid arguments rejected
   - **Baseline:** Compare against formal logic textbook examples

2. **Probabilistic Reasoning Tools**
   - **Bayes Theorem Accuracy:** Max 0.01 error tolerance for P(H|E) calculations
   - **Probability Coherence:** Sum of complementary events = 1.0
   - **Benchmark:** Known probability problems (Monty Hall, medical testing, etc.)

3. **Causal Reasoning Tools**
   - **Causal Graph Precision:** % of correct causal links identified
   - **Causal Graph Recall:** % of actual causal links identified
   - **Intervention Accuracy:** Correct prediction of downstream effects
   - **Benchmark:** Pearl's causal inference examples

4. **Fallacy Detection Tools**
   - **Precision:** True fallacies / All detected fallacies
   - **Recall:** Detected fallacies / All actual fallacies
   - **F1 Score:** Harmonic mean of precision and recall
   - **Benchmark:** Labeled fallacy corpus (40+ fallacy types)

5. **Argument Analysis Tools**
   - **Premise Extraction Accuracy:** % premises correctly identified
   - **Hidden Assumption Detection:** % unstated assumptions found
   - **Argument Strength Correlation:** Agreement with expert ratings (r > 0.7)

### 2.1.2 Qualitative Metrics

1. **Explanatory Quality**
   - Clarity: Are explanations understandable?
   - Completeness: Do explanations cover key aspects?
   - Actionability: Can users act on the insights?

2. **Consistency**
   - Temporal: Same input → same output
   - Cross-tool: Tools produce compatible results
   - Logical: No internal contradictions

3. **Robustness**
   - Edge cases: Handles unusual inputs gracefully
   - Error recovery: Provides useful error messages
   - Input variation: Handles synonyms, paraphrasing

### 2.2 Performance Baselines

#### Gold Standard Test Sets

1. **Logical Reasoning** (100 examples)
   - Valid/invalid syllogisms
   - Modus ponens/tollens
   - Formal fallacies
   - Source: Logic textbooks, philosophy papers

2. **Probabilistic Reasoning** (50 examples)
   - Classic probability problems
   - Bayesian update scenarios
   - Base rate neglect cases
   - Source: Statistical inference textbooks

3. **Causal Reasoning** (75 examples)
   - Established causal relationships (smoking→cancer)
   - Confounding scenarios
   - Intervention predictions
   - Source: Pearl's "Causality" book

4. **Fallacy Detection** (200 examples)
   - 40+ fallacy types
   - Each with 3-5 examples
   - Both clear and borderline cases
   - Source: Fallacy databases, critical thinking texts

5. **Argument Analysis** (60 examples)
   - Real-world arguments (op-eds, debates)
   - Expert-annotated with premises, conclusions
   - Strength ratings from philosophy experts
   - Source: Argument mining corpora

---

## 3. Test Scenarios by Category

### 3.1 Logical Reasoning Tests

#### 3.1.1 Formal Logic Tests (validate, prove, check-syntax)

**Test Cases:**

```javascript
// Test Case LR-001: Valid Modus Ponens
{
  premises: ["If it rains, the ground is wet", "It is raining"],
  conclusion: "The ground is wet",
  expected: {
    is_provable: true,
    is_valid: true,
    confidence: >= 0.9
  }
}

// Test Case LR-002: Invalid Affirming Consequent
{
  premises: ["If P then Q", "Q is true"],
  conclusion: "P is true",
  expected: {
    is_provable: false,
    fallacy_detected: "affirming_consequent",
    confidence: >= 0.7
  }
}

// Test Case LR-003: Valid Modus Tollens
{
  premises: ["If P then Q", "Not Q"],
  conclusion: "Not P",
  expected: {
    is_provable: true,
    is_valid: true
  }
}

// Test Case LR-004: Undistributed Middle
{
  premises: ["All cats are mammals", "All dogs are mammals"],
  conclusion: "All cats are dogs",
  expected: {
    is_provable: false,
    fallacy_detected: "undistributed_middle"
  }
}

// Test Case LR-005: Syllogism Validity
{
  premises: ["All humans are mortal", "Socrates is human"],
  conclusion: "Socrates is mortal",
  expected: {
    is_provable: true,
    steps: ["universal_instantiation", "modus_ponens"]
  }
}
```

**Success Criteria:**
- 95%+ accuracy on valid/invalid classification
- Correct fallacy identification for 90%+ of formal fallacies
- Clear explanation of why argument is valid/invalid

#### 3.1.2 Fallacy Detection Tests (detect-biases, validate)

**Test Cases:**

```javascript
// Test Case FD-001: Ad Hominem
{
  content: "Your argument is wrong because you're an idiot",
  expected: {
    fallacy: "ad_hominem",
    confidence: >= 0.8,
    explanation: includes("attacking person's character")
  }
}

// Test Case FD-002: Straw Man
{
  content: "They claim we should regulate banks. They want to destroy free markets!",
  expected: {
    fallacy: "straw_man",
    confidence: >= 0.6,
    explanation: includes("misrepresenting opponent")
  }
}

// Test Case FD-003: False Dilemma
{
  content: "Either you're with us or you're against us",
  expected: {
    fallacy: "false_dilemma",
    confidence: >= 0.8
  }
}

// Test Case FD-004: Slippery Slope
{
  content: "If we allow gay marriage, next we'll allow people to marry animals",
  expected: {
    fallacy: "slippery_slope",
    confidence: >= 0.7
  }
}

// Test Case FD-005: Appeal to Authority (illegitimate)
{
  content: "Einstein believed in God, so atheism must be wrong",
  expected: {
    fallacy: "appeal_to_authority",
    explanation: includes("outside their area of expertise")
  }
}

// Test Case FD-006: Base Rate Neglect
{
  content: "The test is 99% accurate, so if it's positive, I definitely have the disease",
  expected: {
    fallacy: "base_rate_neglect",
    category: "statistical"
  }
}

// Test Case FD-007: Multiple Fallacies
{
  content: "Climate scientists are idiots. They claim warming is happening, but it was cold last winter!",
  expected: {
    fallacies: ["ad_hominem", "hasty_generalization"],
    count: >= 2
  }
}
```

**Success Criteria:**
- Precision >= 0.80 (few false positives)
- Recall >= 0.75 (catch most fallacies)
- F1 Score >= 0.77
- Handle 40+ fallacy types
- Detect multiple fallacies in single text

### 3.2 Probabilistic Reasoning Tests

#### 3.2.1 Bayesian Inference (probabilistic-reasoning)

**Test Cases:**

```javascript
// Test Case PR-001: Simple Bayes Update
{
  operation: "create",
  statement: "Patient has disease",
  prior_prob: 0.01,  // 1% base rate
  then: {
    operation: "update",
    likelihood: 0.95,  // P(positive test | disease)
    evidence_prob: 0.05,  // P(positive test)
  },
  expected: {
    posterior: 0.19,  // P(disease | positive test) ≈ 19%
    tolerance: 0.01
  }
}

// Test Case PR-002: Multiple Evidence Updates
{
  initial: { statement: "Hypothesis H", prior: 0.5 },
  updates: [
    { evidence: "E1", likelihood: 0.8, evidence_prob: 0.6 },
    { evidence: "E2", likelihood: 0.9, evidence_prob: 0.5 }
  ],
  expected: {
    final_probability: calculate_sequential_bayes(0.5, updates),
    monotonic: true  // Probability should move monotonically
  }
}

// Test Case PR-003: Combine Independent Beliefs (AND)
{
  beliefs: [
    { id: "B1", probability: 0.8 },
    { id: "B2", probability: 0.9 }
  ],
  operation: "and",
  expected: {
    combined: 0.72,  // 0.8 * 0.9
    tolerance: 0.001
  }
}

// Test Case PR-004: Combine Independent Beliefs (OR)
{
  beliefs: [
    { id: "B1", probability: 0.3 },
    { id: "B2", probability: 0.4 }
  ],
  operation: "or",
  expected: {
    combined: 0.58,  // 1 - (1-0.3)*(1-0.4) = 1 - 0.42
    tolerance: 0.001
  }
}

// Test Case PR-005: Probability Coherence
{
  belief: { statement: "It will rain", probability: 0.7 },
  complement: { statement: "It will not rain" },
  expected: {
    sum: 1.0,
    tolerance: 0.001
  }
}

// Test Case PR-006: Edge Cases
{
  test_cases: [
    { prior: 0.0, likelihood: 0.5, evidence_prob: 0.5, expected: 0.0 },
    { prior: 1.0, likelihood: 0.5, evidence_prob: 0.5, expected: 1.0 },
    { prior: 0.5, likelihood: 1.0, evidence_prob: 0.5, expected: 1.0 }
  ]
}
```

**Success Criteria:**
- Bayes theorem calculations accurate to 0.01
- Handles edge cases (0, 1 probabilities)
- Probability values always in [0, 1]
- Evidence tracking maintains history
- Multiple updates produce coherent results

#### 3.2.2 Evidence Assessment (assess-evidence)

**Test Cases:**

```javascript
// Test Case EA-001: High Quality Evidence
{
  content: "Peer-reviewed study published in Nature, n=10,000, RCT design",
  source: "Nature journal",
  claim_id: "claim_1",
  supports_claim: true,
  expected: {
    quality: >= 0.85,
    reliability: >= 0.90,
    relevance: >= 0.80,
    overall_score: >= 0.85
  }
}

// Test Case EA-002: Low Quality Evidence
{
  content: "My friend told me it works",
  source: "anecdote",
  supports_claim: true,
  expected: {
    quality: <= 0.40,
    reliability: <= 0.30,
    overall_score: <= 0.35
  }
}

// Test Case EA-003: Conflicting Evidence
{
  evidence_set: [
    { content: "Study A shows effect", score: 0.8, supports: true },
    { content: "Study B shows no effect", score: 0.8, supports: false }
  ],
  expected: {
    conflict_detected: true,
    overall_confidence: "low"
  }
}
```

**Success Criteria:**
- Distinguishes high/low quality evidence
- Quality scores correlate with evidence hierarchy (RCT > observational > anecdote)
- Reliability assessment considers source credibility
- Overall score combines quality, reliability, relevance

### 3.3 Causal Reasoning Tests

#### 3.3.1 Causal Graph Construction (build-causal-graph)

**Test Cases:**

```javascript
// Test Case CG-001: Simple Causal Chain
{
  description: "Smoking and lung cancer",
  observations: [
    "Smoking causes lung cancer",
    "Lung cancer increases mortality risk",
    "Smoking is associated with tar exposure"
  ],
  expected: {
    variables: ["smoking", "lung cancer", "mortality", "tar exposure"],
    links: [
      { from: "smoking", to: "lung cancer", type: "positive" },
      { from: "lung cancer", to: "mortality", type: "positive" },
      { from: "smoking", to: "tar exposure", type: "positive" }
    ],
    variable_count: 4,
    link_count: 3
  }
}

// Test Case CG-002: Confounding Variable
{
  description: "Ice cream and drowning",
  observations: [
    "Ice cream sales correlate with drowning rates",
    "Hot weather increases ice cream sales",
    "Hot weather leads to more swimming"
  ],
  expected: {
    confounding_variable: "hot weather",
    spurious_correlation: { from: "ice cream", to: "drowning" },
    explanation: includes("common cause")
  }
}

// Test Case CG-003: Variable Type Inference
{
  observations: [
    "Temperature increases ice melting rate",
    "Treatment was either given or withheld",
    "Patient status was critical, stable, or improving"
  ],
  expected: {
    variable_types: {
      temperature: "continuous",
      treatment: "binary",
      "patient status": "categorical"
    }
  }
}
```

**Success Criteria:**
- Correctly extracts variables from natural language
- Identifies causal relationships vs correlations
- Infers variable types (continuous, binary, categorical)
- Detects confounding variables
- Estimates link strength and confidence

#### 3.3.2 Intervention Simulation (simulate-intervention)

**Test Cases:**

```javascript
// Test Case IS-001: Simple Intervention
{
  graph_id: "smoking_graph",
  variable: "smoking",
  intervention: "decrease",
  expected: {
    predicted_effects: [
      { variable: "lung cancer", effect: "decrease", probability: >= 0.8 },
      { variable: "mortality", effect: "decrease", probability: >= 0.6 }
    ],
    confidence: >= 0.7
  }
}

// Test Case IS-002: Cascade Effects
{
  graph_id: "economic_model",
  variable: "interest_rate",
  intervention: "increase",
  expected: {
    immediate_effects: ["borrowing"],
    downstream_effects: ["investment", "employment", "GDP"],
    path_lengths: { borrowing: 1, investment: 2, employment: 3, GDP: 4 }
  }
}

// Test Case IS-003: Negative Feedback Loop
{
  graph: {
    variables: ["A", "B", "C"],
    links: [
      { from: "A", to: "B", type: "positive" },
      { from: "B", to: "C", type: "positive" },
      { from: "C", to: "A", type: "negative" }
    ]
  },
  intervention: { variable: "A", type: "increase" },
  expected: {
    stabilizing_effect: true,
    explanation: includes("negative feedback")
  }
}
```

**Success Criteria:**
- Correctly traces downstream effects
- Handles multi-step causal chains
- Estimates effect magnitude and probability
- Detects feedback loops
- Explains causal mechanism

#### 3.3.3 Counterfactual Reasoning (generate-counterfactual)

**Test Cases:**

```javascript
// Test Case CF-001: Historical Counterfactual
{
  graph_id: "ww2_causes",
  scenario: "What if Treaty of Versailles was more lenient?",
  changes: {
    "war_reparations": "low",
    "territorial_losses": "minimal"
  },
  expected: {
    outcomes: {
      "german_resentment": "low",
      "economic_hardship": "moderate",
      "rise_of_extremism": "less_likely"
    },
    plausibility: >= 0.6
  }
}

// Test Case CF-002: Personal Decision
{
  scenario: "What if I had taken the job offer?",
  changes: { "job": "company_A" },
  expected: {
    outcomes: includes(["income", "location", "work_satisfaction"]),
    plausibility: <= 0.8,  // Personal decisions have uncertainty
    comparison: "current_state_vs_counterfactual"
  }
}
```

**Success Criteria:**
- Generates plausible alternative scenarios
- Traces implications of changes
- Estimates plausibility
- Maintains logical consistency

#### 3.3.4 Correlation vs Causation (analyze-correlation-vs-causation)

**Test Cases:**

```javascript
// Test Case CC-001: Clear Causation
{
  observation: "Pressing the brake pedal causes the car to slow down",
  expected: {
    relationship: "causal",
    criteria_met: ["temporal_precedence", "mechanism", "experimental_evidence"],
    confidence: >= 0.9
  }
}

// Test Case CC-002: Spurious Correlation
{
  observation: "Countries with more Nobel laureates consume more chocolate",
  expected: {
    relationship: "correlation",
    warning: includes("confounding"),
    alternative_explanations: includes(["wealth", "education", "research_funding"])
  }
}

// Test Case CC-003: Reverse Causation
{
  observation: "Hospitals are associated with death",
  expected: {
    warning: includes("reverse causation"),
    explanation: "Sick people go to hospitals, not hospitals make people sick"
  }
}
```

**Success Criteria:**
- Distinguishes causation from correlation
- Identifies confounding variables
- Detects reverse causation
- Requires evidence for causal claims
- Provides actionable guidance

### 3.4 Argument Analysis Tests

#### 3.4.1 Argument Decomposition (argument analyzer - not yet exposed as tool)

**Test Cases:**

```javascript
// Test Case AD-001: Simple Deductive Argument
{
  text: "All men are mortal. Socrates is a man. Therefore, Socrates is mortal.",
  expected: {
    main_claim: "Socrates is mortal",
    premises: [
      { statement: "All men are mortal", type: "factual", certainty: >= 0.9 },
      { statement: "Socrates is a man", type: "factual", certainty: >= 0.9 }
    ],
    argument_type: "deductive",
    strength: >= 0.9,
    hidden_assumptions: []
  }
}

// Test Case AD-002: Inductive Argument
{
  text: "Every swan I've seen is white. Therefore, all swans are white.",
  expected: {
    argument_type: "inductive",
    strength: <= 0.6,
    vulnerabilities: includes("hasty_generalization"),
    hidden_assumptions: includes("sample_is_representative")
  }
}

// Test Case AD-003: Argument with Hidden Assumptions
{
  text: "We should ban guns because they cause violence.",
  expected: {
    premises: [{ statement: matches("guns.*violence") }],
    hidden_assumptions: includes([
      "causal_relationship",
      "value_assumption",
      "no_alternative_solutions"
    ])
  }
}

// Test Case AD-004: Weak Premise Detection
{
  text: "Some people say X is true. My friend agrees. Therefore X is definitely true.",
  expected: {
    weak_premises: count >= 2,
    strength: <= 0.4,
    vulnerabilities: includes(["appeal_to_authority", "weak_evidence"])
  }
}
```

**Success Criteria:**
- Extracts premises accurately (90%+ recall)
- Identifies main claim correctly
- Detects argument type (deductive/inductive/abductive)
- Finds hidden assumptions
- Calculates reasonable strength estimates
- Identifies vulnerabilities

#### 3.4.2 Counter-Argument Generation

**Test Cases:**

```javascript
// Test Case CA-001: Deny Premise Strategy
{
  argument_id: "arg_123",
  expected: {
    counter_arguments: includes({
      strategy: "deny_premise",
      targets: "weakest_premise",
      strength: proportional_to_premise_weakness
    })
  }
}

// Test Case CA-002: Break Inference Link
{
  argument_id: "arg_456",
  expected: {
    counter_arguments: includes({
      strategy: "break_link",
      explanation: "conclusion doesn't necessarily follow"
    })
  }
}

// Test Case CA-003: Alternative Explanation (Abductive)
{
  argument_id: "arg_789",  // Abductive argument
  expected: {
    counter_arguments: includes({
      strategy: "alternative_explanation",
      criteria: ["simpler", "better_evidence_fit"]
    })
  }
}
```

**Success Criteria:**
- Generates 3-4 counter-arguments per argument
- Uses appropriate strategy for argument type
- Targets actual weaknesses
- Provides supporting reasoning

### 3.5 Decision Analysis Tests

#### 3.5.1 Multi-Criteria Decision Making (make-decision)

**Test Cases:**

```javascript
// Test Case DM-001: Job Selection
{
  question: "Which job offer should I accept?",
  options: [
    { id: "job_a", name: "Startup", scores: { salary: 7, growth: 9, stability: 4, culture: 8 } },
    { id: "job_b", name: "Corporate", scores: { salary: 9, growth: 5, stability: 9, culture: 6 } }
  ],
  criteria: [
    { id: "salary", weight: 0.3, maximize: true },
    { id: "growth", weight: 0.4, maximize: true },
    { id: "stability", weight: 0.2, maximize: true },
    { id: "culture", weight: 0.1, maximize: true }
  ],
  expected: {
    recommended: "job_a",  // Weighted score: 7.5 vs 7.0
    confidence: >= 0.7,
    sensitivity: "growth_criterion_is_decisive"
  }
}

// Test Case DM-002: Trade-off Analysis
{
  options: [
    { id: "opt_1", pros: ["pro1", "pro2"], cons: ["con1"] },
    { id: "opt_2", pros: ["pro3"], cons: ["con2", "con3"] }
  ],
  expected: {
    trade_offs: identified,
    comparison: "pros_cons_balanced"
  }
}
```

**Success Criteria:**
- Correctly applies weighted scoring
- Identifies recommended option
- Explains trade-offs
- Performs sensitivity analysis
- Handles conflicting criteria

#### 3.5.2 Problem Decomposition (decompose-problem)

**Test Cases:**

```javascript
// Test Case PD-001: Complex Problem
{
  problem: "Reduce company carbon footprint by 50% in 2 years",
  expected: {
    subproblems: [
      { id: "sub_1", description: matches("measure.*emissions"), dependencies: [] },
      { id: "sub_2", description: matches("identify.*sources"), dependencies: ["sub_1"] },
      { id: "sub_3", description: matches("evaluate.*reduction"), dependencies: ["sub_2"] },
      { id: "sub_4", description: matches("implement.*changes"), dependencies: ["sub_3"] }
    ],
    complexity: "high",
    estimated_steps: >= 4
  }
}

// Test Case PD-002: Dependency Detection
{
  problem: "Launch new product",
  expected: {
    dependency_graph: has_topological_order,
    parallel_tasks: identified,
    critical_path: calculated
  }
}
```

**Success Criteria:**
- Breaks problem into logical subproblems
- Identifies dependencies correctly
- Provides actionable steps
- Estimates complexity

#### 3.5.3 Sensitivity Analysis (sensitivity-analysis)

**Test Cases:**

```javascript
// Test Case SA-001: Robust Conclusion
{
  target_claim: "Investment A is better than B",
  assumptions: [
    "Market grows at 5% annually",
    "Interest rates remain stable",
    "No major disruptions"
  ],
  base_confidence: 0.8,
  expected: {
    robustness: "high",
    critical_assumptions: [],
    confidence_range: [0.75, 0.85]  // Small variation
  }
}

// Test Case SA-002: Sensitive Conclusion
{
  target_claim: "Product will be profitable",
  assumptions: [
    "Achieve 10,000 sales in year 1",  // Critical
    "Production costs below $50",       // Critical
    "No major competitors"              // Critical
  ],
  base_confidence: 0.7,
  expected: {
    robustness: "low",
    critical_assumptions: count >= 2,
    confidence_range: [0.3, 0.9],  // Large variation
    recommendation: "gather_more_data"
  }
}
```

**Success Criteria:**
- Identifies critical assumptions
- Quantifies sensitivity
- Provides confidence ranges
- Recommends actions (e.g., gather more data)

### 3.6 Metacognition Tests

#### 3.6.1 Self-Evaluation (self-evaluate)

**Test Cases:**

```javascript
// Test Case SE-001: High Quality Thought
{
  thought_id: "thought_high_quality",
  thought_content: {
    content: "Well-structured analysis with evidence, considers alternatives, acknowledges limitations",
    confidence: 0.8,
    key_points: ["point1", "point2", "point3"]
  },
  expected: {
    completeness: >= 0.8,
    coherence: >= 0.8,
    confidence_calibration: "appropriate",
    overall_quality: >= 0.75
  }
}

// Test Case SE-002: Overconfident Thought
{
  thought_id: "thought_overconfident",
  thought_content: {
    content: "Simple assertion without evidence",
    confidence: 0.95  // Too high for limited reasoning
  },
  expected: {
    confidence_calibration: "overconfident",
    recommendations: includes("provide_evidence"),
    quality_score: <= 0.5
  }
}

// Test Case SE-003: Incomplete Reasoning
{
  thought_id: "thought_incomplete",
  thought_content: {
    content: "Started analysis but missing key considerations",
    key_points: ["point1"]  // Only 1 point
  },
  expected: {
    completeness: <= 0.5,
    missing_elements: includes(["alternative_views", "evidence", "limitations"]),
    recommendations: actionable
  }
}
```

**Success Criteria:**
- Detects overconfidence
- Identifies incomplete reasoning
- Assesses coherence
- Provides actionable recommendations

#### 3.6.2 Bias Detection (detect-biases)

**Test Cases:**

```javascript
// Test Case BD-001: Confirmation Bias
{
  thought_content: "All the evidence supports my initial hypothesis. No contradicting data was found.",
  expected: {
    biases: includes({
      type: "confirmation_bias",
      confidence: >= 0.7,
      mitigation: "actively_seek_disconfirming_evidence"
    })
  }
}

// Test Case BD-002: Anchoring Bias
{
  thought_content: "The first estimate was $100, so my final estimate of $105 seems reasonable.",
  expected: {
    biases: includes({
      type: "anchoring_bias",
      evidence: "relying_on_initial_value"
    })
  }
}

// Test Case BD-003: Availability Heuristic
{
  thought_content: "Plane crashes are common because I've seen several in the news recently.",
  expected: {
    biases: includes({
      type: "availability_heuristic",
      correction: "use_base_rates"
    })
  }
}

// Test Case BD-004: Multiple Biases
{
  thought_content: "My friend succeeded with this method (sample_size_1), so it must work for everyone. Besides, everyone I know agrees with me.",
  expected: {
    biases: count >= 2,
    includes: ["hasty_generalization", "selection_bias"]
  }
}
```

**Success Criteria:**
- Detects 10+ cognitive bias types
- Provides mitigation strategies
- Explains why bias is present
- Handles multiple biases

### 3.7 Integration Tests

#### 3.7.1 Cross-Mode Synthesis (synthesize-insights)

**Test Cases:**

```javascript
// Test Case CM-001: Complementary Insights
{
  context: "Should we expand to new market?",
  inputs: [
    { mode: "causal", content: "Market size causes revenue potential", confidence: 0.8 },
    { mode: "temporal", content: "Short-term costs, long-term gains", confidence: 0.7 },
    { mode: "probabilistic", content: "70% chance of success given market data", confidence: 0.8 }
  ],
  expected: {
    synergies: count >= 2,
    conflicts: count == 0,
    integrated_recommendation: "expand_with_staged_approach",
    confidence: >= 0.7
  }
}

// Test Case CM-002: Conflicting Insights
{
  inputs: [
    { mode: "causal", content: "X causes Y", confidence: 0.8 },
    { mode: "probabilistic", content: "X and Y are independent", confidence: 0.7 }
  ],
  expected: {
    conflicts: count >= 1,
    resolution: "investigate_further",
    highlighted_contradiction: true
  }
}
```

**Success Criteria:**
- Identifies synergies between modes
- Detects conflicts
- Provides integrated view
- Handles 3+ reasoning modes

#### 3.7.2 Emergent Pattern Detection (detect-emergent-patterns)

**Test Cases:**

```javascript
// Test Case EP-001: Convergent Evidence
{
  inputs: [
    { mode: "causal", content: "Factor A influences outcome" },
    { mode: "temporal", content: "Factor A's effect increases over time" },
    { mode: "perspective", content: "All stakeholders agree Factor A is key" }
  ],
  expected: {
    patterns: includes("convergent_evidence_for_factor_A"),
    confidence_boost: true,
    explanation: "multiple_independent_lines_of_evidence"
  }
}

// Test Case EP-002: Hidden Assumptions Across Modes
{
  inputs: [
    { mode: "decision", content: "Assumes stable market" },
    { mode: "causal", content: "Assumes linear relationships" },
    { mode: "temporal", content: "Assumes no major disruptions" }
  ],
  expected: {
    patterns: includes("shared_stability_assumption"),
    vulnerability: "all_modes_fail_if_market_disrupted"
  }
}
```

**Success Criteria:**
- Finds patterns not visible in single mode
- Identifies shared assumptions
- Detects convergent/divergent evidence

---

## 4. Integration Testing

### 4.1 Workflow Orchestration (orchestration package)

**Test Cases:**

```javascript
// Test Case WF-001: Sequential Workflow
{
  workflow_type: "sequential",
  steps: [
    { tool: "decompose-problem", store_as: "decomposition" },
    { tool: "build-causal-graph", depends_on: ["decomposition"], store_as: "causal" },
    { tool: "make-decision", depends_on: ["causal"], store_as: "decision" }
  ],
  expected: {
    execution_order: ["decompose-problem", "build-causal-graph", "make-decision"],
    all_steps_executed: true,
    context_propagated: true
  }
}

// Test Case WF-002: Parallel Workflow
{
  workflow_type: "parallel",
  steps: [
    { tool: "analyze-perspectives" },
    { tool: "analyze-temporal" },
    { tool: "build-causal-graph" }
  ],
  expected: {
    concurrent_execution: true,
    all_steps_complete: true,
    no_deadlocks: true
  }
}

// Test Case WF-003: Conditional Workflow
{
  workflow_type: "conditional",
  steps: [
    { tool: "self-evaluate", store_as: "eval" },
    {
      tool: "detect-biases",
      condition: { field: "eval.quality", operator: "lt", value: 0.7 },
      only_if_low_quality: true
    }
  ],
  expected: {
    conditional_execution: true,
    step_skipping: based_on_conditions
  }
}

// Test Case WF-004: Error Handling
{
  workflow: "complex_workflow",
  inject_error: { at_step: 2, type: "tool_failure" },
  expected: {
    error_propagated: true,
    partial_results: available,
    status: "failed",
    error_message: informative
  }
}
```

**Success Criteria:**
- Sequential workflows execute in order
- Parallel workflows run concurrently
- Conditional branches work correctly
- Context propagates across steps
- Errors handled gracefully

### 4.2 Evidence Pipeline (integration/evidence_pipeline)

**Test Cases:**

```javascript
// Test Case EVP-001: Evidence Accumulation
{
  claim: "Product will succeed in market",
  evidence_sequence: [
    { content: "Market research positive", score: 0.7, supports: true },
    { content: "Competitor launched similar product", score: 0.6, supports: false },
    { content: "Beta testers love it", score: 0.8, supports: true }
  ],
  expected: {
    belief_trajectory: [0.5, 0.65, 0.55, 0.72],  // Prior → updates
    final_confidence: approximately(0.72),
    evidence_count: 3
  }
}

// Test Case EVP-002: Contradictory Evidence
{
  claim: "Climate change is real",
  evidence: [
    { content: "97% of climate scientists agree", score: 0.95, supports: true },
    { content: "Blog post disagrees", score: 0.2, supports: false }
  ],
  expected: {
    high_quality_evidence_dominant: true,
    final_confidence: >= 0.9,
    low_quality_evidence_minimal_impact: true
  }
}
```

**Success Criteria:**
- Evidence correctly updates beliefs
- High quality evidence weighted appropriately
- Contradictory evidence handled
- Bayesian updates are accurate

### 4.3 Causal-Temporal Integration

**Test Cases:**

```javascript
// Test Case CTI-001: Time-Dependent Causal Effects
{
  causal_graph: {
    variables: ["policy_change", "behavior", "outcome"],
    links: [
      { from: "policy_change", to: "behavior", lag: "short" },
      { from: "behavior", to: "outcome", lag: "long" }
    ]
  },
  temporal_analysis: {
    short_term: "costs",
    long_term: "benefits"
  },
  expected: {
    integrated_view: "short_term_pain_long_term_gain",
    optimal_timing: "implement_gradually",
    causal_path_aligns_with_temporal: true
  }
}
```

**Success Criteria:**
- Causal graphs incorporate temporal dimensions
- Temporal analysis uses causal relationships
- Recommendations consider both causal and temporal factors

---

## 5. Improvement Identification

### 5.1 Metrics to Track

#### 5.1.1 Accuracy Metrics

| Metric | Target | Current | Priority |
|--------|--------|---------|----------|
| **Logical Reasoning** |
| Valid/Invalid Classification Accuracy | 95% | TBD | Critical |
| Formal Fallacy Detection Precision | 90% | TBD | Critical |
| Formal Fallacy Detection Recall | 85% | TBD | Critical |
| **Probabilistic Reasoning** |
| Bayes Calculation Error (MAE) | < 0.01 | TBD | Critical |
| Probability Coherence | 100% | TBD | Critical |
| **Causal Reasoning** |
| Variable Extraction Recall | 80% | TBD | High |
| Causal Link Precision | 75% | TBD | High |
| Intervention Prediction Accuracy | 70% | TBD | High |
| **Fallacy Detection** |
| Overall F1 Score | 0.77 | TBD | High |
| Informal Fallacy Precision | 80% | TBD | High |
| Informal Fallacy Recall | 75% | TBD | High |
| **Argument Analysis** |
| Premise Extraction Recall | 90% | 0% (untested) | Critical |
| Hidden Assumption Detection | 60% | 0% (untested) | High |
| Strength Estimate Correlation | 0.70 | 0% (untested) | Medium |

#### 5.1.2 Quality Metrics

| Metric | Measurement | Target |
|--------|-------------|--------|
| Explanation Clarity | User comprehension rate | 80% |
| Actionability | % recommendations acted upon | 60% |
| Consistency | Test-retest reliability | 95% |
| Robustness | % edge cases handled | 85% |

### 5.2 Weak Area Identification

#### Current Gaps (Based on 0% Coverage):

1. **Argument Analysis Module** (0% coverage)
   - No tests for argument decomposition
   - No tests for counter-argument generation
   - No validation of premise extraction
   - **Action:** Create comprehensive test suite (60+ test cases)

2. **Orchestration Module** (0% coverage)
   - Workflow execution untested
   - Context propagation untested
   - Conditional logic untested
   - **Action:** Create workflow integration tests (30+ test cases)

3. **Partial Coverage Areas:**
   - Causal reasoning: Variable extraction edge cases
   - Probabilistic reasoning: Complex belief networks
   - Temporal reasoning: Long-term extrapolation
   - **Action:** Expand existing test suites

#### Potential Accuracy Issues (Hypothesized):

1. **Natural Language Understanding**
   - Causal variable extraction may miss context
   - Premise extraction may struggle with complex sentences
   - **Improvement:** Add NLP preprocessing, entity recognition

2. **Edge Case Handling**
   - Extreme probability values (0, 1)
   - Circular causal relationships
   - Deeply nested arguments
   - **Improvement:** Add boundary condition tests

3. **Quantitative Estimation**
   - Link strength estimation is heuristic
   - Argument strength calculation not validated
   - **Improvement:** Calibrate against expert judgments

### 5.3 Prioritization Framework

**Priority Matrix:**

| Priority | Criteria | Actions |
|----------|----------|---------|
| **P0 - Critical** | Core functionality, 0% coverage, correctness-critical | Argument analysis tests, Bayes calculation validation |
| **P1 - High** | Widely used, accuracy impacts decisions | Causal graph validation, fallacy detection improvement |
| **P2 - Medium** | Important but not core, has partial coverage | Temporal reasoning edge cases, bias detection expansion |
| **P3 - Low** | Nice-to-have, secondary features | UI improvements, performance optimization |

**Recommended Priorities:**

1. **Week 1-2:** Argument analysis test suite (P0)
2. **Week 3:** Orchestration integration tests (P0)
3. **Week 4-5:** Probabilistic reasoning validation (P0)
4. **Week 6-7:** Causal reasoning accuracy tests (P1)
5. **Week 8:** Fallacy detection precision/recall benchmarking (P1)
6. **Week 9-10:** Edge case coverage across all modules (P2)

---

## 6. Self-Testing Strategy

### 6.1 Meta-Validation Approaches

#### 6.1.1 Cross-Validation Between Tools

Tools can validate each other:

```javascript
// Example: Validate probabilistic reasoning with logical reasoning
{
  test: "Probability coherence check",
  steps: [
    { tool: "probabilistic-reasoning", operation: "create", belief: "P", prob: 0.7 },
    { tool: "probabilistic-reasoning", operation: "create", belief: "not P", prob: 0.3 },
    { tool: "validate", check: "P and not P should have sum = 1.0" }
  ]
}

// Example: Validate causal reasoning with contradiction detection
{
  test: "Causal consistency check",
  steps: [
    { tool: "build-causal-graph", claim: "A causes B" },
    { tool: "build-causal-graph", claim: "B causes A" },
    { tool: "detect-contradictions", expect: "circular_causation_detected" }
  ]
}
```

#### 6.1.2 Automated Test Generation

**Technique 1: Mutation Testing**
- Mutate valid arguments → should detect invalidity
- Mutate invalid arguments → should remain invalid
- Mutate probabilities → Bayes updates should reflect changes

**Technique 2: Property-Based Testing**
```go
// Example property: Bayes theorem identity
func TestBayesTheorem(t *testing.T) {
  for i := 0; i < 1000; i++ {
    prior := randomProb()
    likelihood := randomProb()
    evidenceProb := randomProb()

    posterior := bayesUpdate(prior, likelihood, evidenceProb)

    // Property: posterior must be in [0, 1]
    assert(posterior >= 0 && posterior <= 1)

    // Property: if likelihood = 1, posterior should increase
    if likelihood > prior {
      assert(posterior >= prior)
    }
  }
}
```

**Technique 3: Metamorphic Testing**
- Test transformations that should preserve properties
- Example: Reversing premise order shouldn't change validity
- Example: Paraphrasing shouldn't change fallacy detection

### 6.2 Self-Evaluation Loops

#### 6.2.1 Confidence Calibration

Track prediction accuracy vs confidence:

```javascript
{
  test: "Calibration Curve",
  method: {
    collect: "1000 predictions with confidence scores",
    bin: "Group by confidence (0-0.1, 0.1-0.2, ..., 0.9-1.0)",
    calculate: "Actual accuracy in each bin",
    ideal: "Confidence 0.7 → 70% accuracy"
  },
  improvement: {
    if: "Overconfident (confidence > accuracy)",
    then: "Reduce confidence scores",
    if: "Underconfident (confidence < accuracy)",
    then: "Increase confidence scores"
  }
}
```

#### 6.2.2 Error Analysis Loop

```javascript
{
  workflow: "Continuous Improvement",
  steps: [
    { collect: "Failed test cases" },
    { analyze: "Common failure patterns" },
    { categorize: "Error types (parsing, logic, edge cases)" },
    { prioritize: "By frequency and impact" },
    { implement: "Targeted improvements" },
    { retest: "Verify fixes don't break existing tests" }
  ]
}
```

### 6.3 Feedback Mechanisms

#### 6.3.1 Human-in-the-Loop Validation

```javascript
{
  test: "Expert Review",
  frequency: "Monthly",
  sample: "Random sample of 100 predictions",
  experts: "Domain experts (logicians, statisticians, philosophers)",
  metrics: {
    agreement: "Expert judgment vs tool output",
    target: "> 80% agreement"
  },
  feedback_loop: "Incorporate expert corrections into training data"
}
```

#### 6.3.2 A/B Testing for Improvements

```javascript
{
  test: "Algorithm Improvement",
  setup: {
    version_A: "Current algorithm",
    version_B: "Improved algorithm"
  },
  test_set: "Gold standard examples",
  metrics: ["accuracy", "precision", "recall", "F1"],
  decision_rule: "Deploy B if F1 improvement > 5% and no accuracy regression"
}
```

---

## 7. Test Implementation Plan

### 7.1 Phase 1: Foundation (Weeks 1-3)

#### Week 1: Test Infrastructure
- [ ] Set up test data repository
- [ ] Create gold standard test sets
  - [ ] 100 logical reasoning examples
  - [ ] 50 probabilistic reasoning examples
  - [ ] 75 causal reasoning examples
  - [ ] 200 fallacy examples
  - [ ] 60 argument analysis examples
- [ ] Implement test harness for batch testing
- [ ] Create result aggregation and reporting

#### Week 2: Argument Analysis Tests (P0)
- [ ] Test suite for premise extraction (20 cases)
- [ ] Test suite for hidden assumption detection (15 cases)
- [ ] Test suite for argument strength calculation (15 cases)
- [ ] Test suite for counter-argument generation (10 cases)
- [ ] **Target:** 80% coverage of argument.go

#### Week 3: Orchestration Tests (P0)
- [ ] Sequential workflow tests (10 cases)
- [ ] Parallel workflow tests (8 cases)
- [ ] Conditional workflow tests (8 cases)
- [ ] Error handling tests (4 cases)
- [ ] **Target:** 80% coverage of orchestration package

### 7.2 Phase 2: Accuracy Validation (Weeks 4-7)

#### Week 4: Probabilistic Reasoning (P0)
- [ ] Bayes theorem calculation tests (20 cases)
- [ ] Belief combination tests (10 cases)
- [ ] Edge case tests (10 cases)
- [ ] **Target:** Mean Absolute Error < 0.01

#### Week 5: Logical Reasoning (P0)
- [ ] Valid/invalid argument classification (40 cases)
- [ ] Formal fallacy detection (30 cases)
- [ ] Proof generation tests (20 cases)
- [ ] **Target:** 95% accuracy

#### Week 6: Causal Reasoning (P1)
- [ ] Variable extraction tests (25 cases)
- [ ] Causal link identification (20 cases)
- [ ] Intervention simulation (15 cases)
- [ ] Counterfactual generation (15 cases)
- [ ] **Target:** 75% precision, 80% recall

#### Week 7: Fallacy Detection (P1)
- [ ] Informal fallacy tests (160 cases, 40 types × 4 each)
- [ ] Statistical fallacy tests (40 cases)
- [ ] **Target:** F1 score ≥ 0.77

### 7.3 Phase 3: Integration & Edge Cases (Weeks 8-10)

#### Week 8: Integration Testing
- [ ] Cross-mode synthesis tests (15 cases)
- [ ] Evidence pipeline tests (10 cases)
- [ ] Causal-temporal integration (8 cases)
- [ ] **Target:** All integration workflows functional

#### Week 9: Edge Cases
- [ ] Boundary value tests (30 cases across all tools)
- [ ] Malformed input tests (20 cases)
- [ ] Performance tests (large inputs)
- [ ] **Target:** 85% edge case coverage

#### Week 10: Regression & Documentation
- [ ] Full regression test suite (all 400+ test cases)
- [ ] Performance benchmarking
- [ ] Test documentation
- [ ] Coverage report
- [ ] **Target:** 90% overall coverage

### 7.4 Phase 4: Continuous Improvement (Ongoing)

#### Monthly Activities
- [ ] Run full test suite
- [ ] Analyze failures
- [ ] Expert review of random sample (100 cases)
- [ ] Update test data with new examples
- [ ] Calibration analysis

#### Quarterly Activities
- [ ] Benchmark against state-of-the-art
- [ ] User feedback analysis
- [ ] Major algorithm improvements
- [ ] Test suite expansion

---

## 8. Success Criteria & Acceptance

### 8.1 Minimum Acceptable Performance

| Category | Metric | Target | Minimum |
|----------|--------|--------|---------|
| **Logical Reasoning** | Accuracy | 95% | 90% |
| **Probabilistic Reasoning** | MAE | < 0.01 | < 0.02 |
| **Causal Reasoning** | F1 Score | 0.80 | 0.70 |
| **Fallacy Detection** | F1 Score | 0.77 | 0.70 |
| **Argument Analysis** | Premise Recall | 90% | 80% |
| **Overall Coverage** | Line Coverage | 90% | 80% |
| **Integration** | Workflow Success | 95% | 90% |

### 8.2 Quality Gates

**Before Production Deployment:**
1. All P0 tests passing (100%)
2. All P1 tests passing (≥ 95%)
3. No critical bugs
4. Test coverage ≥ 80%
5. Performance benchmarks met
6. Expert review approval (≥ 80% agreement)

**Before Major Release:**
1. All tests passing (≥ 98%)
2. Test coverage ≥ 90%
3. No high-priority bugs
4. Regression tests passing
5. Documentation complete

### 8.3 Continuous Monitoring

**Real-time Metrics:**
- Tool usage statistics
- Error rates by tool
- Confidence calibration curves
- User feedback scores

**Alerting Thresholds:**
- Error rate > 5% → Investigation
- Confidence miscalibration > 15% → Recalibration
- User satisfaction < 70% → Review

---

## 9. Test Data Repository

### 9.1 Gold Standard Sources

1. **Logical Reasoning**
   - Copi & Cohen, "Introduction to Logic" (14th ed.)
   - Hurley, "A Concise Introduction to Logic" (13th ed.)
   - Stanford Encyclopedia of Philosophy

2. **Probabilistic Reasoning**
   - Kahneman & Tversky, probability problem sets
   - Gelman et al., "Bayesian Data Analysis"
   - Classic problems: Monty Hall, medical testing, etc.

3. **Causal Reasoning**
   - Pearl, "Causality" (2nd ed.)
   - Morgan & Winship, "Counterfactuals and Causal Inference"
   - Real-world causal studies

4. **Fallacies**
   - Fallacy Files (fallacyfiles.org)
   - Informal Logic journal examples
   - Debate transcripts

5. **Arguments**
   - Argument Mining corpora (Stab & Gurevych)
   - Political debate transcripts
   - Philosophy papers

### 9.2 Test Data Format

```json
{
  "id": "test_LR_001",
  "category": "logical_reasoning",
  "subcategory": "modus_ponens",
  "difficulty": "easy",
  "input": {
    "premises": ["If P then Q", "P"],
    "conclusion": "Q"
  },
  "expected_output": {
    "is_valid": true,
    "is_provable": true,
    "steps": ["modus_ponens"],
    "confidence": 0.95
  },
  "metadata": {
    "source": "Copi & Cohen, p. 245",
    "expert_agreement": 1.0,
    "notes": "Classic modus ponens example"
  }
}
```

---

## 10. Appendix

### 10.1 Tool Reference

Complete tool listing with parameters and outputs (31 tools):

1. **think** - Multi-mode reasoning
2. **history** - Thought history
3. **search** - Full-text search
4. **get-metrics** - System metrics
5. **list-branches** - Branch listing
6. **focus-branch** - Switch branch
7. **branch-history** - Branch details
8. **recent-branches** - Recent branches
9. **validate** - Logical validation
10. **prove** - Formal proof
11. **check-syntax** - Syntax check
12. **probabilistic-reasoning** - Bayesian inference
13. **assess-evidence** - Evidence assessment
14. **detect-contradictions** - Find contradictions
15. **make-decision** - Multi-criteria decision
16. **decompose-problem** - Problem decomposition
17. **sensitivity-analysis** - Robustness testing
18. **self-evaluate** - Self-assessment
19. **detect-biases** - Bias detection
20. **analyze-perspectives** - Stakeholder analysis
21. **analyze-temporal** - Temporal analysis
22. **compare-time-horizons** - Time comparison
23. **identify-optimal-timing** - Timing optimization
24. **build-causal-graph** - Causal graph construction
25. **simulate-intervention** - Intervention simulation
26. **generate-counterfactual** - Counterfactual generation
27. **analyze-correlation-vs-causation** - Causation analysis
28. **get-causal-graph** - Retrieve causal graph
29. **synthesize-insights** - Cross-mode synthesis
30. **detect-emergent-patterns** - Pattern detection

### 10.2 Testing Tools & Frameworks

**Go Testing Framework:**
- `testing` - Standard library
- `testify/assert` - Assertions
- `testify/mock` - Mocking
- `go-cmp` - Deep comparison

**Coverage Analysis:**
- `go test -cover` - Coverage measurement
- `go tool cover` - Coverage visualization

**Performance Testing:**
- `testing.B` - Benchmarks
- `pprof` - Profiling

**Test Data:**
- JSON test fixtures
- Table-driven tests
- Property-based testing (gopter)

### 10.3 References

1. Pearl, J. (2009). *Causality: Models, Reasoning, and Inference* (2nd ed.). Cambridge University Press.

2. Kahneman, D., Slovic, P., & Tversky, A. (1982). *Judgment under Uncertainty: Heuristics and Biases*. Cambridge University Press.

3. Walton, D. (2013). *Methods of Argumentation*. Cambridge University Press.

4. Copi, I. M., Cohen, C., & McMahon, K. (2014). *Introduction to Logic* (14th ed.). Routledge.

5. Gelman, A., et al. (2013). *Bayesian Data Analysis* (3rd ed.). CRC Press.

6. Stab, C., & Gurevych, I. (2017). "Parsing Argumentation Structures in Persuasive Essays." *Computational Linguistics*, 43(3), 619-659.

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-10-03 | Test Coverage Engineer | Initial comprehensive test plan |

---

**Document Status:** Draft
**Next Review:** 2025-10-10
**Approval Required:** Technical Lead, QA Lead
