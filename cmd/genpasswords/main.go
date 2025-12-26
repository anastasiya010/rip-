package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	passwords := map[string]string{
		"user1123": "user1",
		"user2123": "user2",
		"admin123": "admin",
	}

	for password, login := range passwords {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			fmt.Printf("Error hashing %s: %v\n", password, err)
			continue
		}
		fmt.Printf("-- %s (%s): %s\n", login, password, string(hash))
	}
}

