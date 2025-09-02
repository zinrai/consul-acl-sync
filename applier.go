package main

import (
	"fmt"
	"strings"
)

// Apply applies the differences to Consul
func Apply(client *ConsulClient, diff *DiffResult) error {
	if !diff.HasChanges() {
		fmt.Println("No changes to apply.")
		return nil
	}

	fmt.Println("\nApplying changes...")
	fmt.Println("=" + strings.Repeat("=", 50))

	var errors []string
	successCount := 0

	// Apply in order to respect dependencies
	// 1. Create policies first (tokens depend on them)
	for _, policy := range diff.PoliciesToCreate {
		fmt.Printf("Creating policy '%s'... ", policy.Name)
		if _, err := client.CreatePolicy(policy); err != nil {
			fmt.Printf("FAILED\n")
			errors = append(errors, fmt.Sprintf("  - Failed to create policy '%s': %v", policy.Name, err))
		} else {
			fmt.Printf("OK\n")
			successCount++
		}
	}

	// 2. Update policies
	for _, update := range diff.PoliciesToUpdate {
		fmt.Printf("Updating policy '%s'... ", update.Desired.Name)
		if err := client.UpdatePolicy(update.Current.ID, update.Desired); err != nil {
			fmt.Printf("FAILED\n")
			errors = append(errors, fmt.Sprintf("  - Failed to update policy '%s': %v", update.Desired.Name, err))
		} else {
			fmt.Printf("OK\n")
			successCount++
		}
	}

	// 3. Create tokens (after policies exist)
	for _, token := range diff.TokensToCreate {
		desc := getTokenDescription(token)
		fmt.Printf("Creating token '%s'... ", desc)
		if err := client.CreateToken(token); err != nil {
			fmt.Printf("FAILED\n")
			errors = append(errors, fmt.Sprintf("  - Failed to create token '%s': %v", desc, err))
		} else {
			fmt.Printf("OK\n")
			successCount++
		}
	}

	// 4. Update tokens
	for _, update := range diff.TokensToUpdate {
		desc := getTokenDescription(update.Desired)
		fmt.Printf("Updating token '%s'... ", desc)
		if err := client.UpdateToken(update.Current.AccessorID, update.Desired); err != nil {
			fmt.Printf("FAILED\n")
			errors = append(errors, fmt.Sprintf("  - Failed to update token '%s': %v", desc, err))
		} else {
			fmt.Printf("OK\n")
			successCount++
		}
	}

	// Print summary
	fmt.Println("\n" + strings.Repeat("-", 51))
	totalChanges := len(diff.PoliciesToCreate) + len(diff.PoliciesToUpdate) +
		len(diff.TokensToCreate) + len(diff.TokensToUpdate)

	if len(errors) > 0 {
		fmt.Printf("\nApply incomplete! %d of %d changes applied successfully.\n", successCount, totalChanges)
		fmt.Println("\nErrors:")
		for _, err := range errors {
			fmt.Println(err)
		}
		return fmt.Errorf("some changes failed to apply")
	}

	fmt.Printf("\nApply complete! %d changes applied successfully.\n", successCount)
	return nil
}

// getTokenDescription returns a description for a token
func getTokenDescription(token Token) string {
	if token.Description != "" {
		return token.Description
	}
	return "(no description)"
}

// ConfirmApply asks for user confirmation before applying changes
func ConfirmApply(diff *DiffResult) bool {
	if !diff.HasChanges() {
		return false
	}

	totalChanges := len(diff.PoliciesToCreate) + len(diff.PoliciesToUpdate) +
		len(diff.TokensToCreate) + len(diff.TokensToUpdate)

	fmt.Printf("\nDo you want to perform these actions?\n")
	fmt.Printf("  Consul ACL Sync will perform %d actions.\n\n", totalChanges)
	fmt.Print("  Enter a value (yes/no): ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	return response == "yes" || response == "y"
}
