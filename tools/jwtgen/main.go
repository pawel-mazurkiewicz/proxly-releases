package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func main() {
	secret := flag.String("secret", "", "JWT signing secret (required)")
	subject := flag.String("sub", "admin", "subject claim")
	role := flag.String("role", "admin", "role claim")
	hours := flag.Int("hours", 24, "token validity in hours")
	flag.Parse()

	if *secret == "" {
		fmt.Fprintln(os.Stderr, "missing --secret")
		os.Exit(1)
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  *subject,
		"role": *role,
		"iat":  now.Unix(),
		"exp":  now.Add(time.Duration(*hours) * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(*secret))
	if err != nil {
		fmt.Fprintf(os.Stderr, "signing failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(signed)
}
