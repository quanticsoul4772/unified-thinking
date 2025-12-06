package types

import (
	"errors"
	"testing"
)

// Test request types for adapter testing
type SimpleRequest struct {
	Name    string  `json:"name"`
	Value   int     `json:"value"`
	Enabled bool    `json:"enabled"`
	Score   float64 `json:"score"`
}

type NestedRequest struct {
	ID     string            `json:"id"`
	Config map[string]string `json:"config"`
	Tags   []string          `json:"tags"`
}

type ValidatableRequest struct {
	Name string `json:"name"`
}

func (r ValidatableRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	return nil
}

// TestUnmarshalRequest tests the generic unmarshal function
func TestUnmarshalRequest(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		want    SimpleRequest
		wantErr bool
	}{
		{
			name: "basic fields",
			params: map[string]interface{}{
				"name":    "test",
				"value":   42,
				"enabled": true,
				"score":   0.95,
			},
			want: SimpleRequest{
				Name:    "test",
				Value:   42,
				Enabled: true,
				Score:   0.95,
			},
			wantErr: false,
		},
		{
			name: "json number conversion",
			params: map[string]interface{}{
				"name":    "test",
				"value":   float64(100), // JSON numbers are float64
				"enabled": false,
				"score":   3.14,
			},
			want: SimpleRequest{
				Name:    "test",
				Value:   100,
				Enabled: false,
				Score:   3.14,
			},
			wantErr: false,
		},
		{
			name:   "nil params returns zero value",
			params: nil,
			want:   SimpleRequest{},
		},
		{
			name:   "empty params returns zero value",
			params: map[string]interface{}{},
			want:   SimpleRequest{},
		},
		{
			name: "partial fields",
			params: map[string]interface{}{
				"name": "partial",
			},
			want: SimpleRequest{
				Name: "partial",
			},
		},
		{
			name: "extra fields ignored",
			params: map[string]interface{}{
				"name":  "test",
				"extra": "ignored",
			},
			want: SimpleRequest{
				Name: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UnmarshalRequest[SimpleRequest](tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("UnmarshalRequest() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalRequest_Nested(t *testing.T) {
	params := map[string]interface{}{
		"id": "test-123",
		"config": map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		"tags": []interface{}{"tag1", "tag2", "tag3"},
	}

	got, err := UnmarshalRequest[NestedRequest](params)
	if err != nil {
		t.Fatalf("UnmarshalRequest() error = %v", err)
	}

	if got.ID != "test-123" {
		t.Errorf("ID = %v, want test-123", got.ID)
	}

	if len(got.Tags) != 3 {
		t.Errorf("Tags len = %d, want 3", len(got.Tags))
	}

	if got.Tags[0] != "tag1" {
		t.Errorf("Tags[0] = %v, want tag1", got.Tags[0])
	}
}

func TestUnmarshalRequestWithValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid request",
			params: map[string]interface{}{
				"name": "valid",
			},
			wantErr: false,
		},
		{
			name:    "invalid - empty name",
			params:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "invalid - name is empty string",
			params: map[string]interface{}{
				"name": "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := UnmarshalRequestWithValidation[ValidatableRequest](tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalRequestWithValidation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMarshalResponse(t *testing.T) {
	tests := []struct {
		name    string
		resp    interface{}
		want    string
		wantErr bool
	}{
		{
			name: "simple struct",
			resp: SimpleRequest{
				Name:    "test",
				Value:   42,
				Enabled: true,
				Score:   0.95,
			},
			want: `{"name":"test","value":42,"enabled":true,"score":0.95}`,
		},
		{
			name: "nil response",
			resp: nil,
			want: "{}",
		},
		{
			name: "map response",
			resp: map[string]string{"key": "value"},
			want: `{"key":"value"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarshalResponse(tt.resp)
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if string(got) != tt.want {
				t.Errorf("MarshalResponse() = %s, want %s", string(got), tt.want)
			}
		})
	}
}

func TestMarshalResponseString(t *testing.T) {
	resp := StatusResponse{Status: "ok"}
	got, err := MarshalResponseString(resp)
	if err != nil {
		t.Fatalf("MarshalResponseString() error = %v", err)
	}
	want := `{"status":"ok"}`
	if got != want {
		t.Errorf("MarshalResponseString() = %s, want %s", got, want)
	}
}

func TestToJSONContent(t *testing.T) {
	resp := StatusResponse{Status: "success"}
	got := ToJSONContent(resp)
	want := `{"status":"success"}`
	if got != want {
		t.Errorf("ToJSONContent() = %s, want %s", got, want)
	}
}

func TestExtractString(t *testing.T) {
	params := map[string]interface{}{
		"name":   "test",
		"number": 42,
	}

	if got := ExtractString(params, "name"); got != "test" {
		t.Errorf("ExtractString(name) = %v, want test", got)
	}

	if got := ExtractString(params, "missing"); got != "" {
		t.Errorf("ExtractString(missing) = %v, want empty", got)
	}

	if got := ExtractString(params, "number"); got != "" {
		t.Errorf("ExtractString(number) = %v, want empty (wrong type)", got)
	}

	if got := ExtractString(nil, "name"); got != "" {
		t.Errorf("ExtractString(nil) = %v, want empty", got)
	}
}

func TestExtractInt(t *testing.T) {
	params := map[string]interface{}{
		"int_val":   42,
		"float_val": float64(100),
		"int64_val": int64(200),
		"string":    "not a number",
	}

	if got := ExtractInt(params, "int_val"); got != 42 {
		t.Errorf("ExtractInt(int_val) = %v, want 42", got)
	}

	if got := ExtractInt(params, "float_val"); got != 100 {
		t.Errorf("ExtractInt(float_val) = %v, want 100", got)
	}

	if got := ExtractInt(params, "int64_val"); got != 200 {
		t.Errorf("ExtractInt(int64_val) = %v, want 200", got)
	}

	if got := ExtractInt(params, "missing"); got != 0 {
		t.Errorf("ExtractInt(missing) = %v, want 0", got)
	}

	if got := ExtractInt(params, "string"); got != 0 {
		t.Errorf("ExtractInt(string) = %v, want 0", got)
	}

	if got := ExtractInt(nil, "int_val"); got != 0 {
		t.Errorf("ExtractInt(nil) = %v, want 0", got)
	}
}

func TestExtractFloat(t *testing.T) {
	params := map[string]interface{}{
		"float_val": 3.14,
		"int_val":   42,
		"int64_val": int64(100),
		"string":    "not a number",
	}

	if got := ExtractFloat(params, "float_val"); got != 3.14 {
		t.Errorf("ExtractFloat(float_val) = %v, want 3.14", got)
	}

	if got := ExtractFloat(params, "int_val"); got != 42.0 {
		t.Errorf("ExtractFloat(int_val) = %v, want 42.0", got)
	}

	if got := ExtractFloat(params, "int64_val"); got != 100.0 {
		t.Errorf("ExtractFloat(int64_val) = %v, want 100.0", got)
	}

	if got := ExtractFloat(params, "missing"); got != 0 {
		t.Errorf("ExtractFloat(missing) = %v, want 0", got)
	}

	if got := ExtractFloat(nil, "float_val"); got != 0 {
		t.Errorf("ExtractFloat(nil) = %v, want 0", got)
	}
}

func TestExtractBool(t *testing.T) {
	params := map[string]interface{}{
		"true_val":  true,
		"false_val": false,
		"string":    "true",
	}

	if got := ExtractBool(params, "true_val"); got != true {
		t.Errorf("ExtractBool(true_val) = %v, want true", got)
	}

	if got := ExtractBool(params, "false_val"); got != false {
		t.Errorf("ExtractBool(false_val) = %v, want false", got)
	}

	if got := ExtractBool(params, "missing"); got != false {
		t.Errorf("ExtractBool(missing) = %v, want false", got)
	}

	if got := ExtractBool(params, "string"); got != false {
		t.Errorf("ExtractBool(string) = %v, want false", got)
	}

	if got := ExtractBool(nil, "true_val"); got != false {
		t.Errorf("ExtractBool(nil) = %v, want false", got)
	}
}

func TestExtractStringSlice(t *testing.T) {
	params := map[string]interface{}{
		"string_slice":    []string{"a", "b", "c"},
		"interface_slice": []interface{}{"x", "y", "z"},
		"mixed_slice":     []interface{}{"a", 1, "c"},
		"not_slice":       "not a slice",
	}

	if got := ExtractStringSlice(params, "string_slice"); len(got) != 3 || got[0] != "a" {
		t.Errorf("ExtractStringSlice(string_slice) = %v, want [a b c]", got)
	}

	if got := ExtractStringSlice(params, "interface_slice"); len(got) != 3 || got[0] != "x" {
		t.Errorf("ExtractStringSlice(interface_slice) = %v, want [x y z]", got)
	}

	// Mixed slice only extracts strings
	if got := ExtractStringSlice(params, "mixed_slice"); len(got) != 2 {
		t.Errorf("ExtractStringSlice(mixed_slice) = %v, want 2 strings", got)
	}

	if got := ExtractStringSlice(params, "missing"); got != nil {
		t.Errorf("ExtractStringSlice(missing) = %v, want nil", got)
	}

	if got := ExtractStringSlice(params, "not_slice"); got != nil {
		t.Errorf("ExtractStringSlice(not_slice) = %v, want nil", got)
	}

	if got := ExtractStringSlice(nil, "string_slice"); got != nil {
		t.Errorf("ExtractStringSlice(nil) = %v, want nil", got)
	}
}

func TestExtractMap(t *testing.T) {
	nested := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}
	params := map[string]interface{}{
		"nested": nested,
		"string": "not a map",
	}

	if got := ExtractMap(params, "nested"); got == nil || got["key1"] != "value1" {
		t.Errorf("ExtractMap(nested) = %v, want map with key1", got)
	}

	if got := ExtractMap(params, "missing"); got != nil {
		t.Errorf("ExtractMap(missing) = %v, want nil", got)
	}

	if got := ExtractMap(params, "string"); got != nil {
		t.Errorf("ExtractMap(string) = %v, want nil", got)
	}

	if got := ExtractMap(nil, "nested"); got != nil {
		t.Errorf("ExtractMap(nil) = %v, want nil", got)
	}
}

// Test compatibility with EmptyRequest type
func TestUnmarshalRequest_EmptyRequest(t *testing.T) {
	params := map[string]interface{}{
		"extra": "ignored",
	}

	got, err := UnmarshalRequest[EmptyRequest](params)
	if err != nil {
		t.Fatalf("UnmarshalRequest[EmptyRequest]() error = %v", err)
	}

	want := EmptyRequest{}
	if got != want {
		t.Errorf("UnmarshalRequest[EmptyRequest]() = %+v, want %+v", got, want)
	}
}

// Test compatibility with StatusResponse type
func TestMarshalResponse_StatusResponse(t *testing.T) {
	resp := StatusResponse{Status: "ok"}
	got, err := MarshalResponse(resp)
	if err != nil {
		t.Fatalf("MarshalResponse() error = %v", err)
	}

	want := `{"status":"ok"}`
	if string(got) != want {
		t.Errorf("MarshalResponse() = %s, want %s", string(got), want)
	}
}

// Benchmark tests
func BenchmarkUnmarshalRequest(b *testing.B) {
	params := map[string]interface{}{
		"name":    "benchmark",
		"value":   100,
		"enabled": true,
		"score":   0.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = UnmarshalRequest[SimpleRequest](params)
	}
}

func BenchmarkExtractString(b *testing.B) {
	params := map[string]interface{}{
		"name": "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractString(params, "name")
	}
}

func BenchmarkMarshalResponse(b *testing.B) {
	resp := SimpleRequest{
		Name:    "benchmark",
		Value:   100,
		Enabled: true,
		Score:   0.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MarshalResponse(resp)
	}
}

// TestToParams tests the typed struct to params conversion
func TestToParams(t *testing.T) {
	tests := []struct {
		name string
		req  SimpleRequest
		want map[string]interface{}
	}{
		{
			name: "basic fields",
			req: SimpleRequest{
				Name:    "test",
				Value:   42,
				Enabled: true,
				Score:   0.95,
			},
			want: map[string]interface{}{
				"name":    "test",
				"value":   float64(42), // JSON numbers are float64
				"enabled": true,
				"score":   0.95,
			},
		},
		{
			name: "zero values",
			req:  SimpleRequest{},
			want: map[string]interface{}{
				"name":    "",
				"value":   float64(0),
				"enabled": false,
				"score":   float64(0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToParams(tt.req)
			if got["name"] != tt.want["name"] {
				t.Errorf("ToParams().name = %v, want %v", got["name"], tt.want["name"])
			}
			if got["value"] != tt.want["value"] {
				t.Errorf("ToParams().value = %v, want %v", got["value"], tt.want["value"])
			}
			if got["enabled"] != tt.want["enabled"] {
				t.Errorf("ToParams().enabled = %v, want %v", got["enabled"], tt.want["enabled"])
			}
			if got["score"] != tt.want["score"] {
				t.Errorf("ToParams().score = %v, want %v", got["score"], tt.want["score"])
			}
		})
	}
}

// TestToParams_Nested tests nested struct conversion
func TestToParams_Nested(t *testing.T) {
	req := NestedRequest{
		ID:     "test-123",
		Config: map[string]string{"key": "value"},
		Tags:   []string{"tag1", "tag2"},
	}

	params := ToParams(req)

	if params["id"] != "test-123" {
		t.Errorf("ToParams().id = %v, want test-123", params["id"])
	}

	config, ok := params["config"].(map[string]interface{})
	if !ok {
		t.Fatalf("ToParams().config is not a map")
	}
	if config["key"] != "value" {
		t.Errorf("ToParams().config.key = %v, want value", config["key"])
	}

	tags, ok := params["tags"].([]interface{})
	if !ok {
		t.Fatalf("ToParams().tags is not a slice")
	}
	if len(tags) != 2 || tags[0] != "tag1" {
		t.Errorf("ToParams().tags = %v, want [tag1 tag2]", tags)
	}
}

// TestMustToParams tests the panic version
func TestMustToParams(t *testing.T) {
	req := SimpleRequest{Name: "test", Value: 42}
	params := MustToParams(req)

	if params["name"] != "test" {
		t.Errorf("MustToParams().name = %v, want test", params["name"])
	}
}

// TestToParams_RoundTrip verifies ToParams and UnmarshalRequest are inverses
func TestToParams_RoundTrip(t *testing.T) {
	original := SimpleRequest{
		Name:    "roundtrip",
		Value:   999,
		Enabled: true,
		Score:   3.14,
	}

	// Convert to params
	params := ToParams(original)

	// Convert back to struct
	result, err := UnmarshalRequest[SimpleRequest](params)
	if err != nil {
		t.Fatalf("UnmarshalRequest() error = %v", err)
	}

	// Verify round-trip preserves values
	if result != original {
		t.Errorf("Round-trip failed: got %+v, want %+v", result, original)
	}
}

func BenchmarkToParams(b *testing.B) {
	req := SimpleRequest{
		Name:    "benchmark",
		Value:   100,
		Enabled: true,
		Score:   0.99,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ToParams(req)
	}
}
