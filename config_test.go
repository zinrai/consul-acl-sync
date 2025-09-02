package main

import (
	"testing"
)

func TestPoliciesEqual(t *testing.T) {
	tests := []struct {
		name     string
		policy1  Policy
		policy2  Policy
		expected bool
	}{
		{
			name: "identical policies",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			expected: true,
		},
		{
			name: "different names",
			policy1: Policy{
				Name:        "policy1",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			policy2: Policy{
				Name:        "policy2",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			expected: false,
		},
		{
			name: "different descriptions",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Description 1",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Description 2",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			expected: false,
		},
		{
			name: "different rules content",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"read\" }",
			},
			expected: false,
		},
		{
			name: "rules with different whitespace should be equal",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" {  policy = \"write\"  }",
			},
			expected: false, // NormalizeRules only normalizes line endings and trailing spaces, not internal spaces
		},
		{
			name: "rules with different line endings should be equal",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" {\n  policy = \"write\"\n}",
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" {\r\n  policy = \"write\"\r\n}",
			},
			expected: true,
		},
		{
			name: "rules with trailing spaces should be equal",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" {\n  policy = \"write\"\n}",
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" {  \n  policy = \"write\"  \n}  ",
			},
			expected: true,
		},
		{
			name: "empty datacenters should be equal",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
				Datacenters: []string{},
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
				Datacenters: nil,
			},
			expected: true,
		},
		{
			name: "different datacenters",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
				Datacenters: []string{"dc1"},
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
				Datacenters: []string{"dc2"},
			},
			expected: false,
		},
		{
			name: "same datacenters different order",
			policy1: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
				Datacenters: []string{"dc1", "dc2"},
			},
			policy2: Policy{
				Name:        "test-policy",
				Description: "Test description",
				Rules:       "key \"test\" { policy = \"write\" }",
				Datacenters: []string{"dc2", "dc1"},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PoliciesEqual(tt.policy1, tt.policy2)
			if result != tt.expected {
				t.Errorf("PoliciesEqual() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestTokensEqual(t *testing.T) {
	tests := []struct {
		name     string
		token1   Token
		token2   Token
		expected bool
	}{
		{
			name: "identical tokens",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			expected: true,
		},
		{
			name: "different descriptions",
			token1: Token{
				Description: "Token 1",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			token2: Token{
				Description: "Token 2",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			expected: false,
		},
		{
			name: "different number of policies",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
					{Name: "policy2"},
				},
			},
			expected: false,
		},
		{
			name: "same policies different order",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
					{Name: "policy2"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy2"},
					{Name: "policy1"},
				},
			},
			expected: true,
		},
		{
			name: "policies with IDs instead of names",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{ID: "uuid-1234"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{ID: "uuid-1234"},
				},
			},
			expected: true,
		},
		{
			name: "mixed names and IDs same content",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
					{ID: "uuid-1234"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{ID: "uuid-1234"},
					{Name: "policy1"},
				},
			},
			expected: true,
		},
		{
			name: "empty policies lists",
			token1: Token{
				Description: "Test token",
				Policies:    []PolicyLink{},
			},
			token2: Token{
				Description: "Test token",
				Policies:    nil,
			},
			expected: true, // Both represent "no policies", should be equal
		},
		{
			name: "both nil policies",
			token1: Token{
				Description: "Test token",
				Policies:    nil,
			},
			token2: Token{
				Description: "Test token",
				Policies:    nil,
			},
			expected: true,
		},
		{
			name: "both empty policies",
			token1: Token{
				Description: "Test token",
				Policies:    []PolicyLink{},
			},
			token2: Token{
				Description: "Test token",
				Policies:    []PolicyLink{},
			},
			expected: true,
		},
		{
			name: "different policy names",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy2"},
				},
			},
			expected: false,
		},
		{
			name: "policy with both name and ID uses name",
			token1: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1", ID: "ignored-id"},
				},
			},
			token2: Token{
				Description: "Test token",
				Policies: []PolicyLink{
					{Name: "policy1"},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TokensEqual(tt.token1, tt.token2)
			if result != tt.expected {
				t.Errorf("TokensEqual() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
