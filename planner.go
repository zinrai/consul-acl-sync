package main

import "fmt"

// CalculatePlan compares the config against the live Consul state and returns
// the additive changes needed. It never plans a deletion.
func CalculatePlan(client *ConsulClient, cfg *Config) (*Plan, error) {
	plan := &Plan{}
	if err := planPolicies(client, cfg, plan); err != nil {
		return nil, err
	}
	if err := planTokens(client, cfg, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func planPolicies(client *ConsulClient, cfg *Config, plan *Plan) error {
	consulPolicies, err := client.ListPolicies()
	if err != nil {
		return fmt.Errorf("failed to list policies: %w", err)
	}
	byName := make(map[string]consulPolicy, len(consulPolicies))
	for _, p := range consulPolicies {
		byName[p.Name] = p
	}

	for _, desired := range cfg.Policies {
		current, ok := byName[desired.Name]
		if !ok {
			plan.PoliciesToCreate = append(plan.PoliciesToCreate, desired)
			continue
		}

		// Rules are absent from the list response, so fetch the full policy.
		full, err := client.PolicyRules(current.ID)
		if err != nil {
			return fmt.Errorf("failed to read policy %q: %w", desired.Name, err)
		}
		if policyNeedsUpdate(full, desired) {
			plan.PoliciesToUpdate = append(plan.PoliciesToUpdate, PolicyUpdate{ID: current.ID, Desired: desired})
		}
	}
	return nil
}

func planTokens(client *ConsulClient, cfg *Config, plan *Plan) error {
	consulTokens, err := client.ListTokens()
	if err != nil {
		return fmt.Errorf("failed to list tokens: %w", err)
	}
	byAccessor := make(map[string]consulToken, len(consulTokens))
	for _, t := range consulTokens {
		byAccessor[t.AccessorID] = t
	}

	for _, desired := range cfg.Tokens {
		current, ok := byAccessor[desired.AccessorID]
		if !ok {
			plan.TokensToCreate = append(plan.TokensToCreate, desired)
			continue
		}
		if tokenNeedsUpdate(current, desired) {
			plan.TokensToUpdate = append(plan.TokensToUpdate, desired)
		}
	}
	return nil
}

func policyNeedsUpdate(current consulPolicy, desired Policy) bool {
	if current.Description != desired.Description {
		return true
	}
	if normalizeRules(current.Rules) != normalizeRules(desired.Rules) {
		return true
	}
	return !stringSetEqual(current.Datacenters, desired.Datacenters)
}

func tokenNeedsUpdate(current consulToken, desired Token) bool {
	if current.Description != desired.Description {
		return true
	}
	return !stringSetEqual(policyLinkNames(current.Policies), desired.Policies)
}

func policyLinkNames(links []consulPolicyLink) []string {
	names := make([]string, 0, len(links))
	for _, l := range links {
		names = append(names, l.Name)
	}
	return names
}

// stringSetEqual reports whether two slices hold the same multiset of strings,
// ignoring order.
func stringSetEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	counts := make(map[string]int, len(a))
	for _, s := range a {
		counts[s]++
	}
	for _, s := range b {
		counts[s]--
		if counts[s] < 0 {
			return false
		}
	}
	return true
}
