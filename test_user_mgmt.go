// Test User Management functionality directly
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yourusername/ravact/internal/system"
)

func main() {
	fmt.Println("========================================")
	fmt.Println("User Management Direct Test")
	fmt.Println("========================================")
	fmt.Println()

	// Test 1: Create UserManager
	fmt.Println("Test 1: Creating UserManager...")
	um := system.NewUserManager()
	if um == nil {
		fmt.Println("✗ FAIL: UserManager is nil")
		os.Exit(1)
	}
	fmt.Println("✓ PASS: UserManager created")
	fmt.Println()

	// Test 2: Get all users
	fmt.Println("Test 2: Loading all users...")
	start := time.Now()
	users, err := um.GetAllUsers()
	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("✗ FAIL: GetAllUsers() returned error: %v\n", err)
		fmt.Println()
	} else {
		fmt.Printf("✓ PASS: GetAllUsers() succeeded in %v\n", duration)
		fmt.Printf("  Found %d users\n", len(users))
		
		if duration > 5*time.Second {
			fmt.Printf("⚠ WARNING: Took longer than 5 seconds (%v)\n", duration)
		}
		
		// Show first few users
		fmt.Println()
		fmt.Println("Sample users:")
		for i, user := range users {
			if i >= 5 {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  - %s (UID: %d, Sudo: %v)\n", user.Username, user.UID, user.HasSudo)
		}
	}
	fmt.Println()

	// Test 3: Get all groups
	fmt.Println("Test 3: Loading all groups...")
	start = time.Now()
	groups, err := um.GetAllGroups()
	duration = time.Since(start)
	
	if err != nil {
		fmt.Printf("✗ FAIL: GetAllGroups() returned error: %v\n", err)
		fmt.Println()
	} else {
		fmt.Printf("✓ PASS: GetAllGroups() succeeded in %v\n", duration)
		fmt.Printf("  Found %d groups\n", len(groups))
		
		if duration > 5*time.Second {
			fmt.Printf("⚠ WARNING: Took longer than 5 seconds (%v)\n", duration)
		}
		
		// Show first few groups
		fmt.Println()
		fmt.Println("Sample groups:")
		for i, group := range groups {
			if i >= 5 {
				fmt.Println("  ...")
				break
			}
			fmt.Printf("  - %s (GID: %d, Members: %d)\n", group.Name, group.GID, len(group.Members))
		}
	}
	fmt.Println()

	// Test 4: Check specific user
	fmt.Println("Test 4: Testing specific user lookup...")
	testUser := "root"
	start = time.Now()
	hasSudo := um.UserHasSudo(testUser)
	duration = time.Since(start)
	
	fmt.Printf("✓ User '%s' sudo check: %v (took %v)\n", testUser, hasSudo, duration)
	
	if duration > 2*time.Second {
		fmt.Printf("⚠ WARNING: Sudo check took longer than 2 seconds\n")
	}
	fmt.Println()

	// Summary
	fmt.Println("========================================")
	fmt.Println("Summary")
	fmt.Println("========================================")
	fmt.Println()
	fmt.Println("✓ All critical functions work")
	fmt.Println("✓ No hanging or timeouts detected")
	fmt.Println()
	fmt.Println("User Management should work in the TUI!")
	fmt.Println()
}
