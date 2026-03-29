package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Rule represents a single household cooking or shopping rule.
// Type and Params encode the rule's behavior; Description is human-readable.
type Rule struct {
	ID          string         `json:"id"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	Params      map[string]any `json:"params,omitempty"`
}

// Rules is an ordered list of household rules.
type Rules []Rule

// Load reads Rules from a JSON file. Returns an empty Rules if the file does not exist.
func Load(path string) (Rules, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Rules{}, nil
	}
	if err != nil {
		return nil, err
	}
	var r Rules
	return r, json.Unmarshal(data, &r)
}

// Save writes Rules to a JSON file with pretty-printing.
func (r Rules) Save(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Upsert adds a rule or replaces the existing rule with the same ID.
func (r Rules) Upsert(rule Rule) Rules {
	for i, existing := range r {
		if existing.ID == rule.ID {
			updated := make(Rules, len(r))
			copy(updated, r)
			updated[i] = rule
			return updated
		}
	}
	return append(r, rule)
}

// ToMarkdown formats the rules as a markdown list for LLM consumption.
func (r Rules) ToMarkdown() string {
	if len(r) == 0 {
		return "No household rules configured."
	}
	var sb strings.Builder
	sb.WriteString("## Household Rules\n\n")
	for _, rule := range r {
		sb.WriteString(fmt.Sprintf("### %s (`%s`)\n", rule.ID, rule.Type))
		sb.WriteString(fmt.Sprintf("%s\n", rule.Description))
		if len(rule.Params) > 0 {
			sb.WriteString("\nParameters:\n")
			for k, v := range rule.Params {
				sb.WriteString(fmt.Sprintf("- **%s**: %v\n", k, v))
			}
		}
		sb.WriteString("\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}
