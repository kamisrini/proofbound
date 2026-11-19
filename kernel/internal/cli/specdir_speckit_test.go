package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

// CLI-INV-22: when HEAD's committed sdd profile declares kit "speckit" rooted at specs/, the P5
// intent provider set (sync intent all, sync all, verify) leaves out specdir, whose convention is
// also rooted at specs/ and would reject every Spec Kit file; otherwise it is unchanged.
func TestSpecKitProfileRetiresSpecdirFromTheIntentSet(t *testing.T) {
	commitProfile := func(t *testing.T, body string) string {
		t.Helper()
		root := gitRepo(t)
		if body != "" {
			if err := os.MkdirAll(filepath.Join(root, ".proofbound"), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, ".proofbound", "sdd-profile.json"), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		for _, args := range [][]string{{"add", "-A"}, {"-c", "user.name=t", "-c", "user.email=t@example.invalid", "commit", "-q", "--allow-empty", "-m", "x"}} {
			if out, err := exec.Command("git", append([]string{"-C", root}, args...)...).CombinedOutput(); err != nil {
				t.Fatalf("git %v: %v %s", args, err, out)
			}
		}
		return root
	}
	for name, c := range map[string]struct {
		profile string
		want    int
	}{
		"no profile":         {"", 2},
		"openspec profile":   {`{"kit":"openspec","changes":{"root":"openspec/changes"}}`, 2},
		"speckit on specs":   {`{"kit":"speckit","changes":{"root":"specs"}}`, 1},
		"speckit elsewhere":  {`{"kit":"speckit","changes":{"root":"features"}}`, 2},
		"unreadable profile": {`not json`, 2},
	} {
		t.Run(name, func(t *testing.T) {
			providers, err := intentProviders(commitProfile(t, c.profile), "all", time.Unix(1, 0))
			if err != nil {
				t.Fatal(err)
			}
			if len(providers) != c.want {
				t.Fatalf("%d providers, want %d", len(providers), c.want)
			}
			if c.want == 1 && providers[0].Name() != "records" {
				t.Fatalf("kept %s, want records", providers[0].Name())
			}
		})
	}
}
