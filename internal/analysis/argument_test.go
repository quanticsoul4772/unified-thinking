package analysis

import (
	"testing"
)

func TestDecomposeArgument(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	tests := []struct {
		name     string
		argument string
		wantErr  bool
	}{
		{
			name:     "simple premise-conclusion",
			argument: "We should adopt policy X because studies show it reduces Y by 30%",
			wantErr:  false,
		},
		{
			name:     "multiple premises",
			argument: "All humans are mortal. Socrates is human. Therefore, Socrates is mortal.",
			wantErr:  false,
		},
		{
			name:     "empty argument",
			argument: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := analyzer.DecomposeArgument(tt.argument)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecomposeArgument() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if result == nil {
					t.Fatal("expected decomposition result")
				}
				if result.ID == "" {
					t.Error("expected argument ID")
				}
				if result.MainClaim == "" {
					t.Error("expected main claim")
				}
			}
		})
	}
}

func TestGenerateCounterArguments(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	// First decompose an argument
	argument := "We should adopt remote work because it increases productivity and reduces costs"
	decomposition, err := analyzer.DecomposeArgument(argument)
	if err != nil {
		t.Fatalf("DecomposeArgument() error = %v", err)
	}

	// Generate counter arguments
	counterArgs, err := analyzer.GenerateCounterArguments(decomposition.ID)
	if err != nil {
		t.Fatalf("GenerateCounterArguments() error = %v", err)
	}

	if len(counterArgs) == 0 {
		t.Error("expected at least one counter argument")
	}

	for _, ca := range counterArgs {
		if ca.ID == "" {
			t.Error("counter argument missing ID")
		}
		if ca.Strategy == "" {
			t.Error("counter argument missing strategy")
		}
		if ca.CounterClaim == "" {
			t.Error("counter argument missing counter claim")
		}
	}
}

func TestGetArgument(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	argument := "The sky is blue because of Rayleigh scattering"
	decomposition, err := analyzer.DecomposeArgument(argument)
	if err != nil {
		t.Fatalf("DecomposeArgument() error = %v", err)
	}

	// Retrieve the argument
	retrieved, err := analyzer.GetArgument(decomposition.ID)
	if err != nil {
		t.Fatalf("GetArgument() error = %v", err)
	}

	if retrieved.ID != decomposition.ID {
		t.Errorf("retrieved ID = %s, want %s", retrieved.ID, decomposition.ID)
	}

	if retrieved.MainClaim == "" {
		t.Error("expected main claim in retrieved argument")
	}
}

func TestGetArgument_NotFound(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	_, err := analyzer.GetArgument("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent argument")
	}
}

func TestGetCounterArgument(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	// Decompose and generate counter arguments
	argument := "Exercise improves health"
	decomposition, err := analyzer.DecomposeArgument(argument)
	if err != nil {
		t.Fatalf("DecomposeArgument() error = %v", err)
	}

	counterArgs, err := analyzer.GenerateCounterArguments(decomposition.ID)
	if err != nil {
		t.Fatalf("GenerateCounterArguments() error = %v", err)
	}

	if len(counterArgs) == 0 {
		t.Skip("no counter arguments generated")
	}

	// Retrieve a counter argument
	retrieved, err := analyzer.GetCounterArgument(counterArgs[0].ID)
	if err != nil {
		t.Fatalf("GetCounterArgument() error = %v", err)
	}

	if retrieved.ID != counterArgs[0].ID {
		t.Errorf("retrieved ID = %s, want %s", retrieved.ID, counterArgs[0].ID)
	}
}

func TestExtractMainClaim(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	tests := []struct {
		name     string
		argument string
		wantText bool
	}{
		{
			name:     "conclusion with therefore",
			argument: "All men are mortal. Socrates is a man. Therefore, Socrates is mortal.",
			wantText: true,
		},
		{
			name:     "conclusion with thus",
			argument: "Prices are rising. Thus, inflation is increasing.",
			wantText: true,
		},
		{
			name:     "simple statement",
			argument: "Climate change is real.",
			wantText: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := analyzer.extractMainClaim(tt.argument)
			if tt.wantText && claim == "" {
				t.Error("expected main claim to be extracted")
			}
		})
	}
}

func TestExtractPremises(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	argument := "All men are mortal because they are living beings. Socrates is a man. Therefore, Socrates is mortal."
	premises := analyzer.extractPremises(argument)

	if len(premises) == 0 {
		t.Error("expected at least one premise")
	}

	for _, p := range premises {
		if p.Statement == "" {
			t.Error("premise statement should not be empty")
		}
	}
}

func TestDetermineArgumentType(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	text := "All humans are mortal. Socrates is human. Therefore, Socrates is mortal."
	chain := []*InferenceStep{{ID: "step1", Rule: "universal"}}

	argType := analyzer.determineArgumentType(text, chain)
	if argType == "" {
		t.Error("expected non-empty argument type")
	}

	validTypes := map[ArgumentType]bool{
		ArgumentDeductive: true,
		ArgumentInductive: true,
		ArgumentAbductive: true,
	}

	if !validTypes[argType] {
		t.Errorf("unexpected argument type: %v", argType)
	}
}

func TestCalculateArgumentStrength(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	tests := []struct {
		name     string
		argType  ArgumentType
		premises []*Premise
		chain    []*InferenceStep
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "deductive argument",
			argType:  ArgumentDeductive,
			premises: []*Premise{{Statement: "All men are mortal", Certainty: 1.0}},
			chain:    []*InferenceStep{{Confidence: 0.9}},
			wantMin:  0.7,
			wantMax:  1.0,
		},
		{
			name:     "inductive argument",
			argType:  ArgumentInductive,
			premises: []*Premise{{Statement: "Most birds fly", Certainty: 0.8}},
			chain:    []*InferenceStep{{Confidence: 0.7}},
			wantMin:  0.4,
			wantMax:  0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := analyzer.calculateArgumentStrength(tt.premises, tt.chain, tt.argType)
			if strength < tt.wantMin || strength > tt.wantMax {
				t.Errorf("calculateArgumentStrength() = %v, want between %v and %v", strength, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGenerateCounterArgumentsNotFound(t *testing.T) {
	analyzer := NewArgumentAnalyzer()

	_, err := analyzer.GenerateCounterArguments("nonexistent-id")
	if err == nil {
		t.Fatal("expected error for nonexistent argument")
	}
}
