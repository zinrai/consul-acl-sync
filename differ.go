package main

import (
	"fmt"
	"strings"
)

// CalculateDiff compares current and desired state and returns the differences
func CalculateDiff(client *ConsulClient, desired *Config) (*DiffResult, error) {
	diff := &DiffResult{}

	// Calculate policy differences
	for _, desiredPolicy := range desired.Policies {
		currentPolicy, err := client.GetPolicyByName(desiredPolicy.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check policy '%s': %w", desiredPolicy.Name, err)
		}

		if currentPolicy == nil {
			// Policy doesn't exist, need to create
			diff.PoliciesToCreate = append(diff.PoliciesToCreate, desiredPolicy)
		} else if !PoliciesEqual(*currentPolicy, desiredPolicy) {
			// Policy exists but is different, need to update
			diff.PoliciesToUpdate = append(diff.PoliciesToUpdate, PolicyUpdate{
				Current: *currentPolicy,
				Desired: desiredPolicy,
			})
		}
		// If policies are equal, no action needed
	}

	// Calculate token differences
	for _, desiredToken := range desired.Tokens {
		currentToken, err := client.GetTokenByDescription(desiredToken.Description)
		if err != nil {
			return nil, fmt.Errorf("failed to check token '%s': %w", desiredToken.Description, err)
		}

		if currentToken == nil {
			// Token doesn't exist, need to create
			diff.TokensToCreate = append(diff.TokensToCreate, desiredToken)
		} else if !TokensEqual(*currentToken, desiredToken) {
			// Token exists but is different, need to update
			// Preserve the accessor ID for updates
			desiredToken.AccessorID = currentToken.AccessorID
			diff.TokensToUpdate = append(diff.TokensToUpdate, TokenUpdate{
				Current: *currentToken,
				Desired: desiredToken,
			})
		}
		// If tokens are equal, no action needed
	}

	return diff, nil
}

// PrintPlan prints the diff in a human-readable format
func PrintPlan(diff *DiffResult) {
	fmt.Println("\nConsul ACL Sync Plan:")
	fmt.Println("=" + strings.Repeat("=", 50))

	if !diff.HasChanges() {
		fmt.Println("\nNo changes required. Infrastructure is up-to-date.")
		return
	}

	// Print policy changes
	printPolicyChanges(diff)

	// Print token changes
	printTokenChanges(diff)

	// Print summary
	printSummary(diff)
}

// printPolicyChanges prints policy-related changes
func printPolicyChanges(diff *DiffResult) {
	if !hasPolicyChanges(diff) {
		return
	}

	fmt.Println("\n## Policies")
	fmt.Println()

	for _, policy := range diff.PoliciesToCreate {
		printPolicyCreate(policy)
	}

	for _, update := range diff.PoliciesToUpdate {
		printPolicyUpdate(update)
	}
}

// printTokenChanges prints token-related changes
func printTokenChanges(diff *DiffResult) {
	if !hasTokenChanges(diff) {
		return
	}

	fmt.Println("\n## Tokens")
	fmt.Println()

	for _, token := range diff.TokensToCreate {
		printTokenCreate(token)
	}

	for _, update := range diff.TokensToUpdate {
		printTokenUpdate(update)
	}
}

// printPolicyCreate prints a policy creation entry
func printPolicyCreate(policy Policy) {
	fmt.Printf("  + %s\n", policy.Name)
	if policy.Description != "" {
		fmt.Printf("      Description: %s\n", policy.Description)
	}
}

// printPolicyUpdate prints a policy update entry
func printPolicyUpdate(update PolicyUpdate) {
	fmt.Printf("  ~ %s\n", update.Desired.Name)

	if update.Current.Description != update.Desired.Description {
		fmt.Printf("      Description: %q → %q\n", update.Current.Description, update.Desired.Description)
	}

	currentRules := NormalizeRules(update.Current.Rules)
	desiredRules := NormalizeRules(update.Desired.Rules)
	if currentRules != desiredRules {
		fmt.Println("      Rules: (changed)")
	}
}

// printTokenCreate prints a token creation entry
func printTokenCreate(token Token) {
	desc := token.Description
	if desc == "" {
		desc = "(no description)"
	}
	fmt.Printf("  + %s\n", desc)

	fmt.Print("      Policies: [")
	printPolicyList(token.Policies)
	fmt.Println("]")
}

// printTokenUpdate prints a token update entry
func printTokenUpdate(update TokenUpdate) {
	desc := update.Desired.Description
	if desc == "" {
		desc = "(no description)"
	}
	fmt.Printf("  ~ %s\n", desc)

	if !policiesListEqual(update.Current.Policies, update.Desired.Policies) {
		fmt.Print("      Policies: [")
		printPolicyList(update.Current.Policies)
		fmt.Print("] → [")
		printPolicyList(update.Desired.Policies)
		fmt.Println("]")
	}
}

// printPolicyList prints a list of policies inline
func printPolicyList(policies []PolicyLink) {
	for i, p := range policies {
		if i > 0 {
			fmt.Print(", ")
		}
		if p.Name != "" {
			fmt.Print(p.Name)
		} else {
			fmt.Print(p.ID)
		}
	}
}

// printSummary prints the change summary
func printSummary(diff *DiffResult) {
	fmt.Println("\n" + strings.Repeat("-", 51))
	fmt.Printf("\nPlan: %d to create, %d to update\n",
		len(diff.PoliciesToCreate)+len(diff.TokensToCreate),
		len(diff.PoliciesToUpdate)+len(diff.TokensToUpdate))
}

// hasPolicyChanges checks if there are any policy changes
func hasPolicyChanges(diff *DiffResult) bool {
	return len(diff.PoliciesToCreate) > 0 ||
		len(diff.PoliciesToUpdate) > 0
}

// hasTokenChanges checks if there are any token changes
func hasTokenChanges(diff *DiffResult) bool {
	return len(diff.TokensToCreate) > 0 ||
		len(diff.TokensToUpdate) > 0
}

// policiesListEqual compares two lists of PolicyLink
func policiesListEqual(a, b []PolicyLink) bool {
	if len(a) != len(b) {
		return false
	}

	aMap := make(map[string]bool)
	for _, p := range a {
		key := p.Name
		if key == "" {
			key = p.ID
		}
		aMap[key] = true
	}

	for _, p := range b {
		key := p.Name
		if key == "" {
			key = p.ID
		}
		if !aMap[key] {
			return false
		}
	}

	return true
}
