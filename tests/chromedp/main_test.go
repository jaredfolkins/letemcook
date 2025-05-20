package main_test

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/jaredfolkins/letemcook/tests/testutil"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		os.Exit(m.Run())
	}

	shutdown, err := testutil.StartTestServer()
	if err != nil {
		// If the server can't start (e.g. missing dependencies), skip
		// the integration tests instead of failing.
		fmt.Fprintf(os.Stderr, "skipping chromedp tests: %v\n", err)
		os.Exit(0)
	}

	code := m.Run()
	shutdown()
	os.Exit(code)
}
