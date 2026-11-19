package reviews

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// R-INV-13 (R2 M3 SPEC draft 5 §7 row 11, as amended by drafts 6-8): the hold, the refusal and
// duplicate-identity, proved with a scripted LedgerView.

type admitAll struct{}

func (admitAll) Ledgered(context.Context, string, string) (bool, error) { return false, nil }
func (admitAll) CommitClaimed(context.Context, string, string, string, string) (bool, error) {
	return true, nil
}
func (admitAll) RequirementLedgered(context.Context, core.Source, string, string) (bool, error) {
	return true, nil
}

type admitAllPtr struct{}

func (*admitAllPtr) Ledgered(context.Context, string, string) (bool, error) { return false, nil }
func (*admitAllPtr) CommitClaimed(context.Context, string, string, string, string) (bool, error) {
	return true, nil
}
func (*admitAllPtr) RequirementLedgered(context.Context, core.Source, string, string) (bool, error) {
	return true, nil
}

// scriptedLedger answers from explicit sets; anything absent is not ledgered.
type scriptedLedger struct {
	identities map[string]bool // schema \x00 native id
	claims     map[string]bool // sha \x00 provider \x00 id \x00 digest
	reqs       map[string]bool // source \x00 id \x00 digest
	err        error
	calls      int
}

func (l *scriptedLedger) Ledgered(_ context.Context, schema, id string) (bool, error) {
	l.calls++
	return l.identities[schema+"\x00"+id], l.err
}
func (l *scriptedLedger) CommitClaimed(_ context.Context, sha, provider, id, digest string) (bool, error) {
	l.calls++
	return l.claims[sha+"\x00"+provider+"\x00"+id+"\x00"+digest], l.err
}
func (l *scriptedLedger) RequirementLedgered(_ context.Context, source core.Source, id, digest string) (bool, error) {
	l.calls++
	return l.reqs[string(source)+"\x00"+id+"\x00"+digest], l.err
}

const (
	v2Schema = "proofbound.obligation-verdict.v2"
	rrSchema = "proofbound.requirement-review.v1"
)

// fixtureLedger admits the fixtures' claim and requirement; callers delete what a case lacks.
func fixtureLedger(t *testing.T) *scriptedLedger {
	t.Helper()
	v := parsedV2(t, obligationArtifact(t, func(*ObligationVerdict) {}))
	l := &scriptedLedger{identities: map[string]bool{}, claims: map[string]bool{}, reqs: map[string]bool{}}
	l.claims[claimKey(v)] = true
	for _, r := range v.Requirements {
		l.reqs[r.Source+"\x00"+r.RecordID+"\x00"+r.ArtifactSHA256] = true
	}
	return l
}

func claimKey(v ObligationVerdict) string {
	return v.ReviewedCommit + "\x00" + strings.TrimPrefix(v.ChangeIntent.Source, "intent.") + "\x00" + v.ChangeIntent.RecordID + "\x00" + v.ChangeIntent.ArtifactSHA256
}

func parsedV2(t *testing.T, a Artifact) ObligationVerdict {
	t.Helper()
	v, err := ParseObligationVerdict(a.Path, a.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// v2At / rrAt build a valid artifact at path, after mutate.
func v2At(t *testing.T, path string, mutate func(*ObligationVerdict)) Artifact {
	t.Helper()
	a := obligationArtifact(t, func(v *ObligationVerdict) {
		v.ArtifactPath = path
		mutate(v)
	})
	a.Path = path
	return a
}

func rrAt(t *testing.T, path string, mutate func(*RequirementReview)) Artifact {
	t.Helper()
	a := requirementReviewArtifact(t, func(v *RequirementReview) {
		v.ArtifactPath = path
		mutate(v)
	})
	a.Path = path
	return a
}

func noop[T any](*T) {}

// withSecondRequirement adds a second, never-ledgered requirement (sorted after the fixture's) with
// one outcome, so a check of only the first requirement cannot pass.
func withSecondRequirement(v *ObligationVerdict) {
	second := v.Requirements[0]
	second.RecordID, second.ArtifactSHA256 = "BR-zeta-item-abc123", strings.Repeat("a", 64)
	v.Requirements = append(v.Requirements, second)
	outcome := v.Obligations[0]
	outcome.RequirementID, outcome.ArtifactSHA256 = second.RecordID, second.ArtifactSHA256
	v.Obligations = append(v.Obligations, outcome)
}

func syncWith(t *testing.T, l LedgerView, artifacts ...Artifact) (Result, *appender, error) {
	t.Helper()
	c, err := New(&Deps{Reader: &reader{artifacts: artifacts}, IDs: ids(t), Logger: connector(t, &reader{}).logger, Ledger: l})
	if err != nil {
		t.Fatal(err)
	}
	app := &appender{}
	res, err := c.Sync(context.Background(), app)
	return res, app, err
}

func appendedIDs(app *appender) []string {
	var out []string
	for _, e := range app.events {
		out = append(out, e.NativeID)
	}
	return out
}

func cursorPaths(t *testing.T, res Result) []string {
	t.Helper()
	var paths []string
	if err := json.Unmarshal(res.Cursor, &paths); err != nil {
		t.Fatal(err)
	}
	return paths
}

const vdir = "docs/verification/verdicts/"

// First revisions wait: each missing prerequisite holds the artifact with its cause, out of the
// appender and the cursor; the sync succeeds; a later artifact still appends; a later sync appends
// the held artifact once the view admits it.
func TestHoldFirstRevisionsUntilPrerequisitesAreLedgered(t *testing.T) {
	base := parsedV2(t, obligationArtifact(t, func(*ObligationVerdict) {}))
	reqKey := base.Requirements[0].Source + "\x00" + base.Requirements[0].RecordID + "\x00" + base.Requirements[0].ArtifactSHA256
	cases := []struct {
		name, cause string
		strip       func(*scriptedLedger)
		artifact    Artifact
	}{
		{"v2 claim", "claim-not-ledgered", func(l *scriptedLedger) { delete(l.claims, claimKey(base)) }, v2At(t, vdir+"a-v2.md", noop[ObligationVerdict])},
		{"v2 requirement", "requirement-not-ledgered", func(l *scriptedLedger) { delete(l.reqs, reqKey) }, v2At(t, vdir+"a-v2.md", noop[ObligationVerdict])},
		{"review requirement", "requirement-not-ledgered", func(l *scriptedLedger) { delete(l.reqs, reqKey) }, rrAt(t, vdir+"a-rr.md", noop[RequirementReview])},
		{"v2 second requirement", "requirement-not-ledgered", func(*scriptedLedger) {}, v2At(t, vdir+"a-v2.md", withSecondRequirement)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			l := fixtureLedger(t)
			c.strip(l)
			later := artifact(vdir+"z-v1.md", "ACCEPTABLE")
			res, app, err := syncWith(t, l, c.artifact, later)
			if err != nil {
				t.Fatalf("a hold failed the sync: %v", err)
			}
			if len(res.Held) != 1 || res.Held[0] != (Hold{Path: c.artifact.Path, Cause: c.cause}) || len(res.Refused) != 0 {
				t.Fatalf("held=%+v refused=%+v", res.Held, res.Refused)
			}
			if len(app.events) != 1 || res.Appended != 1 || res.Malformed != 0 {
				t.Fatalf("appended %v (result %+v)", appendedIDs(app), res)
			}
			for _, p := range cursorPaths(t, res) {
				if p == c.artifact.Path {
					t.Fatal("a held path is in the cursor")
				}
			}
			res, app, err = syncWith(t, admitAll{}, c.artifact, later)
			if err != nil || len(res.Held) != 0 || len(app.events) != 2 {
				t.Fatalf("once admitted: err=%v held=%+v appended=%v", err, res.Held, appendedIDs(app))
			}
		})
	}
}

// Later revisions never wait: a revision (its schema+id ledgered) with a missing prerequisite is
// refused, the remaining artifacts are still processed, and the sync then fails naming it. With its
// prerequisites present a revision appends. A v2 sharing a ledgered v1's id is a FIRST revision.
func TestRefuseLaterRevisionsLoudly(t *testing.T) {
	base := parsedV2(t, obligationArtifact(t, func(*ObligationVerdict) {}))
	revision := v2At(t, vdir+"a-v2.md", noop[ObligationVerdict])
	later := artifact(vdir+"z-v1.md", "ACCEPTABLE")

	l := fixtureLedger(t)
	delete(l.claims, claimKey(base))
	l.identities[v2Schema+"\x00"+base.VerdictID] = true
	res, app, err := syncWith(t, l, revision, later)
	if !errors.Is(err, ErrRevisionRefused) || !strings.Contains(err.Error(), "claim-not-ledgered "+revision.Path) {
		t.Fatalf("err=%v", err)
	}
	if len(res.Refused) != 1 || res.Refused[0] != (Hold{Path: revision.Path, Cause: "claim-not-ledgered"}) || len(res.Held) != 0 {
		t.Fatalf("refused=%+v held=%+v", res.Refused, res.Held)
	}
	if len(app.events) != 1 || app.events[0].NativeID == base.VerdictID {
		t.Fatalf("the remaining artifact was not processed, or the refused one appended: %v", appendedIDs(app))
	}

	ok := fixtureLedger(t)
	ok.identities[v2Schema+"\x00"+base.VerdictID] = true
	if res, app, err := syncWith(t, ok, revision); err != nil || len(app.events) != 1 || len(res.Refused) != 0 {
		t.Fatalf("a revision with its prerequisites present: err=%v appended=%v", err, appendedIDs(app))
	}

	v1Only := fixtureLedger(t)
	delete(v1Only.claims, claimKey(base))
	v1Only.identities["vera.verdict.v1\x00"+base.VerdictID] = true
	res, _, err = syncWith(t, v1Only, revision)
	if err != nil || len(res.Held) != 1 || len(res.Refused) != 0 {
		t.Fatalf("a v2 sharing a v1 id: err=%v held=%+v refused=%+v", err, res.Held, res.Refused)
	}
}

// duplicate-identity: two listed artifacts with one (schema, id) are both refused, whatever the path
// order and whichever of them has its prerequisites.
func TestDuplicateIdentitiesAreRefusedInEitherOrder(t *testing.T) {
	base := parsedV2(t, obligationArtifact(t, func(*ObligationVerdict) {}))
	withClaim := func(path string) Artifact { return v2At(t, path, noop[ObligationVerdict]) }
	withoutClaim := func(path string) Artifact {
		return v2At(t, path, func(v *ObligationVerdict) { v.ReviewedCommit = strings.Repeat("e", 40) })
	}
	for name, pair := range map[string][2]Artifact{
		"claimless first": {withoutClaim(vdir + "a-v2.md"), withClaim(vdir + "b-v2.md")},
		"claimless last":  {withClaim(vdir + "a-v2.md"), withoutClaim(vdir + "b-v2.md")},
	} {
		t.Run(name, func(t *testing.T) {
			res, app, err := syncWith(t, fixtureLedger(t), pair[0], pair[1])
			if !errors.Is(err, ErrRevisionRefused) || len(app.events) != 0 {
				t.Fatalf("err=%v appended=%v", err, appendedIDs(app))
			}
			if len(res.Refused) != 2 || res.Refused[0].Cause != "duplicate-identity" || res.Refused[1].Cause != "duplicate-identity" {
				t.Fatalf("refused=%+v", res.Refused)
			}
		})
	}
	_ = base
	res, _, err := syncWith(t, fixtureLedger(t), rrAt(t, vdir+"a-rr.md", noop[RequirementReview]), rrAt(t, vdir+"b-rr.md", noop[RequirementReview]))
	if !errors.Is(err, ErrRevisionRefused) || len(res.Refused) != 2 {
		t.Fatalf("duplicate reviews: err=%v refused=%+v", err, res.Refused)
	}
}

// A malformed artifact after a refusal still stops the sync at once (R-INV-5); the returned error
// joins both, and Result keeps the refusal.
func TestRefusalJoinsALaterMalformedError(t *testing.T) {
	base := parsedV2(t, obligationArtifact(t, func(*ObligationVerdict) {}))
	l := fixtureLedger(t)
	delete(l.claims, claimKey(base))
	l.identities[v2Schema+"\x00"+base.VerdictID] = true
	bad := Artifact{Path: vdir + "z-bad.md", Bytes: []byte("---\nschema: vera.verdict.v1\nnot yaml\n")}
	res, _, err := syncWith(t, l, v2At(t, vdir+"a-v2.md", noop[ObligationVerdict]), bad)
	if !errors.Is(err, ErrRevisionRefused) || !strings.Contains(err.Error(), "z-bad.md") || res.Malformed != 1 || len(res.Refused) != 1 {
		t.Fatalf("err=%v result=%+v", err, res)
	}
}

// v1 never consults the view; a view error fails the sync; an artifact with every prerequisite
// present appends exactly as before (neutral control).
func TestHoldConsultsTheViewOnlyForV2AndReviews(t *testing.T) {
	broken := &scriptedLedger{err: errors.New("ledger on fire")}
	if _, app, err := syncWith(t, broken, artifact(vdir+"v1.md", "ACCEPTABLE")); err != nil || len(app.events) != 1 || broken.calls != 0 {
		t.Fatalf("v1 with a broken view: err=%v calls=%d", err, broken.calls)
	}
	if _, _, err := syncWith(t, broken, v2At(t, vdir+"a-v2.md", noop[ObligationVerdict])); err == nil || errors.Is(err, ErrRevisionRefused) {
		t.Fatalf("a view error was not a sync failure: %v", err)
	}
	res, app, err := syncWith(t, fixtureLedger(t), v2At(t, vdir+"a-v2.md", noop[ObligationVerdict]), rrAt(t, vdir+"a-rr.md", noop[RequirementReview]))
	if err != nil || len(app.events) != 2 || len(res.Held)+len(res.Refused) != 0 {
		t.Fatalf("neutral control: err=%v appended=%v", err, appendedIDs(app))
	}
}
