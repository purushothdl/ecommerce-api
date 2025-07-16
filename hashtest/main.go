// hashtest/main.go
package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "my-super-secret-password-123"

	fmt.Println("Running bcrypt benchmark...")

	// Run it multiple times to see the variation
	for i := 0; i < 5; i++ {
		startTime := time.Now()

		_, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			log.Fatalf("Hashing failed: %v", err)
		}

		duration := time.Since(startTime)
		fmt.Printf("Run %d: Hashing took %v\n", i+1, duration)
	}
}