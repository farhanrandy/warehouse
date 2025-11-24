package main

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// go run tools/hash_password.go <plain_password>
func main() {
    if len(os.Args) < 2 {
        log.Fatal("usage: go run tools/hash_password.go <password>")
    }
    pwd := os.Args[1]
    b, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(b))
}
