package main

import (
	"fmt"
	"uptime-gopher-std/checks"
)

func main() {
	check := checks.DomainCheck()

	result := check.Run("google.com", map[string]string{})

	fmt.Println("Success:", result.Success)
	fmt.Println("Severity:", result.Severity)
	fmt.Println("Message:", result.Message)
}
