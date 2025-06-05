package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== Password Hash Generator ===")
	fmt.Println()

	// Get password
	fmt.Print("Enter password to hash: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)

	if password == "" {
		log.Fatal("Password cannot be empty")
	}

	// Generate hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to generate hash: %v", err)
	}

	fmt.Println()
	fmt.Println("✅ Password hashed successfully!")
	fmt.Printf("Hash: %s\n", string(hash))
	fmt.Println()
	fmt.Println("You can now use this hash in the database:")
	fmt.Printf("INSERT INTO admin_users (username, password, is_active, created_at, updated_at)\n")
	fmt.Printf("VALUES ('admin', '%s', true, NOW(), NOW());\n", string(hash))
}
