// Package analysis provides analytical reasoning capabilities.
package analysis

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"unified-thinking/internal/types"
)

// PerspectiveAnalyzer generates and analyzes multiple stakeholder perspectives
type PerspectiveAnalyzer struct {
	mu      sync.RWMutex
	counter int
}

// NewPerspectiveAnalyzer creates a new perspective analyzer
func NewPerspectiveAnalyzer() *PerspectiveAnalyzer {
	return &PerspectiveAnalyzer{}
}

// AnalyzePerspectives generates multiple stakeholder perspectives for a situation
func (pa *PerspectiveAnalyzer) AnalyzePerspectives(situation string, stakeholderHints []string) ([]*types.Perspective, error) {
	if situation == "" {
		return nil, fmt.Errorf("situation cannot be empty")
	}

	pa.mu.Lock()
	defer pa.mu.Unlock()

	perspectives := make([]*types.Perspective, 0)

	// If stakeholder hints provided, analyze from those perspectives
	if len(stakeholderHints) > 0 {
		for _, stakeholder := range stakeholderHints {
			perspective := pa.generatePerspective(situation, stakeholder)
			perspectives = append(perspectives, perspective)
		}
	} else {
		// Auto-detect relevant stakeholders from situation
		detectedStakeholders := pa.detectStakeholders(situation)
		for _, stakeholder := range detectedStakeholders {
			perspective := pa.generatePerspective(situation, stakeholder)
			perspectives = append(perspectives, perspective)
		}
	}

	// Detect conflicts between perspectives
	conflicts := pa.detectPerspectiveConflicts(perspectives)
	if len(conflicts) > 0 {
		// Add conflict metadata to perspectives
		for _, p := range perspectives {
			p.Metadata["conflicts"] = conflicts
		}
	}

	return perspectives, nil
}

// generatePerspective creates a perspective for a specific stakeholder
func (pa *PerspectiveAnalyzer) generatePerspective(situation, stakeholder string) *types.Perspective {
	pa.counter++

	// Extract key concerns based on stakeholder type
	concerns := pa.extractConcerns(situation, stakeholder)
	priorities := pa.extractPriorities(stakeholder)
	constraints := pa.extractConstraints(stakeholder)
	viewpoint := pa.synthesizeViewpoint(situation, stakeholder, concerns)

	// Confidence based on how well-defined the stakeholder is
	confidence := pa.estimateConfidence(stakeholder, situation)

	return &types.Perspective{
		ID:          fmt.Sprintf("perspective-%d", pa.counter),
		Stakeholder: stakeholder,
		Viewpoint:   viewpoint,
		Concerns:    concerns,
		Priorities:  priorities,
		Constraints: constraints,
		Confidence:  confidence,
		Metadata:    map[string]interface{}{},
		CreatedAt:   time.Now(),
	}
}

// detectStakeholders identifies relevant stakeholders from the situation description
func (pa *PerspectiveAnalyzer) detectStakeholders(situation string) []string {
	situationLower := strings.ToLower(situation)
	stakeholders := make([]string, 0)

	// Check for common stakeholder indicators
	stakeholderPatterns := map[string][]string{
		"users":      {"user", "customer", "client", "consumer"},
		"employees":  {"employee", "worker", "staff", "team"},
		"management": {"manager", "executive", "leadership", "ceo", "director"},
		"investors":  {"investor", "shareholder", "stakeholder", "board"},
		"community":  {"community", "public", "society", "citizen"},
		"regulators": {"regulator", "government", "compliance", "legal"},
		"partners":   {"partner", "supplier", "vendor", "contractor"},
	}

	for stakeholder, patterns := range stakeholderPatterns {
		for _, pattern := range patterns {
			if strings.Contains(situationLower, pattern) {
				stakeholders = append(stakeholders, stakeholder)
				break
			}
		}
	}

	// If no stakeholders detected, use generic set
	if len(stakeholders) == 0 {
		stakeholders = []string{"decision-maker", "affected-parties", "implementers"}
	}

	return stakeholders
}

// extractConcerns identifies key concerns for a stakeholder type
func (pa *PerspectiveAnalyzer) extractConcerns(situation, stakeholder string) []string {
	concerns := make([]string, 0)
	situationLower := strings.ToLower(situation)
	stakeholderLower := strings.ToLower(stakeholder)

	// Stakeholder-specific concern patterns
	concernPatterns := map[string][]string{
		// Business stakeholders
		"user":       {"usability", "accessibility", "privacy", "cost", "reliability"},
		"customer":   {"value", "quality", "support", "price", "experience"},
		"employee":   {"workload", "job security", "compensation", "work environment", "career growth"},
		"management": {"profitability", "efficiency", "risk", "scalability", "market position"},
		"investor":   {"return on investment", "risk", "growth potential", "market share", "valuation"},
		"community":  {"social impact", "environmental impact", "fairness", "accessibility", "safety"},
		"regulator":  {"compliance", "safety", "fairness", "transparency", "accountability"},
		"partner":    {"reliability", "communication", "mutual benefit", "contract terms", "long-term viability"},

		// Professional/Academic stakeholders
		"cognitive":     {"mental models", "cognitive load", "learning efficacy", "information processing", "decision biases"},
		"scientist":     {"empirical validity", "methodological rigor", "replicability", "theoretical coherence", "measurement precision"},
		"psychologist":  {"psychological impact", "behavioral patterns", "emotional well-being", "developmental stages", "therapeutic efficacy"},
		"therapist":     {"emotional safety", "trauma sensitivity", "therapeutic alliance", "healing process", "boundaries"},
		"mathematician": {"logical consistency", "formal proof", "computational complexity", "optimization", "precision"},
		"engineer":      {"technical feasibility", "system reliability", "performance metrics", "maintenance burden", "scalability"},
		"researcher":    {"research validity", "data integrity", "peer review", "reproducibility", "ethical considerations"},

		// Creative/Artistic stakeholders
		"artist":   {"creative expression", "aesthetic integrity", "emotional resonance", "artistic freedom", "cultural impact"},
		"creative": {"originality", "innovation", "inspiration", "artistic vision", "experiential quality"},
		"designer": {"user experience", "visual harmony", "functional beauty", "accessibility", "iterative refinement"},
		"writer":   {"narrative coherence", "authentic voice", "reader engagement", "clarity of message", "emotional truth"},

		// Spiritual/Philosophical stakeholders
		"spiritual":   {"meaning and purpose", "inner peace", "compassion", "interconnectedness", "transcendence"},
		"philosopher": {"ethical implications", "logical validity", "conceptual clarity", "existential meaning", "moral framework"},
		"ethicist":    {"moral permissibility", "harm prevention", "fairness", "autonomy", "justice"},
		"religious":   {"spiritual growth", "faith alignment", "community values", "sacred respect", "moral guidance"},

		// Healthcare/Wellness stakeholders
		"doctor":    {"patient safety", "clinical efficacy", "evidence-based practice", "side effects", "treatment outcomes"},
		"nurse":     {"patient comfort", "holistic care", "safety protocols", "workload management", "compassionate delivery"},
		"counselor": {"client welfare", "confidentiality", "professional boundaries", "therapeutic effectiveness", "ethical practice"},
		"grief":     {"emotional processing", "loss acknowledgment", "healing timeline", "support systems", "meaning-making"},

		// Technical/IT stakeholders
		"developer": {"code quality", "maintainability", "technical debt", "testing coverage", "developer experience"},
		"security":  {"vulnerability exposure", "attack surface", "data protection", "compliance", "access control"},
		"devops":    {"deployment reliability", "system uptime", "monitoring coverage", "infrastructure costs", "automation"},
		"architect": {"system scalability", "integration complexity", "architectural patterns", "technical roadmap", "modularity"},
		"qa":        {"test coverage", "defect risk", "regression potential", "edge cases", "quality metrics"},
		"product":   {"user value", "market fit", "feature prioritization", "success metrics", "competitive advantage"},
		"ops":       {"operational burden", "incident response", "system reliability", "runbook coverage", "on-call impact"},
		"data":      {"data quality", "privacy compliance", "data governance", "analytical value", "data architecture"},

		// Finance/Business stakeholders
		"finance":     {"cost control", "ROI", "budget impact", "cash flow", "financial risk"},
		"cfo":         {"capital allocation", "financial sustainability", "shareholder value", "risk exposure", "growth investment"},
		"procurement": {"vendor risk", "contract terms", "total cost", "supplier relationships", "competitive pricing"},
		"legal":       {"legal liability", "regulatory compliance", "contractual risk", "intellectual property", "litigation risk"},

		// Leadership/Strategy stakeholders
		"executive": {"strategic alignment", "competitive position", "organizational impact", "stakeholder value", "opportunity cost"},
		"cto":       {"technical strategy", "innovation capability", "technology roadmap", "engineering excellence", "technical leadership"},
		"founder":   {"product-market fit", "company mission", "growth trajectory", "company culture", "vision alignment"},

		// Team/HR stakeholders
		"hr":        {"employee impact", "organizational culture", "training needs", "talent retention", "workforce planning"},
		"team lead": {"team workload", "skill development", "team morale", "delivery commitments", "resource allocation"},

		// Default fallback
		"default": {"impact", "feasibility", "risks", "benefits", "implementation"},
	}

	// Find matching patterns
	var relevantPatterns []string
	for key, patterns := range concernPatterns {
		if strings.Contains(stakeholderLower, key) {
			relevantPatterns = patterns
			break
		}
	}
	if len(relevantPatterns) == 0 {
		relevantPatterns = concernPatterns["default"]
	}

	// Extract concerns that appear relevant to the situation
	for _, concern := range relevantPatterns {
		if strings.Contains(situationLower, concern) || len(concerns) < 3 {
			concerns = append(concerns, concern)
			if len(concerns) >= 5 {
				break
			}
		}
	}

	return concerns
}

// extractPriorities determines priorities for a stakeholder type
func (pa *PerspectiveAnalyzer) extractPriorities(stakeholder string) []string {
	stakeholderLower := strings.ToLower(stakeholder)

	priorityMap := map[string][]string{
		// Business stakeholders
		"user":       {"ease of use", "reliability", "value for money"},
		"customer":   {"quality", "price", "customer service"},
		"employee":   {"fair compensation", "work-life balance", "job security"},
		"management": {"profitability", "growth", "operational efficiency"},
		"investor":   {"returns", "risk mitigation", "long-term growth"},
		"community":  {"social benefit", "environmental sustainability", "equity"},
		"regulator":  {"public safety", "compliance", "consumer protection"},
		"partner":    {"mutual success", "clear communication", "reliable execution"},

		// Professional/Academic stakeholders
		"cognitive":     {"understanding cognitive processes", "reducing mental strain", "enhancing learning"},
		"scientist":     {"advancing knowledge", "methodological excellence", "empirical truth"},
		"psychologist":  {"promoting mental health", "understanding behavior", "evidence-based intervention"},
		"therapist":     {"client healing", "safe therapeutic space", "ethical practice"},
		"mathematician": {"mathematical rigor", "elegant solutions", "logical clarity"},
		"engineer":      {"robust solutions", "efficient systems", "practical implementation"},
		"researcher":    {"scientific integrity", "reproducible results", "advancing understanding"},

		// Creative/Artistic stakeholders
		"artist":   {"authentic expression", "emotional impact", "creative freedom"},
		"creative": {"innovative thinking", "original solutions", "experiential quality"},
		"designer": {"beautiful functionality", "intuitive experience", "accessibility"},
		"writer":   {"clear communication", "engaging narrative", "authentic voice"},

		// Spiritual/Philosophical stakeholders
		"spiritual":   {"inner growth", "compassionate action", "meaningful existence"},
		"philosopher": {"conceptual clarity", "ethical soundness", "truth-seeking"},
		"ethicist":    {"moral integrity", "harm prevention", "justice"},
		"religious":   {"spiritual alignment", "community welfare", "sacred respect"},

		// Healthcare/Wellness stakeholders
		"doctor":    {"patient outcomes", "evidence-based care", "do no harm"},
		"nurse":     {"patient advocacy", "holistic wellness", "compassionate care"},
		"counselor": {"client empowerment", "professional ethics", "therapeutic effectiveness"},
		"grief":     {"honoring loss", "supporting healing", "facilitating meaning"},

		// Technical/IT stakeholders
		"developer": {"clean code", "efficient delivery", "maintainable systems"},
		"security":  {"defense in depth", "data protection", "compliance"},
		"devops":    {"deployment reliability", "system observability", "automation"},
		"architect": {"scalable design", "technical coherence", "long-term maintainability"},
		"qa":        {"quality assurance", "comprehensive testing", "defect prevention"},
		"product":   {"user satisfaction", "business value", "market success"},
		"ops":       {"system reliability", "operational efficiency", "incident prevention"},
		"data":      {"data integrity", "analytical insight", "privacy compliance"},

		// Finance/Business stakeholders
		"finance":     {"financial sustainability", "cost efficiency", "accurate forecasting"},
		"cfo":         {"value creation", "capital efficiency", "risk management"},
		"procurement": {"cost optimization", "vendor quality", "supply chain resilience"},
		"legal":       {"risk mitigation", "regulatory compliance", "organizational protection"},

		// Leadership/Strategy stakeholders
		"executive": {"strategic success", "stakeholder value", "organizational excellence"},
		"cto":       {"technical leadership", "innovation enablement", "engineering culture"},
		"founder":   {"vision realization", "sustainable growth", "mission impact"},

		// Team/HR stakeholders
		"hr":        {"employee wellbeing", "organizational health", "talent development"},
		"team lead": {"team success", "member growth", "delivery excellence"},
	}

	for key, priorities := range priorityMap {
		if strings.Contains(stakeholderLower, key) {
			return priorities
		}
	}

	return []string{"positive outcomes", "minimal risk", "clear benefits"}
}

// extractConstraints identifies constraints for a stakeholder type
func (pa *PerspectiveAnalyzer) extractConstraints(stakeholder string) []string {
	stakeholderLower := strings.ToLower(stakeholder)

	constraintMap := map[string][]string{
		// Business stakeholders
		"user":       {"limited budget", "limited technical expertise", "time constraints"},
		"customer":   {"budget limitations", "alternative options available", "switching costs"},
		"employee":   {"limited authority", "resource constraints", "existing workload"},
		"management": {"budget constraints", "timeline pressure", "stakeholder expectations"},
		"investor":   {"fiduciary duty", "portfolio diversification", "liquidity needs"},
		"community":  {"limited resources", "diverse needs", "existing infrastructure"},
		"regulator":  {"legal framework", "enforcement capacity", "political pressures"},
		"partner":    {"contractual obligations", "resource limitations", "competing priorities"},

		// Professional/Academic stakeholders
		"cognitive":     {"measurement limitations", "individual variability", "confounding factors"},
		"scientist":     {"funding constraints", "publication pressure", "peer review timeline"},
		"psychologist":  {"ethical boundaries", "individual differences", "treatment limitations"},
		"therapist":     {"confidentiality requirements", "session time limits", "scope of practice"},
		"mathematician": {"computational limits", "axiomatic constraints", "proof complexity"},
		"engineer":      {"physical laws", "material constraints", "budget limitations"},
		"researcher":    {"sample size limits", "ethical restrictions", "timeframe constraints"},

		// Creative/Artistic stakeholders
		"artist":   {"medium limitations", "financial constraints", "market pressures"},
		"creative": {"resource availability", "time constraints", "commercial viability"},
		"designer": {"technical limitations", "user constraints", "budget restrictions"},
		"writer":   {"publisher requirements", "audience expectations", "deadline pressure"},

		// Spiritual/Philosophical stakeholders
		"spiritual":   {"individual readiness", "cultural context", "practice traditions"},
		"philosopher": {"logical consistency", "conceptual clarity", "argument structure"},
		"ethicist":    {"moral pluralism", "cultural relativism", "practical feasibility"},
		"religious":   {"doctrinal boundaries", "community norms", "sacred texts"},

		// Healthcare/Wellness stakeholders
		"doctor":    {"medical ethics", "resource availability", "regulatory compliance"},
		"nurse":     {"staffing ratios", "time constraints", "institutional policies"},
		"counselor": {"professional boundaries", "confidentiality", "competency limits"},
		"grief":     {"individual pace", "cultural variations", "non-linear process"},

		// Technical/IT stakeholders
		"developer": {"technical constraints", "legacy code", "timeline pressure"},
		"security":  {"threat landscape", "compliance requirements", "attack surface"},
		"devops":    {"infrastructure limits", "deployment complexity", "monitoring gaps"},
		"architect": {"existing systems", "technology constraints", "migration complexity"},
		"qa":        {"test environment limits", "time constraints", "coverage gaps"},
		"product":   {"resource constraints", "market timing", "competitive pressure"},
		"ops":       {"staffing limits", "tooling constraints", "on-call burden"},
		"data":      {"data quality issues", "privacy regulations", "infrastructure limits"},

		// Finance/Business stakeholders
		"finance":     {"budget limits", "forecasting uncertainty", "financial regulations"},
		"cfo":         {"capital constraints", "market conditions", "shareholder expectations"},
		"procurement": {"contract terms", "vendor limitations", "budget constraints"},
		"legal":       {"regulatory framework", "legal precedent", "risk tolerance"},

		// Leadership/Strategy stakeholders
		"executive": {"board expectations", "market dynamics", "organizational inertia"},
		"cto":       {"technical debt", "talent constraints", "technology evolution"},
		"founder":   {"runway constraints", "investor expectations", "market timing"},

		// Team/HR stakeholders
		"hr":        {"policy constraints", "legal requirements", "budget limitations"},
		"team lead": {"resource constraints", "skill gaps", "competing priorities"},
	}

	for key, constraints := range constraintMap {
		if strings.Contains(stakeholderLower, key) {
			return constraints
		}
	}

	return []string{"practical limitations", "resource constraints", "external dependencies"}
}

// viewpointTemplate defines a stakeholder-specific viewpoint generator
type viewpointTemplate struct {
	prefix    string // Opening statement establishing the stakeholder's lens
	questions string // Key questions this stakeholder would ask
	suffix    string // Closing framing or recommendation style
}

// stakeholderViewpoints maps stakeholder types to their unique viewpoint templates
var stakeholderViewpoints = map[string]viewpointTemplate{
	// Business stakeholders
	"user": {
		prefix:    "As a user, I evaluate this based on how it affects my daily experience and outcomes.",
		questions: "Will this be easy to use? Is it reliable? Does it respect my time and privacy?",
		suffix:    "The user community needs clear communication about changes and genuine responsiveness to feedback.",
	},
	"customer": {
		prefix:    "As a customer making a purchase decision, I weigh value against cost and alternatives.",
		questions: "What am I getting for my money? How does this compare to alternatives? What's the support like?",
		suffix:    "Customers need transparency about what they're buying and confidence in long-term value.",
	},
	"employee": {
		prefix:    "As an employee, I consider how this affects my work life, growth, and job security.",
		questions: "How will this change my daily work? Does this align with my career goals? Is my position secure?",
		suffix:    "The workforce needs clarity about expectations and genuine investment in their development.",
	},
	"management": {
		prefix:    "From a management perspective, I must balance operational efficiency with strategic goals.",
		questions: "What's the ROI? How does this affect our competitive position? What are the implementation risks?",
		suffix:    "Leadership must make decisions that serve both short-term performance and long-term sustainability.",
	},
	"investor": {
		prefix:    "As an investor, I analyze risk-adjusted returns and long-term value creation potential.",
		questions: "What's the expected return? What are the downside risks? How does this affect company valuation?",
		suffix:    "Investment decisions require clear metrics and transparent reporting of both opportunities and risks.",
	},
	"community": {
		prefix:    "From a community perspective, I consider the broader social and environmental impact.",
		questions: "How does this affect our neighborhood? Is it equitable? What are the environmental implications?",
		suffix:    "Communities deserve a voice in decisions that affect their quality of life and shared resources.",
	},
	"regulator": {
		prefix:    "As a regulator, I must ensure compliance with laws and protection of public interests.",
		questions: "Does this meet legal requirements? Are consumers protected? What precedents does this set?",
		suffix:    "Regulatory frameworks must balance innovation with accountability and public safety.",
	},
	"partner": {
		prefix:    "As a business partner, I evaluate this through the lens of our mutual success and relationship.",
		questions: "How does this affect our partnership? Are the terms fair? Can we trust long-term commitment?",
		suffix:    "Partnerships thrive on clear expectations, shared goals, and equitable value distribution.",
	},

	// Professional/Academic stakeholders
	"scientist": {
		prefix:    "As a scientist, I evaluate this through the lens of empirical evidence and methodological rigor.",
		questions: "What data supports this claim? Is the methodology sound? Can these results be replicated?",
		suffix:    "The scientific community requires peer review, transparent methodology, and acknowledgment of uncertainty.",
	},
	"researcher": {
		prefix:    "As a researcher, I examine the theoretical foundations and evidentiary basis.",
		questions: "What prior work informs this? Where are the knowledge gaps? What hypotheses can we test?",
		suffix:    "Research integrity demands acknowledgment of limitations and commitment to following evidence.",
	},
	"psychologist": {
		prefix:    "As a psychologist, I consider the cognitive, emotional, and behavioral dimensions.",
		questions: "How does this affect mental well-being? What psychological factors are at play? What interventions might help?",
		suffix:    "Psychological understanding requires attention to individual differences and contextual factors.",
	},
	"engineer": {
		prefix:    "As an engineer, I focus on technical feasibility, reliability, and practical implementation.",
		questions: "Is this technically achievable? What are the failure modes? How do we ensure quality?",
		suffix:    "Engineering decisions must balance innovation with reliability and maintainability.",
	},
	"mathematician": {
		prefix:    "As a mathematician, I analyze the logical structure and formal properties.",
		questions: "Is this logically consistent? What can be proven? What are the boundary conditions?",
		suffix:    "Mathematical rigor demands precision in definitions and completeness in proofs.",
	},

	// Policy/Economic stakeholders
	"policymaker": {
		prefix:    "From a policy perspective, I must balance multiple stakeholder interests and resource constraints.",
		questions: "What is the societal benefit? How do we measure success? What are the opportunity costs?",
		suffix:    "Policy decisions require balancing competing interests while maintaining democratic accountability.",
	},
	"taxpayer": {
		prefix:    "As a taxpayer funding this through public money, I want accountability for how resources are used.",
		questions: "Is this the best use of limited public funds? What practical benefits will citizens see? Are there more pressing priorities?",
		suffix:    "Public spending decisions must demonstrate clear value and transparent accounting to those who fund them.",
	},
	"economist": {
		prefix:    "As an economist, I analyze incentives, trade-offs, and market dynamics.",
		questions: "What are the economic incentives? Who bears the costs and who receives the benefits? What market failures might occur?",
		suffix:    "Economic analysis requires understanding both intended effects and unintended consequences.",
	},

	// Creative/Artistic stakeholders
	"artist": {
		prefix:    "As an artist, I consider aesthetic value, creative expression, and cultural impact.",
		questions: "Does this enable authentic expression? What emotional resonance does it create? How does it contribute to culture?",
		suffix:    "Artistic endeavors require freedom of expression and recognition of subjective experience.",
	},
	"designer": {
		prefix:    "As a designer, I evaluate usability, aesthetics, and the human experience.",
		questions: "Is this intuitive? Does form follow function? How does it feel to use?",
		suffix:    "Design excellence balances beauty with utility and serves human needs.",
	},

	// Healthcare/Wellness stakeholders
	"doctor": {
		prefix:    "As a physician, I prioritize patient safety and evidence-based practice.",
		questions: "What does the clinical evidence show? What are the risks and benefits? Does this follow best practices?",
		suffix:    "Medical decisions must be grounded in evidence while respecting patient autonomy.",
	},
	"patient": {
		prefix:    "As a patient, I'm concerned with my health outcomes and treatment experience.",
		questions: "Will this help me get better? What are the side effects? How will this affect my daily life?",
		suffix:    "Patients need clear information to make informed decisions about their own care.",
	},

	// Philosophical/Ethical stakeholders
	"philosopher": {
		prefix:    "As a philosopher, I examine the conceptual foundations and ethical implications.",
		questions: "What assumptions underlie this? Is it logically coherent? What ethical principles apply?",
		suffix:    "Philosophical inquiry requires questioning assumptions and pursuing conceptual clarity.",
	},
	"ethicist": {
		prefix:    "As an ethicist, I evaluate the moral dimensions and implications for human welfare.",
		questions: "Is this morally permissible? Who might be harmed? Are all stakeholders treated fairly?",
		suffix:    "Ethical analysis demands consideration of all affected parties and long-term consequences.",
	},

	// Scientific domain stakeholders
	"physicist": {
		prefix:    "As a physicist, I evaluate this through the lens of fundamental physical laws and experimental verification.",
		questions: "What physical mechanisms are at play? Is this consistent with established physics? What experiments could test this?",
		suffix:    "Physics demands rigorous mathematical formulation and experimental validation before acceptance.",
	},
	"experimentalist": {
		prefix:    "As an experimentalist, I focus on measurable outcomes and reproducible results.",
		questions: "How can we test this empirically? What are the sources of systematic error? Can others replicate these findings?",
		suffix:    "Experimental science requires careful controls, quantified uncertainties, and independent replication.",
	},
	"theorist": {
		prefix:    "As a theorist, I analyze the mathematical and conceptual framework underlying this problem.",
		questions: "Is the theory internally consistent? Does it make falsifiable predictions? How does it connect to established frameworks?",
		suffix:    "Theoretical work must balance mathematical elegance with empirical grounding and predictive power.",
	},

	// Policy/Governance stakeholders (additional)
	"politician": {
		prefix:    "As a political representative, I consider constituent interests and electoral implications.",
		questions: "How will this affect my constituents? Is there public support? What are the political trade-offs?",
		suffix:    "Political decisions must balance expert advice with democratic representation and public sentiment.",
	},
	"bureaucrat": {
		prefix:    "As a public administrator, I focus on implementation feasibility and compliance.",
		questions: "Is this administratively feasible? What regulations apply? How do we ensure accountability?",
		suffix:    "Public administration requires balancing efficiency with transparency and due process.",
	},

	// Public stakeholders (additional)
	"citizen": {
		prefix:    "As a citizen, I consider the broader impact on society and my community.",
		questions: "How does this serve the public good? Is it fair to all members of society? What are the long-term implications?",
		suffix:    "Citizens deserve transparent information and meaningful input into decisions affecting their lives.",
	},
	"activist": {
		prefix:    "As an activist, I advocate for underrepresented interests and challenge the status quo.",
		questions: "Whose voices are being ignored? What injustices does this perpetuate? How can we push for better outcomes?",
		suffix:    "Activism demands holding power accountable while building coalitions for meaningful change.",
	},

	// Technical/IT stakeholders
	"developer": {
		prefix:    "As a developer, I evaluate this from a code quality, maintainability, and implementation perspective.",
		questions: "Is this technically sound? How will it affect our codebase? What are the testing implications? Will this introduce technical debt?",
		suffix:    "Development decisions should balance speed of delivery with code quality, maintainability, and developer experience.",
	},
	"security": {
		prefix:    "As a security professional, I assess threats, vulnerabilities, and risk mitigation strategies.",
		questions: "What attack vectors does this introduce? Are there authentication/authorization concerns? How do we protect sensitive data? What compliance requirements apply?",
		suffix:    "Security decisions must balance protection with usability while adhering to defense-in-depth principles.",
	},
	"devops": {
		prefix:    "As a DevOps engineer, I focus on deployment, reliability, and operational excellence.",
		questions: "How will this affect our deployment pipeline? What are the monitoring and alerting needs? How do we ensure high availability?",
		suffix:    "DevOps practices should enable rapid, reliable deployments while maintaining system stability.",
	},
	"architect": {
		prefix:    "As a software architect, I consider system design, scalability, and long-term technical strategy.",
		questions: "Does this fit our architectural patterns? How will it scale? What are the integration points? Will this create technical debt?",
		suffix:    "Architectural decisions shape the system's future; they must balance current needs with long-term evolution.",
	},
	"qa": {
		prefix:    "As a QA engineer, I evaluate testability, quality assurance, and risk of defects.",
		questions: "How do we test this thoroughly? What edge cases might break? Where are the highest-risk areas? Is there adequate test coverage?",
		suffix:    "Quality assurance requires systematic testing while balancing thoroughness with time-to-market pressures.",
	},
	"product": {
		prefix:    "As a product manager, I balance user needs, business goals, and technical feasibility.",
		questions: "Does this solve a real user problem? How does it align with our product strategy? What's the expected impact on key metrics?",
		suffix:    "Product decisions should be data-informed, user-centric, and aligned with business objectives.",
	},
	"ops": {
		prefix:    "As an operations team member, I focus on system reliability, incident response, and operational burden.",
		questions: "How will we support this in production? What could go wrong at 3am? Do we have adequate monitoring and runbooks?",
		suffix:    "Operations sustainability requires designing for observability, resilience, and manageable on-call burden.",
	},
	"data": {
		prefix:    "As a data professional, I consider data quality, governance, privacy, and analytical value.",
		questions: "How does this affect our data architecture? Are we maintaining data quality? What privacy implications exist? How does this enable analytics?",
		suffix:    "Data decisions must balance accessibility with governance, and operational needs with analytical value.",
	},

	// Finance/Business stakeholders
	"finance": {
		prefix:    "As a finance professional, I analyze costs, ROI, and financial sustainability.",
		questions: "What is the total cost of ownership? What's the expected ROI? How does this affect our budget and cash flow? Are there hidden costs?",
		suffix:    "Financial decisions require rigorous analysis of both immediate costs and long-term financial implications.",
	},
	"cfo": {
		prefix:    "As CFO, I evaluate financial risk, capital allocation, and value creation.",
		questions: "How does this affect our financial position? What's the payback period? Are we optimizing capital allocation? What financial risks exist?",
		suffix:    "CFO decisions balance growth investment with financial prudence and shareholder value creation.",
	},
	"procurement": {
		prefix:    "As a procurement professional, I assess vendor relationships, contracts, and total cost.",
		questions: "Are we getting competitive pricing? What are the contract terms and risks? How does this fit our vendor strategy?",
		suffix:    "Procurement decisions should optimize total cost of ownership while managing supplier relationships and risks.",
	},
	"legal": {
		prefix:    "As legal counsel, I identify legal risks, contractual issues, and compliance requirements.",
		questions: "What legal risks does this create? Are we compliant with relevant regulations? What contractual protections do we need?",
		suffix:    "Legal guidance must protect the organization while enabling business objectives within regulatory frameworks.",
	},

	// Leadership/Strategy stakeholders
	"executive": {
		prefix:    "As an executive, I consider strategic alignment, organizational impact, and stakeholder value.",
		questions: "Does this align with our strategic vision? What's the opportunity cost? How does this affect our competitive position?",
		suffix:    "Executive decisions must balance short-term execution with long-term strategic positioning.",
	},
	"cto": {
		prefix:    "As CTO, I evaluate technical strategy, innovation, and technology roadmap alignment.",
		questions: "Does this advance our technical capabilities? How does it fit our technology roadmap? What innovation opportunities does it create?",
		suffix:    "CTO decisions shape technological direction while ensuring technical excellence serves business goals.",
	},
	"founder": {
		prefix:    "As a founder, I consider product-market fit, growth potential, and company mission alignment.",
		questions: "Does this move us closer to our vision? Will this help us grow? Is this consistent with our values and mission?",
		suffix:    "Founder decisions must balance pragmatic business needs with the original vision and values.",
	},

	// Team/HR stakeholders
	"hr": {
		prefix:    "As an HR professional, I consider employee impact, culture, and organizational health.",
		questions: "How will this affect our employees? What are the training needs? Does this align with our culture and values?",
		suffix:    "HR decisions must balance organizational needs with employee well-being and development.",
	},
	"team lead": {
		prefix:    "As a team lead, I balance team productivity, morale, and delivery commitments.",
		questions: "How will this affect my team's workload? Do we have the skills needed? What support does the team need?",
		suffix:    "Team leadership requires advocating for team needs while meeting organizational commitments.",
	},
}

// synthesizeViewpoint creates a stakeholder-specific viewpoint that genuinely differs by perspective
func (pa *PerspectiveAnalyzer) synthesizeViewpoint(situation, stakeholder string, concerns []string) string {
	stakeholderLower := strings.ToLower(stakeholder)

	// Format concerns list
	concernsStr := "various factors"
	if len(concerns) > 0 {
		concernsStr = strings.Join(concerns, ", ")
	}

	// Find the best matching viewpoint template
	var template viewpointTemplate
	found := false
	for key, tmpl := range stakeholderViewpoints {
		if strings.Contains(stakeholderLower, key) {
			template = tmpl
			found = true
			break
		}
	}

	// If no specific template, generate a contextual generic one
	if !found {
		return pa.generateGenericViewpoint(situation, stakeholder, concerns)
	}

	// Build the stakeholder-specific viewpoint
	return fmt.Sprintf("%s My primary concerns center on %s. %s %s",
		template.prefix,
		concernsStr,
		template.questions,
		template.suffix)
}

// generateGenericViewpoint creates a varied viewpoint for unknown stakeholder types
func (pa *PerspectiveAnalyzer) generateGenericViewpoint(situation, stakeholder string, concerns []string) string {
	concernsStr := "the overall impact"
	if len(concerns) > 0 {
		concernsStr = strings.Join(concerns, ", ")
	}

	// Use the stakeholder name to create some variation even for unknown types
	return fmt.Sprintf("Speaking as a %s, I bring a distinct perspective shaped by my role and responsibilities. "+
		"My analysis focuses on %s. "+
		"The key questions I would raise are: How does this serve my constituents' interests? "+
		"What are the immediate and long-term implications? How can we ensure accountability? "+
		"Any path forward must address these concerns while remaining practical and achievable.",
		stakeholder, concernsStr)
}

// estimateConfidence estimates confidence in perspective modeling
func (pa *PerspectiveAnalyzer) estimateConfidence(stakeholder, situation string) float64 {
	// Higher confidence for well-defined stakeholders
	wellDefinedStakeholders := []string{"user", "customer", "employee", "investor", "manager"}
	confidence := 0.6 // Base confidence

	stakeholderLower := strings.ToLower(stakeholder)
	for _, wd := range wellDefinedStakeholders {
		if strings.Contains(stakeholderLower, wd) {
			confidence = 0.8
			break
		}
	}

	// Increase confidence if situation mentions stakeholder explicitly
	if strings.Contains(strings.ToLower(situation), stakeholderLower) {
		confidence += 0.1
	}

	// Clamp to valid range
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// detectPerspectiveConflicts identifies conflicts between perspectives
func (pa *PerspectiveAnalyzer) detectPerspectiveConflicts(perspectives []*types.Perspective) []string {
	conflicts := make([]string, 0)

	// Check for priority conflicts
	for i := 0; i < len(perspectives); i++ {
		for j := i + 1; j < len(perspectives); j++ {
			p1 := perspectives[i]
			p2 := perspectives[j]

			// Check if priorities are opposed
			if pa.prioritiesConflict(p1.Priorities, p2.Priorities) {
				conflict := fmt.Sprintf("%s and %s have conflicting priorities", p1.Stakeholder, p2.Stakeholder)
				conflicts = append(conflicts, conflict)
			}

			// Check if concerns are opposed
			if pa.concernsConflict(p1.Concerns, p2.Concerns) {
				conflict := fmt.Sprintf("%s and %s have opposing concerns", p1.Stakeholder, p2.Stakeholder)
				conflicts = append(conflicts, conflict)
			}
		}
	}

	return conflicts
}

// prioritiesConflict checks if two priority lists are in conflict
func (pa *PerspectiveAnalyzer) prioritiesConflict(priorities1, priorities2 []string) bool {
	// Simple heuristic: check for opposing terms
	opposingPairs := map[string]string{
		"speed":       "thoroughness",
		"cost":        "quality",
		"innovation":  "stability",
		"growth":      "sustainability",
		"flexibility": "standardization",
	}

	for _, p1 := range priorities1 {
		for _, p2 := range priorities2 {
			p1Lower := strings.ToLower(p1)
			p2Lower := strings.ToLower(p2)

			// Check for direct opposition
			for key, opposite := range opposingPairs {
				if (strings.Contains(p1Lower, key) && strings.Contains(p2Lower, opposite)) ||
					(strings.Contains(p1Lower, opposite) && strings.Contains(p2Lower, key)) {
					return true
				}
			}
		}
	}

	return false
}

// concernsConflict checks if two concern lists are in conflict
func (pa *PerspectiveAnalyzer) concernsConflict(concerns1, concerns2 []string) bool {
	// Check for mutually exclusive concerns
	mutuallyExclusive := map[string]string{
		"privacy":    "transparency",
		"security":   "accessibility",
		"control":    "autonomy",
		"efficiency": "thoroughness",
		"automation": "human oversight",
	}

	for _, c1 := range concerns1 {
		for _, c2 := range concerns2 {
			c1Lower := strings.ToLower(c1)
			c2Lower := strings.ToLower(c2)

			for key, exclusive := range mutuallyExclusive {
				if (strings.Contains(c1Lower, key) && strings.Contains(c2Lower, exclusive)) ||
					(strings.Contains(c1Lower, exclusive) && strings.Contains(c2Lower, key)) {
					return true
				}
			}
		}
	}

	return false
}

// ComparePerspectives compares two or more perspectives and identifies synergies and conflicts
func (pa *PerspectiveAnalyzer) ComparePerspectives(perspectives []*types.Perspective) (map[string]interface{}, error) {
	if len(perspectives) < 2 {
		return nil, fmt.Errorf("need at least 2 perspectives to compare")
	}

	pa.mu.RLock()
	defer pa.mu.RUnlock()

	result := make(map[string]interface{})

	// Find common concerns
	commonConcerns := pa.findCommonConcerns(perspectives)
	result["common_concerns"] = commonConcerns

	// Find common priorities
	commonPriorities := pa.findCommonPriorities(perspectives)
	result["common_priorities"] = commonPriorities

	// Find conflicts
	conflicts := pa.detectPerspectiveConflicts(perspectives)
	result["conflicts"] = conflicts

	// Generate synthesis
	result["synthesis"] = pa.generateSynthesis(commonConcerns, commonPriorities, conflicts)

	return result, nil
}

// findCommonConcerns identifies concerns shared by multiple perspectives
func (pa *PerspectiveAnalyzer) findCommonConcerns(perspectives []*types.Perspective) []string {
	concernCounts := make(map[string]int)

	for _, p := range perspectives {
		for _, concern := range p.Concerns {
			concernLower := strings.ToLower(concern)
			concernCounts[concernLower]++
		}
	}

	common := make([]string, 0)
	threshold := len(perspectives) / 2 // Appears in at least half
	for concern, count := range concernCounts {
		if count >= threshold {
			common = append(common, concern)
		}
	}

	return common
}

// findCommonPriorities identifies priorities shared by multiple perspectives
func (pa *PerspectiveAnalyzer) findCommonPriorities(perspectives []*types.Perspective) []string {
	priorityCounts := make(map[string]int)

	for _, p := range perspectives {
		for _, priority := range p.Priorities {
			priorityLower := strings.ToLower(priority)
			priorityCounts[priorityLower]++
		}
	}

	common := make([]string, 0)
	threshold := len(perspectives) / 2
	for priority, count := range priorityCounts {
		if count >= threshold {
			common = append(common, priority)
		}
	}

	return common
}

// generateSynthesis creates a synthesis of perspectives
func (pa *PerspectiveAnalyzer) generateSynthesis(commonConcerns, commonPriorities []string, conflicts []string) string {
	synthesis := "Analysis of multiple perspectives reveals: "

	if len(commonConcerns) > 0 {
		synthesis += fmt.Sprintf("Shared concerns include %s. ", strings.Join(commonConcerns, ", "))
	}

	if len(commonPriorities) > 0 {
		synthesis += fmt.Sprintf("Common priorities are %s. ", strings.Join(commonPriorities, ", "))
	}

	if len(conflicts) > 0 {
		synthesis += fmt.Sprintf("However, there are %d conflicts between perspectives that need resolution. ", len(conflicts))
	} else {
		synthesis += "Perspectives are largely aligned. "
	}

	synthesis += "A balanced approach should address shared concerns while navigating conflicts through compromise or phased implementation."

	return synthesis
}
