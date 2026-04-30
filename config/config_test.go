package config

import (
	"os"
	"path/filepath"
	"testing"
)

func tmp(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, ".apiforge.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadMissingFileReturnsEmpty(t *testing.T) {
	c, err := Load("/tmp/does-not-exist-apiforge.json")
	if err != nil {
		t.Fatalf("missing file should not error: %v", err)
	}
	if len(c.Overrides) != 0 {
		t.Errorf("expected empty overrides, got %v", c.Overrides)
	}
}

func TestLoadValidConfig(t *testing.T) {
	c, err := Load(tmp(t, `{"overrides": {"operation-id-present": "error", "method-wording-matches-path": "off"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if c.Overrides["operation-id-present"] != "error" {
		t.Errorf("override missing")
	}
}

func TestLoadMalformedJSONReturnsError(t *testing.T) {
	_, err := Load(tmp(t, "not json"))
	if err == nil {
		t.Errorf("expected error on malformed json")
	}
}

func TestApplyDropsOffFindings(t *testing.T) {
	c := &Config{Overrides: map[string]string{"x": "off"}}
	findings := []Finding{{Rule: "x", Severity: "warning"}, {Rule: "y", Severity: "error"}}
	out := c.Apply(findings)
	if len(out) != 1 || out[0].Rule != "y" {
		t.Errorf("expected 'x' dropped, got %+v", out)
	}
}

func TestApplyRewritesSeverity(t *testing.T) {
	c := &Config{Overrides: map[string]string{"x": "error"}}
	findings := []Finding{{Rule: "x", Severity: "warning"}}
	out := c.Apply(findings)
	if out[0].Severity != "error" {
		t.Errorf("severity not rewritten: %+v", out[0])
	}
}

func TestApplyPassesUnknownRulesThrough(t *testing.T) {
	c := &Config{Overrides: map[string]string{"x": "off"}}
	findings := []Finding{{Rule: "y", Severity: "warning"}}
	out := c.Apply(findings)
	if len(out) != 1 || out[0].Severity != "warning" {
		t.Errorf("unknown rule should pass through")
	}
}
