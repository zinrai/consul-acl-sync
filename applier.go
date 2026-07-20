package main

import "fmt"

// Apply performs the plan in dependency order, policies before tokens, since
// tokens reference policies by name. It stops at the first failure. Every step
// is idempotent, so a re-run resumes cleanly after a partial apply.
func Apply(client *ConsulClient, plan *Plan) error {
	for _, p := range plan.PoliciesToCreate {
		fmt.Printf("creating policy %q... ", p.Name)
		if err := client.CreatePolicy(p); err != nil {
			fmt.Println("failed")
			return err
		}
		fmt.Println("ok")
	}

	for _, u := range plan.PoliciesToUpdate {
		fmt.Printf("updating policy %q... ", u.Desired.Name)
		if err := client.UpdatePolicy(u.ID, u.Desired); err != nil {
			fmt.Println("failed")
			return err
		}
		fmt.Println("ok")
	}

	for _, t := range plan.TokensToCreate {
		fmt.Printf("creating token %s... ", tokenLabel(t))
		if err := client.CreateToken(t); err != nil {
			fmt.Println("failed")
			return err
		}
		fmt.Println("ok")
	}

	for _, t := range plan.TokensToUpdate {
		fmt.Printf("updating token %s... ", tokenLabel(t))
		if err := client.UpdateToken(t); err != nil {
			fmt.Println("failed")
			return err
		}
		fmt.Println("ok")
	}
	return nil
}

// tokenLabel annotates an opaque accessor id with its description when present.
func tokenLabel(t Token) string {
	if t.Description != "" {
		return fmt.Sprintf("%s %q", t.AccessorID, t.Description)
	}
	return t.AccessorID
}
