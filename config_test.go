package main

import "testing"

func TestNormalizeRules(t *testing.T) {
	tests := []struct {
		name  string
		a, b  string
		equal bool
	}{
		{"crlf vs lf", "key \"x\" {\r\n  policy = \"read\"\r\n}", "key \"x\" {\n  policy = \"read\"\n}", true},
		{"trailing whitespace", "key \"x\" {  \n  policy = \"read\"  \n}  ", "key \"x\" {\n  policy = \"read\"\n}", true},
		{"internal whitespace differs", "key \"x\" { policy = \"read\" }", "key \"x\" {  policy = \"read\"  }", false},
		{"content differs", "policy = \"read\"", "policy = \"write\"", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeRules(tt.a) == normalizeRules(tt.b); got != tt.equal {
				t.Errorf("normalizeRules equality = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestStringSetEqual(t *testing.T) {
	tests := []struct {
		name  string
		a, b  []string
		equal bool
	}{
		{"same order", []string{"a", "b"}, []string{"a", "b"}, true},
		{"different order", []string{"a", "b"}, []string{"b", "a"}, true},
		{"different length", []string{"a"}, []string{"a", "b"}, false},
		{"different content", []string{"a"}, []string{"b"}, false},
		{"empty vs nil", []string{}, nil, true},
		{"duplicates matter", []string{"a", "a"}, []string{"a", "b"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringSetEqual(tt.a, tt.b); got != tt.equal {
				t.Errorf("stringSetEqual = %v, want %v", got, tt.equal)
			}
		})
	}
}

func TestPolicyNeedsUpdate(t *testing.T) {
	desired := Policy{
		Name:        "p",
		Description: "d",
		Rules:       "key \"x\" {\n  policy = \"read\"\n}",
		Datacenters: []string{"dc1"},
	}
	current := consulPolicy{
		Name:        "p",
		Description: "d",
		Rules:       "key \"x\" {\r\n  policy = \"read\"\r\n}",
		Datacenters: []string{"dc1"},
	}

	if policyNeedsUpdate(current, desired) {
		t.Error("identical policy (crlf only) should not need update")
	}

	changedDesc := current
	changedDesc.Description = "different"
	if !policyNeedsUpdate(changedDesc, desired) {
		t.Error("description change should need update")
	}

	changedRules := current
	changedRules.Rules = "key \"x\" {\n  policy = \"write\"\n}"
	if !policyNeedsUpdate(changedRules, desired) {
		t.Error("rules change should need update")
	}

	changedDC := current
	changedDC.Datacenters = []string{"dc2"}
	if !policyNeedsUpdate(changedDC, desired) {
		t.Error("datacenter change should need update")
	}
}

func TestTokenNeedsUpdate(t *testing.T) {
	desired := Token{AccessorID: "a", Description: "web", Policies: []string{"p1", "p2"}}
	current := consulToken{
		AccessorID:  "a",
		Description: "web",
		Policies:    []consulPolicyLink{{Name: "p2"}, {Name: "p1"}},
	}

	if tokenNeedsUpdate(current, desired) {
		t.Error("same policy set in different order should not need update")
	}

	changedDesc := current
	changedDesc.Description = "different"
	if !tokenNeedsUpdate(changedDesc, desired) {
		t.Error("description change should need update")
	}

	changedPolicies := current
	changedPolicies.Policies = []consulPolicyLink{{Name: "p1"}}
	if !tokenNeedsUpdate(changedPolicies, desired) {
		t.Error("policy set change should need update")
	}
}
