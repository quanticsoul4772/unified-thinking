// Package validation - String constants for logical validation patterns
package validation

// Logical connective patterns used in validation
const (
	// Implication connectives
	ConnImplies   = " implies "
	ConnThen      = " then "
	ConnTherefore = " therefore "
	ConnThus      = " thus "
	ConnHence     = " hence "

	// Negation prefixes
	NegNot   = "not "
	NegNo    = "no "
	NegNever = "never "
	NegNone  = "none "

	// Universal quantifiers
	QuantAll   = "all "
	QuantEvery = "every "
	QuantEach  = "each "

	// Existential quantifiers (reserved for future use)
	QuantSome     = "some "
	QuantExists   = "exists "
	QuantThereIs  = "there is "
	QuantThereAre = "there are "
)

// Logical operator patterns
var (
	// ImplicationPatterns are patterns indicating logical implications
	ImplicationPatterns = []string{
		ConnImplies,
		ConnThen,
		ConnTherefore,
		ConnThus,
		ConnHence,
	}

	// NegationPatterns are patterns indicating logical negation
	NegationPatterns = []string{
		NegNot,
		NegNo,
		NegNever,
		NegNone,
	}

	// UniversalPatterns are patterns indicating universal quantification
	UniversalPatterns = []string{
		QuantAll,
		QuantEvery,
		QuantEach,
	}

	// ExistentialPatterns are patterns indicating existential quantification
	ExistentialPatterns = []string{
		QuantSome,
		QuantExists,
		QuantThereIs,
		QuantThereAre,
	}
)

// Disjunction patterns
const (
	DisjOr     = " or "
	DisjEither = "either "
)

// Conjunction patterns
const (
	ConjAnd  = " and "
	ConjBoth = "both "
)

// Biconditional patterns
const (
	BiconIff          = " iff "
	BiconIfAndOnlyIf  = " if and only if "
	BiconEquivalent   = " equivalent to "
)

// Conditional patterns
const (
	CondIf     = "if "
	CondWhen   = "when "
	CondGiven  = "given "
	CondAssume = "assume "
)

// Inference patterns
const (
	InfConclusion = "conclusion"
	InfPremise    = "premise"
	InfInference  = "inference"
)

// Proof step types
const (
	ProofModusPonens         = "modus ponens"
	ProofModusTollens        = "modus tollens"
	ProofHypotheticalSyll    = "hypothetical syllogism"
	ProofDisjunctiveSyll     = "disjunctive syllogism"
	ProofUniversalInstant    = "universal instantiation"
	ProofExistentialGeneral  = "existential generalization"
	ProofReductioAdAbsurdum  = "reductio ad absurdum"
	ProofContradiction       = "contradiction"
)

// Validation result messages
const (
	MsgLogicallyConsistent   = "Thought is logically consistent"
	MsgContradictionDetected = "Contradiction detected"
	MsgFallacyDetected       = "Logical fallacy detected"
	MsgInvalidInference      = "Invalid inference"
	MsgValidProof            = "Valid proof"
	MsgInvalidProof          = "Invalid proof"
	MsgInconclusiveProof     = "Proof inconclusive"
)

// Fallacy type names
const (
	FallacyCircularReasoning     = "circular reasoning"
	FallacyAffirmingConsequent   = "affirming the consequent"
	FallacyDenyingAntecedent     = "denying the antecedent"
	FallacyFalseEquivalence      = "false equivalence"
	FallacyHastyGeneralization   = "hasty generalization"
	FallacyAppealToAuthority     = "appeal to authority"
	FallacyAppealToEmotion       = "appeal to emotion"
	FallacyStrawMan              = "straw man"
	FallacyAdHominem             = "ad hominem"
	FallacyFalseDichotomy        = "false dichotomy"
	FallacySlipperySlope         = "slippery slope"
	FallacyRedHerring            = "red herring"
	FallacyBeggingTheQuestion    = "begging the question"
	FallacyEquivocation          = "equivocation"
	FallacyAppealToIgnorance     = "appeal to ignorance"
	FallacyComposition           = "composition fallacy"
	FallacyDivision              = "division fallacy"
	FallacyNoTrueScotsman        = "no true scotsman"
	FallacyGamblers              = "gambler's fallacy"
	FallacyHotHand               = "hot hand fallacy"
)

// Evidence quality levels
const (
	EvidenceStrong   = "strong"
	EvidenceModerate = "moderate"
	EvidenceWeak     = "weak"
	EvidenceAbsent   = "absent"
)

// Confidence thresholds
const (
	ConfidenceHighThreshold   = 0.8
	ConfidenceMediumThreshold = 0.5
	ConfidenceLowThreshold    = 0.3
)
