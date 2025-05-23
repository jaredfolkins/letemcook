package main_test

import (
	"flag"
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
		panic(err)
	}

	code := m.Run()
	shutdown()
	os.Exit(code)
}
