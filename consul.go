package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ConsulClient is a client for the Consul ACL HTTP API.
type ConsulClient struct {
	addr   string
	token  string
	client *http.Client
}

func NewConsulClient(addr, token string) *ConsulClient {
	if addr == "" {
		addr = "http://127.0.0.1:8500"
	}
	addr = strings.TrimRight(addr, "/")
	return &ConsulClient{addr: addr, token: token, client: &http.Client{}}
}

func (c *ConsulClient) do(method, path string, body, out interface{}) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequest(method, c.addr+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Consul-Token", c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request to %s failed: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, strings.TrimSpace(string(b)))
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

// ListPolicies returns all policies. The list endpoint does not include Rules,
// so PolicyRules fills them in per policy.
func (c *ConsulClient) ListPolicies() ([]consulPolicy, error) {
	var policies []consulPolicy
	if err := c.do(http.MethodGet, "/v1/acl/policies", nil, &policies); err != nil {
		return nil, err
	}
	return policies, nil
}

// PolicyRules fetches a single policy so its Rules can be compared.
func (c *ConsulClient) PolicyRules(id string) (consulPolicy, error) {
	var p consulPolicy
	if err := c.do(http.MethodGet, "/v1/acl/policy/"+id, nil, &p); err != nil {
		return consulPolicy{}, err
	}
	return p, nil
}

// ListTokens returns all tokens. Each entry already carries its policy links.
func (c *ConsulClient) ListTokens() ([]consulToken, error) {
	var tokens []consulToken
	if err := c.do(http.MethodGet, "/v1/acl/tokens", nil, &tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

type policyRequest struct {
	ID          string   `json:"ID,omitempty"`
	Name        string   `json:"Name"`
	Description string   `json:"Description,omitempty"`
	Rules       string   `json:"Rules"`
	Datacenters []string `json:"Datacenters,omitempty"`
}

func (c *ConsulClient) CreatePolicy(p Policy) error {
	body := policyRequest{Name: p.Name, Description: p.Description, Rules: p.Rules, Datacenters: p.Datacenters}
	return c.do(http.MethodPut, "/v1/acl/policy", body, nil)
}

func (c *ConsulClient) UpdatePolicy(id string, p Policy) error {
	body := policyRequest{ID: id, Name: p.Name, Description: p.Description, Rules: p.Rules, Datacenters: p.Datacenters}
	return c.do(http.MethodPut, "/v1/acl/policy/"+id, body, nil)
}

type tokenRequest struct {
	AccessorID  string              `json:"AccessorID,omitempty"`
	SecretID    string              `json:"SecretID,omitempty"`
	Description string              `json:"Description,omitempty"`
	Policies    []policyLinkRequest `json:"Policies"`
}

type policyLinkRequest struct {
	Name string `json:"Name"`
}

// tokenBody builds a token request. Consul resolves policy links by name, so
// the policies created earlier in the same run are already resolvable.
func tokenBody(t Token) tokenRequest {
	links := make([]policyLinkRequest, 0, len(t.Policies))
	for _, name := range t.Policies {
		links = append(links, policyLinkRequest{Name: name})
	}
	return tokenRequest{
		AccessorID:  t.AccessorID,
		SecretID:    t.SecretID,
		Description: t.Description,
		Policies:    links,
	}
}

func (c *ConsulClient) CreateToken(t Token) error {
	return c.do(http.MethodPut, "/v1/acl/token", tokenBody(t), nil)
}

// UpdateToken addresses the token by AccessorID in the path. SecretID is
// omitted because it is immutable after creation.
func (c *ConsulClient) UpdateToken(t Token) error {
	body := tokenBody(t)
	body.SecretID = ""
	return c.do(http.MethodPut, "/v1/acl/token/"+t.AccessorID, body, nil)
}
