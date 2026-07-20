package main

// Config is the YAML configuration consul-acl-sync applies. The same file is
// read by consul-acl-diff.
type Config struct {
	Policies []Policy `yaml:"policies"`
	Tokens   []Token  `yaml:"tokens"`
}

// Policy is a Consul ACL policy, keyed by Name.
type Policy struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Rules       string   `yaml:"rules"`
	Datacenters []string `yaml:"datacenters"`
}

// Token is a Consul ACL token, keyed by AccessorID. SecretID is the credential,
// kept sops-encrypted in the config and pinned so creation is deterministic.
// Both AccessorID and SecretID are set at create time and immutable afterward.
type Token struct {
	AccessorID  string   `yaml:"accessor_id"`
	SecretID    string   `yaml:"secret_id"`
	Description string   `yaml:"description"`
	Policies    []string `yaml:"policies"`
}

// consulPolicy is the subset of the Consul policy API we read. The list
// endpoint omits Rules, so it is filled in per policy on demand.
type consulPolicy struct {
	ID          string   `json:"ID"`
	Name        string   `json:"Name"`
	Description string   `json:"Description"`
	Rules       string   `json:"Rules"`
	Datacenters []string `json:"Datacenters"`
}

// consulToken is the subset of the Consul token API we read. The list endpoint
// already carries the policy links.
type consulToken struct {
	AccessorID  string             `json:"AccessorID"`
	Description string             `json:"Description"`
	Policies    []consulPolicyLink `json:"Policies"`
}

type consulPolicyLink struct {
	ID   string `json:"ID"`
	Name string `json:"Name"`
}

// Plan is the additive set of changes to apply. consul-acl-sync never deletes:
// resources present only in Consul are left untouched. Surface them with
// consul-acl-diff and remove them by runbook.
type Plan struct {
	PoliciesToCreate []Policy
	PoliciesToUpdate []PolicyUpdate
	TokensToCreate   []Token
	TokensToUpdate   []Token
}

// PolicyUpdate pairs the desired policy with the existing Consul ID that the
// update endpoint addresses.
type PolicyUpdate struct {
	ID      string
	Desired Policy
}

// HasChanges reports whether the plan would modify anything.
func (p *Plan) HasChanges() bool {
	return len(p.PoliciesToCreate) > 0 ||
		len(p.PoliciesToUpdate) > 0 ||
		len(p.TokensToCreate) > 0 ||
		len(p.TokensToUpdate) > 0
}
