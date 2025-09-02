package main

import "encoding/json"

// Config represents the YAML configuration file structure
type Config struct {
	Policies []Policy `yaml:"policies"`
	Tokens   []Token  `yaml:"tokens"`
}

// Policy represents a Consul ACL policy
type Policy struct {
	ID          string   `yaml:"-" json:"ID,omitempty"`
	Name        string   `yaml:"name" json:"Name"`
	Description string   `yaml:"description,omitempty" json:"Description,omitempty"`
	Rules       string   `yaml:"rules" json:"Rules"`
	Datacenters []string `yaml:"datacenters,omitempty" json:"Datacenters,omitempty"`
}

// Token represents a Consul ACL token
type Token struct {
	AccessorID  string       `yaml:"-" json:"AccessorID,omitempty"`
	SecretID    string       `yaml:"-" json:"SecretID,omitempty"`
	Description string       `yaml:"description,omitempty" json:"Description,omitempty"`
	Policies    []PolicyLink `yaml:"policies"`
}

// PolicyLink represents a reference to a policy in a token
type PolicyLink struct {
	ID   string `json:"ID,omitempty"`
	Name string `json:"Name,omitempty"`
}

// MarshalYAML customizes YAML marshaling for Token
func (t Token) MarshalYAML() (interface{}, error) {
	type tokenYAML struct {
		Description string   `yaml:"description,omitempty"`
		Policies    []string `yaml:"policies"`
	}

	policies := make([]string, len(t.Policies))
	for i, p := range t.Policies {
		if p.Name != "" {
			policies[i] = p.Name
		} else {
			policies[i] = p.ID
		}
	}

	return &tokenYAML{
		Description: t.Description,
		Policies:    policies,
	}, nil
}

// UnmarshalYAML customizes YAML unmarshaling for Token
func (t *Token) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type tokenYAML struct {
		Description string   `yaml:"description,omitempty"`
		Policies    []string `yaml:"policies"`
	}

	var aux tokenYAML
	if err := unmarshal(&aux); err != nil {
		return err
	}

	t.Description = aux.Description
	t.Policies = make([]PolicyLink, len(aux.Policies))
	for i, name := range aux.Policies {
		t.Policies[i] = PolicyLink{Name: name}
	}

	return nil
}

// DiffResult represents the differences between current and desired state
type DiffResult struct {
	PoliciesToCreate []Policy
	PoliciesToUpdate []PolicyUpdate
	TokensToCreate   []Token
	TokensToUpdate   []TokenUpdate
}

// PolicyUpdate represents a policy that needs updating
type PolicyUpdate struct {
	Current Policy
	Desired Policy
}

// TokenUpdate represents a token that needs updating
type TokenUpdate struct {
	Current Token
	Desired Token
}

// HasChanges returns true if there are any changes to apply
func (d *DiffResult) HasChanges() bool {
	return len(d.PoliciesToCreate) > 0 ||
		len(d.PoliciesToUpdate) > 0 ||
		len(d.TokensToCreate) > 0 ||
		len(d.TokensToUpdate) > 0
}

// MarshalJSON customizes JSON marshaling for Token
func (t Token) MarshalJSON() ([]byte, error) {
	type tokenAlias Token
	aux := struct {
		tokenAlias
		Policies []PolicyLink `json:"Policies,omitempty"`
	}{
		tokenAlias: tokenAlias(t),
		Policies:   t.Policies,
	}

	if len(t.Policies) > 0 {
		aux.Policies = make([]PolicyLink, len(t.Policies))
		for i, p := range t.Policies {
			aux.Policies[i] = p
		}
	}

	return json.Marshal(&aux)
}
