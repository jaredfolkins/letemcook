package integration_test

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// No longer need shared server setup - each test creates its own instance
	log.Println("Starting integration tests with parallel infrastructure")

	// Run tests
	code := m.Run()

	if code != 0 {
		log.Printf("Integration tests failed with exit code %d", code)
	} else {
		log.Println("All integration tests completed successfully")
	}

	os.Exit(code)
}

func init() {
	fmt.Println("Initializing integration test package with parallel infrastructure")
}
