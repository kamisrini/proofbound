package manifest

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

const (
	h40 = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	h64 = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
)

func caseRef(p string) CaseRef { return CaseRef{Path: p, SHA256: h64} }

func allOutcomes(status CaseOutcomeStatus) Outcomes {
	o := Outcome{Status: status, Evidence: []string{}}
	return Outcomes{o, o, o, o, o, o, o, o, o}
}

// validCaseManifest is a complete, non-runtime case manifest: every outcome is UNVERIFIED,
// so the derived case_status is UNVERIFIED and nothing is PASS by absence.
func validCaseManifest() CaseManifest {
	return CaseManifest{
		Schema:           CaseManifestSchema,
		CaseID:           "replay-iso-001",
		Subject:          ReplayIsolatedSubject,
		CaseStatus:       string(CaseStatusUnverified),
		EvidenceRevision: h40,
		ContractRefs: ContractRefs{
			LockAddendum:        caseRef("docs/plans/lock.md"),
			FixtureMatrix:       caseRef("docs/plans/matrix.md"),
			ImplementationScope: caseRef("docs/plans/scope.md"),
		},
		Inputs: CaseInputs{
			Lock:              caseRef("inputs/lock.json"),
			Source:            caseRef("inputs/source.json"),
			Candidate:         caseRef("inputs/candidate.json"),
			SetupSources:      caseRef("inputs/setup.json"),
			ProjectorManifest: caseRef("inputs/projector.json"),
		},
		Commands:  []Command{},
		Witnesses: []Witness{},
		Outcomes:  allOutcomes(CaseStatusUnverified),
		Artifacts: []Artifact{},
		Review:    CaseReview{AuthorCommit: h40, IndependentVerdict: CaseVerdict{Status: "PENDING"}},
	}
}

func TestCaseManifestCanonicalRoundTripAndDigest(t *testing.T) {
	base := validCaseManifest()
	raw, err := CanonicalCaseManifestBytes(base)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := ParseCaseManifest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, base) {
		t.Fatalf("round-trip differs:\n%+v\n%+v", parsed, base)
	}
	wantSum := sha256.Sum256(raw)
	want := hex.EncodeToString(wantSum[:])
	got, err := CaseManifestDigest(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("digest=%s want=%s", got, want)
	}
	// A changed field changes the digest.
	changed := base
	changed.CaseID = "replay-iso-002"
	changedRaw, err := CanonicalCaseManifestBytes(changed)
	if err != nil {
		t.Fatal(err)
	}
	if string(changedRaw) == string(raw) {
		t.Fatal("distinct case ids produced identical canonical bytes")
	}
}

func TestParseCaseManifestRejectsShapeAndCanonicality(t *testing.T) {
	base, err := CanonicalCaseManifestBytes(validCaseManifest())
	if err != nil {
		t.Fatal(err)
	}
	// Non-canonical bytes (trailing whitespace) are rejected.
	if _, err := ParseCaseManifest(append(append([]byte{}, base...), ' ')); err == nil {
		t.Fatal("non-canonical JSON accepted")
	}
	shapeMutations := map[string]func(map[string]any){
		"unknown top field": func(m map[string]any) { m["extra"] = true },
		"missing field":     func(m map[string]any) { delete(m, "review") },
		"null field":        func(m map[string]any) { m["inputs"] = nil },
		"unknown ref field": func(m map[string]any) {
			m["contract_refs"].(map[string]any)["lock_addendum"].(map[string]any)["extra"] = 1
		},
		"missing outcome key": func(m map[string]any) {
			delete(m["outcomes"].(map[string]any), "cleanup")
		},
		"unknown outcome key": func(m map[string]any) {
			m["outcomes"].(map[string]any)["surprise"] = map[string]any{"status": "PENDING", "evidence": []any{}}
		},
		"review verdict shape": func(m map[string]any) {
			m["review"].(map[string]any)["independent_verdict"].(map[string]any)["extra"] = 1
		},
	}
	for name, mutate := range shapeMutations {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseCaseManifest(remarshal(t, base, mutate)); err == nil {
				t.Fatal("invalid shape accepted")
			}
		})
	}
	valueMutations := map[string]func(*CaseManifest){
		"bad schema":      func(c *CaseManifest) { c.Schema = "x" },
		"bad case_id":     func(c *CaseManifest) { c.CaseID = "Bad_ID" },
		"bad subject":     func(c *CaseManifest) { c.Subject = "Other" },
		"bad revision":    func(c *CaseManifest) { c.EvidenceRevision = "short" },
		"unsafe ref path": func(c *CaseManifest) { c.Inputs.Lock.Path = "../escape" },
		"bad ref digest":  func(c *CaseManifest) { c.ContractRefs.FixtureMatrix.SHA256 = "nothex" },
		"bad verdict":     func(c *CaseManifest) { c.Review.IndependentVerdict.Status = "MAYBE" },
		"bad author":      func(c *CaseManifest) { c.Review.AuthorCommit = "nope" },
	}
	for name, mutate := range valueMutations {
		t.Run(name, func(t *testing.T) {
			c := validCaseManifest()
			mutate(&c)
			if err := ValidateCaseManifest(c); err == nil {
				t.Fatal("invalid value accepted")
			}
		})
	}
}

func TestCaseManifestRejectsKeyAndUniquenessViolations(t *testing.T) {
	cmd := Command{
		ID: "run", Argv: []string{"go", "test"}, CWD: ".", NetworkPolicy: "none",
		EnvironmentDigest: h64, ExitCode: 0, Stdout: caseRef("out/stdout"), Stderr: caseRef("out/stderr"),
	}
	mutations := map[string]func(*CaseManifest){
		"nil commands":         func(c *CaseManifest) { c.Commands = nil },
		"nil witnesses":        func(c *CaseManifest) { c.Witnesses = nil },
		"nil artifacts":        func(c *CaseManifest) { c.Artifacts = nil },
		"empty argv":           func(c *CaseManifest) { bad := cmd; bad.Argv = []string{}; c.Commands = []Command{bad} },
		"empty argv element":   func(c *CaseManifest) { bad := cmd; bad.Argv = []string{""}; c.Commands = []Command{bad} },
		"bad network policy":   func(c *CaseManifest) { bad := cmd; bad.NetworkPolicy = "wan"; c.Commands = []Command{bad} },
		"unsafe cwd":           func(c *CaseManifest) { bad := cmd; bad.CWD = "/etc"; c.Commands = []Command{bad} },
		"duplicate command id": func(c *CaseManifest) { c.Commands = []Command{cmd, cmd} },
		"duplicate witness": func(c *CaseManifest) {
			w := Witness{Kind: "net", Path: "w", SHA256: h64}
			c.Witnesses = []Witness{w, w}
		},
		"duplicate artifact": func(c *CaseManifest) {
			a := Artifact{Role: "r", Path: "p", SizeBytes: 1, SHA256: h64}
			c.Artifacts = []Artifact{a, a}
		},
		"negative artifact size": func(c *CaseManifest) {
			c.Artifacts = []Artifact{{Role: "r", Path: "p", SizeBytes: -1, SHA256: h64}}
		},
	}
	for name, mutate := range mutations {
		t.Run(name, func(t *testing.T) {
			c := validCaseManifest()
			mutate(&c)
			if err := ValidateCaseManifest(c); err == nil {
				t.Fatal("invalid collection accepted")
			}
		})
	}
	// A well-formed command/witness/artifact set is accepted.
	ok := validCaseManifest()
	ok.Commands = []Command{cmd}
	ok.Witnesses = []Witness{{Kind: "ledger-before", Path: "w/before", SHA256: h64}}
	if err := ValidateCaseManifest(ok); err != nil {
		t.Fatalf("valid collections rejected: %v", err)
	}
}

func TestCaseManifestOutcomeEvidenceRules(t *testing.T) {
	// PASS requires evidence naming an existing artifact role.
	passNoEvidence := validCaseManifest()
	passNoEvidence.Outcomes.SourceBinding = Outcome{Status: CaseStatusPass, Evidence: []string{}}
	if err := ValidateCaseManifest(passNoEvidence); err == nil {
		t.Fatal("PASS by absence of evidence accepted")
	}

	passUnknownRole := validCaseManifest()
	passUnknownRole.Outcomes.SourceBinding = Outcome{Status: CaseStatusPass, Evidence: []string{"ghost"}}
	if err := ValidateCaseManifest(passUnknownRole); err == nil {
		t.Fatal("PASS citing unknown artifact role accepted")
	}

	// Non-PASS/FAIL statuses must carry empty evidence.
	unverifiedWithEvidence := validCaseManifest()
	unverifiedWithEvidence.Artifacts = []Artifact{{Role: "ev", Path: "p", SizeBytes: 1, SHA256: h64}}
	unverifiedWithEvidence.Outcomes.Cleanup = Outcome{Status: CaseStatusUnverified, Evidence: []string{"ev"}}
	if err := ValidateCaseManifest(unverifiedWithEvidence); err == nil {
		t.Fatal("UNVERIFIED with evidence accepted")
	}

	// nil (absent) evidence is rejected even for UNVERIFIED.
	nilEvidence := validCaseManifest()
	nilEvidence.Outcomes.NoFetch = Outcome{Status: CaseStatusUnverified, Evidence: nil}
	if err := ValidateCaseManifest(nilEvidence); err == nil {
		t.Fatal("nil evidence array accepted")
	}

	// A fully-evidenced PASS case is accepted and derives PASS.
	pass := passAllOutcomes()
	if err := ValidateCaseManifest(pass); err != nil {
		t.Fatalf("evidenced PASS case rejected: %v", err)
	}
	// Duplicate evidence role within one outcome is rejected.
	dupEvidence := pass
	dupEvidence.Outcomes.SourceBinding = Outcome{Status: CaseStatusPass, Evidence: []string{"ev", "ev"}}
	dupEvidence.CaseStatus = string(CaseStatusPass)
	if err := ValidateCaseManifest(dupEvidence); err == nil {
		t.Fatal("duplicate evidence role accepted")
	}
}

// passAllOutcomes builds a case whose every outcome is an evidenced PASS, deriving case_status PASS.
func passAllOutcomes() CaseManifest {
	c := validCaseManifest()
	c.Artifacts = []Artifact{{Role: "ev", Path: "artifacts/ev", SizeBytes: 1, SHA256: h64}}
	pass := Outcome{Status: CaseStatusPass, Evidence: []string{"ev"}}
	c.Outcomes = Outcomes{pass, pass, pass, pass, pass, pass, pass, pass, pass}
	c.CaseStatus = string(CaseStatusPass)
	return c
}

func TestDeriveCaseStatusPrecedenceAndNoPassByAbsence(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*Outcomes)
		want   CaseOutcomeStatus
	}{
		{"all pass", func(o *Outcomes) {}, CaseStatusPass},
		{"one undeclared folds to unverified", func(o *Outcomes) {
			o.IntentBinding.Status = CaseStatusUndeclared
			o.IntentBinding.Evidence = []string{}
		}, CaseStatusUnverified},
		{"one unverified", func(o *Outcomes) { o.Cleanup.Status = CaseStatusUnverified; o.Cleanup.Evidence = []string{} }, CaseStatusUnverified},
		{"one fail beats unverified", func(o *Outcomes) {
			o.Cleanup.Status = CaseStatusUnverified
			o.Cleanup.Evidence = []string{}
			o.ProofBinding.Status = CaseStatusFail
		}, CaseStatusFail},
		{"one pending beats all", func(o *Outcomes) {
			o.ProofBinding.Status = CaseStatusFail
			o.NoFetch.Status = CaseStatusPending
			o.NoFetch.Evidence = []string{}
		}, CaseStatusPending},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := passAllOutcomes()
			tc.mutate(&c.Outcomes)
			got, err := DeriveCaseStatus(c)
			if err != nil {
				t.Fatal(err)
			}
			if got != string(tc.want) {
				t.Fatalf("derived=%s want=%s", got, tc.want)
			}
			// A manifest whose declared case_status disagrees with the derived fold is rejected.
			c.CaseStatus = "PASS"
			if tc.want != CaseStatusPass {
				if err := ValidateCaseManifest(c); err == nil {
					t.Fatal("mismatched case_status accepted")
				}
			}
		})
	}
	// An unknown outcome status fails the derivation closed.
	bad := validCaseManifest()
	bad.Outcomes.SourceBinding.Status = "WAT"
	if _, err := DeriveCaseStatus(bad); err == nil {
		t.Fatal("unknown outcome status derived a result")
	}
}

// populatedCaseManifest has one command, witness, and artifact so the nested-shape code paths
// in validateCaseShape are exercised through ParseCaseManifest, not only ValidateCaseManifest.
func populatedCaseManifest() CaseManifest {
	c := validCaseManifest()
	c.Commands = []Command{{
		ID: "run", Argv: []string{"go", "test"}, CWD: ".", NetworkPolicy: "none",
		EnvironmentDigest: h64, ExitCode: 0, Stdout: caseRef("out/stdout"), Stderr: caseRef("out/stderr"),
	}}
	c.Witnesses = []Witness{{Kind: "ledger-before", Path: "w/before", SHA256: h64}}
	c.Artifacts = []Artifact{{Role: "ev", Path: "artifacts/ev", SizeBytes: 1, SHA256: h64}}
	return c
}

// TestParseCaseManifestNestedShapeErrorChecks drives the command/witness/artifact and array
// error-handling branches through Parse, so a flipped error check (!= -> ==) accepts a malformed
// entry and the mutant dies.
func TestParseCaseManifestNestedShapeErrorChecks(t *testing.T) {
	base, err := CanonicalCaseManifestBytes(populatedCaseManifest())
	if err != nil {
		t.Fatal(err)
	}
	// The populated manifest round-trips through Parse (nested shape checks must pass).
	parsed, err := ParseCaseManifest(base)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(parsed, populatedCaseManifest()) {
		t.Fatalf("populated round-trip differs:\n%+v", parsed)
	}
	firstOf := func(key string) func(map[string]any) map[string]any {
		return func(m map[string]any) map[string]any {
			return m[key].([]any)[0].(map[string]any)
		}
	}
	rejections := map[string]func(map[string]any){
		"command unknown field":  func(m map[string]any) { firstOf("commands")(m)["extra"] = 1 },
		"command missing field":  func(m map[string]any) { delete(firstOf("commands")(m), "cwd") },
		"command stdout unknown": func(m map[string]any) { firstOf("commands")(m)["stdout"].(map[string]any)["extra"] = 1 },
		"command stdout missing": func(m map[string]any) { delete(firstOf("commands")(m)["stdout"].(map[string]any), "sha256") },
		"witness unknown field":  func(m map[string]any) { firstOf("witnesses")(m)["extra"] = 1 },
		"artifact unknown field": func(m map[string]any) { firstOf("artifacts")(m)["extra"] = 1 },
	}
	for name, mutate := range rejections {
		t.Run(name, func(t *testing.T) {
			if _, err := ParseCaseManifest(remarshal(t, base, mutate)); err == nil {
				t.Fatal("malformed nested entry accepted by Parse")
			}
		})
	}
}

// remarshal decodes base JSON into a generic map, applies mutate, and re-canonicalizes it.
func remarshal(t *testing.T, base []byte, mutate func(map[string]any)) []byte {
	t.Helper()
	var object map[string]any
	if err := json.Unmarshal(base, &object); err != nil {
		t.Fatal(err)
	}
	mutate(object)
	raw, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := core.Canonicalize(raw)
	if err != nil {
		// A mutation that produces non-canonicalizable JSON still exercises rejection.
		return raw
	}
	return canonical
}
