// Package config holds the per-project rule overrides.
//
// v3 hard-codes severity in each rule. v4 lets each project ship a
// .apiforge.json that overrides per-rule severity (downgrade
// warnings to info, escalate the operation-id rule to error,
// disable the method-wording rule entirely) without touching the
// rule code.
package config

import (
	"encoding/json"
	"os"
)

// Override maps rule name → severity ("error" / "warning" / "info"
// / "off"). "off" disables the rule.
type Config struct {
	Overrides map[string]string `json:"overrides,omitempty"`
}

// Load reads `.apiforge.json` from the given path. Missing file =
// empty config (no overrides). Malformed JSON returns an error so
// the caller can decide whether to abort or continue with defaults.
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Overrides: map[string]string{}}, nil
		}
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	if c.Overrides == nil {
		c.Overrides = map[string]string{}
	}
	return &c, nil
}

// Apply rewrites a list of findings according to the per-rule
// overrides. Findings whose rule maps to "off" are dropped.
func (c *Config) Apply(findings []Finding) []Finding {
	if c == nil || len(c.Overrides) == 0 {
		return findings
	}
	out := make([]Finding, 0, len(findings))
	for _, f := range findings {
		override, ok := c.Overrides[f.Rule]
		if !ok {
			out = append(out, f)
			continue
		}
		if override == "off" {
			continue
		}
		f.Severity = override
		out = append(out, f)
	}
	return out
}

// Finding is a minimal mirror of lint.Finding so this package
// doesn't import `lint` (avoids an import cycle when lint wants
// to use config).
type Finding struct {
	Rule     string
	Severity string
	Path     string
	Method   string
	Message  string
}
