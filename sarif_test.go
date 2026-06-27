package sarif

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWrite(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf,
		Tool{Name: "mytool", Version: "1.2.3", InformationURI: "https://example.com/mytool"},
		[]Rule{{ID: "complexity", Name: "CyclomaticComplexity", Description: "Function complexity"}},
		[]Result{
			{RuleID: "complexity", Level: "warning", Message: "F is complex", URI: "a/b.go", StartLine: 10, EndLine: 20},
			{RuleID: "complexity", Message: "G (file-level)", URI: "a/c.go"}, // no line → no region, default level
		})
	if err != nil {
		t.Fatalf("Write: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if doc["version"] != "2.1.0" {
		t.Errorf("version = %v, want 2.1.0", doc["version"])
	}
	if !strings.Contains(buf.String(), "json.schemastore.org/sarif-2.1.0.json") {
		t.Error("missing $schema")
	}

	runs := doc["runs"].([]any)
	r0 := runs[0].(map[string]any)
	driver := r0["tool"].(map[string]any)["driver"].(map[string]any)
	if driver["name"] != "mytool" || driver["version"] != "1.2.3" {
		t.Errorf("driver = %v", driver)
	}
	if driver["informationUri"] != "https://example.com/mytool" {
		t.Errorf("driver informationUri = %v", driver["informationUri"])
	}
	results := r0["results"].([]any)
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}

	// First result: has a region with the line range.
	res0 := results[0].(map[string]any)
	if res0["ruleId"] != "complexity" || res0["level"] != "warning" {
		t.Errorf("result0 ruleId/level = %v / %v", res0["ruleId"], res0["level"])
	}
	loc0 := res0["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)
	if loc0["artifactLocation"].(map[string]any)["uri"] != "a/b.go" {
		t.Errorf("result0 uri = %v", loc0["artifactLocation"])
	}
	if loc0["region"].(map[string]any)["startLine"].(float64) != 10 {
		t.Errorf("result0 startLine = %v", loc0["region"])
	}

	// Second result: no line → region omitted, level defaulted to warning.
	res1 := results[1].(map[string]any)
	if res1["level"] != "warning" {
		t.Errorf("result1 level defaulted = %v, want warning", res1["level"])
	}
	loc1 := res1["locations"].([]any)[0].(map[string]any)["physicalLocation"].(map[string]any)
	if _, hasRegion := loc1["region"]; hasRegion {
		t.Errorf("result1 should have no region (no line): %v", loc1)
	}
}

// TestWrite_OmitsEmptyToolFields: version + informationUri are omitted when the
// Tool leaves them empty (a generic analyzer may have neither).
func TestWrite_OmitsEmptyToolFields(t *testing.T) {
	var buf bytes.Buffer
	if err := Write(&buf, Tool{Name: "bare"}, nil, nil); err != nil {
		t.Fatalf("Write: %v", err)
	}
	s := buf.String()
	if strings.Contains(s, "informationUri") {
		t.Errorf("informationUri should be omitted when empty:\n%s", s)
	}
	var doc map[string]any
	if err := json.Unmarshal([]byte(s), &doc); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	r0 := doc["runs"].([]any)[0].(map[string]any)
	driver := r0["tool"].(map[string]any)["driver"].(map[string]any)
	if _, hasVersion := driver["version"]; hasVersion {
		t.Errorf("driver.version should be omitted when empty: %v", driver)
	}
	// Still a valid run with an (empty) results array.
	if _, ok := r0["results"].([]any); !ok {
		t.Errorf("results should be an array, got %T", r0["results"])
	}
}
