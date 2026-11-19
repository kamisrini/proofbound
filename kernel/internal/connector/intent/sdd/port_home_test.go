package sdd_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// LESSON class=embedded-default-port-contention (journal 2026-11-09 and 2026-11-19), compiled: the
// embedded database listens on ONE fixed default port (store.embeddedPort(0) == 55432, a product
// contract), and go test runs packages in parallel. Two packages whose untagged tests open the store
// at that port collide and make `make check` flaky. cli.Run always opens at the default port, so every
// untagged test that drives cli.Run, or opens store.Config without a Port, must live in THIS package,
// where tests run one after another. Tests elsewhere pick an ephemeral port (freeTestPort,
// temporaryPort) or carry the integration build tag.

// defaultPortOpeners are the syntactic signs of a default-port open: cli.Run from another package,
// forcing the embedded store with an empty DATABASE_URL (how a test inside package cli would reach
// it), and a store.Config literal without a Port. A detector cannot see every semantic path (a bare
// Run in package cli with DATABASE_URL already unset); these are the shapes the 2026-11-19 collision
// and the 2026-11-09 flake actually took.
var defaultPortOpeners = []*regexp.Regexp{
	regexp.MustCompile(`\bcli\.Run\(`),
	regexp.MustCompile(`Setenv\("DATABASE_URL", ""\)`),
	regexp.MustCompile(`store\.Config\{[^}\n]*\}`),
}

func TestDefaultPortStoreTestsLiveInOnePackage(t *testing.T) {
	kernel, err := filepath.Abs(filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	home, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	var offenders []string
	homeHits := 0
	err = filepath.WalkDir(kernel, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		src := string(b)
		if strings.HasPrefix(src, "//go:build integration") {
			return nil
		}
		for _, re := range defaultPortOpeners {
			for _, hit := range re.FindAllString(src, -1) {
				if strings.HasPrefix(hit, "store.Config{") && strings.Contains(hit, "Port:") {
					continue
				}
				if filepath.Dir(p) == home {
					homeHits++
					continue
				}
				rel, _ := filepath.Rel(kernel, p)
				offenders = append(offenders, rel+": "+hit)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(offenders) > 0 {
		t.Fatalf("untagged tests outside %s open the embedded store at the default port:\n%s", home, strings.Join(offenders, "\n"))
	}
	// Calibration: the detector must see this package's own default-port openers, or it is inert.
	if homeHits < 2 {
		t.Fatalf("detector inert: found %d default-port openers in this package", homeHits)
	}
}
