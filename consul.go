package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ConsulClient represents a client for Consul HTTP API
type ConsulClient struct {
	addr   string
	token  string
	client *http.Client
}

// NewConsulClient creates a new Consul API client
func NewConsulClient(addr, token string) *ConsulClient {
	// Set default address if not provided
	if addr == "" {
		addr = "http://localhost:8500"
	}

	// Ensure address doesn't have trailing slash
	addr = strings.TrimRight(addr, "/")

	return &ConsulClient{
		addr:   addr,
		token:  token,
		client: &http.Client{},
	}
}

// makeRequest performs an HTTP request to Consul API
func (c *ConsulClient) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.addr+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("X-Consul-Token", c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// GetPolicyByName retrieves a policy by name
func (c *ConsulClient) GetPolicyByName(name string) (*Policy, error) {
	// First, get all policies to find the one with matching name
	resp, err := c.makeRequest("GET", "/v1/acl/policies", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list policies: %s (status: %d)", string(body), resp.StatusCode)
	}

	var policies []struct {
		ID   string `json:"ID"`
		Name string `json:"Name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&policies); err != nil {
		return nil, fmt.Errorf("failed to decode policies: %w", err)
	}

	// Find the policy with matching name
	for _, p := range policies {
		if p.Name == name {
			// Get full policy details
			return c.GetPolicy(p.ID)
		}
	}

	return nil, nil
}

// GetPolicy retrieves a single policy by ID
func (c *ConsulClient) GetPolicy(id string) (*Policy, error) {
	resp, err := c.makeRequest("GET", "/v1/acl/policy/"+id, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get policy: %s (status: %d)", string(body), resp.StatusCode)
	}

	var policy Policy
	if err := json.NewDecoder(resp.Body).Decode(&policy); err != nil {
		return nil, fmt.Errorf("failed to decode policy: %w", err)
	}

	return &policy, nil
}

// CreatePolicy creates a new ACL policy
func (c *ConsulClient) CreatePolicy(policy Policy) (*Policy, error) {
	resp, err := c.makeRequest("PUT", "/v1/acl/policy", policy)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create policy '%s': %s", policy.Name, string(body))
	}

	var created Policy
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, fmt.Errorf("failed to decode created policy: %w", err)
	}

	return &created, nil
}

// UpdatePolicy updates an existing ACL policy
func (c *ConsulClient) UpdatePolicy(id string, policy Policy) error {
	policy.ID = id
	resp, err := c.makeRequest("PUT", "/v1/acl/policy/"+id, policy)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update policy '%s': %s", policy.Name, string(body))
	}

	return nil
}

// GetTokenByDescription retrieves a token by description
func (c *ConsulClient) GetTokenByDescription(description string) (*Token, error) {
	resp, err := c.makeRequest("GET", "/v1/acl/tokens", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list tokens: %s (status: %d)", string(body), resp.StatusCode)
	}

	var tokens []Token
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, fmt.Errorf("failed to decode tokens: %w", err)
	}

	// Find the token with matching description
	for _, t := range tokens {
		if t.Description == description {
			// Get full token details
			return c.GetToken(t.AccessorID)
		}
	}

	return nil, nil
}

// GetToken retrieves a single token by accessor ID
func (c *ConsulClient) GetToken(accessorID string) (*Token, error) {
	resp, err := c.makeRequest("GET", "/v1/acl/token/"+accessorID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get token: %s (status: %d)", string(body), resp.StatusCode)
	}

	var token Token
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	return &token, nil
}

// CreateToken creates a new ACL token
func (c *ConsulClient) CreateToken(token Token) error {
	// Ensure policies have IDs, not names
	for i, policy := range token.Policies {
		if policy.ID == "" && policy.Name != "" {
			// Resolve name to ID
			p, err := c.GetPolicyByName(policy.Name)
			if err != nil {
				return fmt.Errorf("failed to resolve policy '%s': %w", policy.Name, err)
			}
			if p == nil {
				return fmt.Errorf("policy '%s' not found", policy.Name)
			}
			token.Policies[i].ID = p.ID
			token.Policies[i].Name = "" // Clear name, use only ID
		}
	}

	resp, err := c.makeRequest("PUT", "/v1/acl/token", token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create token '%s': %s", token.Description, string(body))
	}

	return nil
}

// UpdateToken updates an existing ACL token
func (c *ConsulClient) UpdateToken(accessorID string, token Token) error {
	token.AccessorID = accessorID

	// Ensure policies have IDs, not names
	for i, policy := range token.Policies {
		if policy.ID == "" && policy.Name != "" {
			// Resolve name to ID
			p, err := c.GetPolicyByName(policy.Name)
			if err != nil {
				return fmt.Errorf("failed to resolve policy '%s': %w", policy.Name, err)
			}
			if p == nil {
				return fmt.Errorf("policy '%s' not found", policy.Name)
			}
			token.Policies[i].ID = p.ID
			token.Policies[i].Name = "" // Clear name, use only ID
		}
	}

	resp, err := c.makeRequest("PUT", "/v1/acl/token/"+accessorID, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update token '%s': %s", token.Description, string(body))
	}

	return nil
}
