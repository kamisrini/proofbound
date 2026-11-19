package git

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// These tests derive from the R2 M3 SPEC (docs/plans/R2-sdd-m3-git-binding-spec-draft5.md §7 rows
// 1-4, as amended by drafts 6-8) and prove INV-31 and INV-32.

var sddDigest = strings.Repeat("b", 64)

// scriptedResolver echoes records/specdir claims and answers sdd claims per change name.
type scriptedResolver struct {
	sdd   map[string]IntentRef
	errs  map[string]error
	calls []string
}

func (r *scriptedResolver) Resolve(_ context.Context, sha, provider, id string) (IntentRef, error) {
	r.calls = append(r.calls, sha+"|"+provider+"|"+id)
	if provider != "sdd" {
		return IntentRef{Provider: provider, RecordID: id, ArtifactSHA256: strings.Repeat("a", 64)}, nil
	}
	if err, ok := r.errs[id]; ok {
		return IntentRef{}, err
	}
	if ref, ok := r.sdd[id]; ok {
		return ref, nil
	}
	return IntentRef{}, &ClaimUnboundError{Cause: "unknown-change"}
}

func addAuthResolver() *scriptedResolver {
	return &scriptedResolver{sdd: map[string]IntentRef{
		"add-auth": {Provider: "sdd", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: sddDigest},
	}}
}

func syncCommits(t *testing.T, resolver IntentResolver, commits ...Commit) (Result, map[string][]IntentRef, error) {
	t.Helper()
	app := &memoryAppender{}
	result, err := connectorWithResolver(t, &fakeRepo{commits: commits, tips: map[string]string{}}, resolver).Sync(context.Background(), app)
	refs := map[string][]IntentRef{}
	for _, e := range app.events {
		var payload Commit
		if jerr := json.Unmarshal(e.Payload, &payload); jerr != nil {
			t.Fatal(jerr)
		}
		refs[payload.SHA] = payload.IntentRefs
	}
	return result, refs, err
}

func claimed(sha string, trailers ...string) Commit {
	c := testCommit(sha)
	c.IntentTrailers = trailers
	return c
}

// M3-INV-1 (unit half): a well-formed sdd claim binds the RESOLVED revision, never the trailer value.
func TestSync_SDDClaimBindsResolvedRevision(t *testing.T) {
	resolver := addAuthResolver()
	result, refs, err := syncCommits(t, resolver, claimed("aaa", "sdd:add-auth"), claimed("bbb", "records:CI-sample-item-acde12"))
	if err != nil {
		t.Fatal(err)
	}
	want := IntentRef{Provider: "sdd", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: sddDigest}
	if len(refs["aaa"]) != 1 || refs["aaa"][0] != want {
		t.Fatalf("sdd refs=%+v", refs["aaa"])
	}
	if resolver.calls[0] != "aaa|sdd|add-auth" {
		t.Fatalf("resolver saw %v", resolver.calls)
	}
	if len(refs["bbb"]) != 1 || refs["bbb"][0].Provider != "records" {
		t.Fatalf("positive control: records claim did not bind: %+v", refs["bbb"])
	}
	if len(result.ClaimSkips) != 0 {
		t.Fatalf("skips=%+v", result.ClaimSkips)
	}
}

// M3-INV-2 (connector half, INV-32): a resolver ref that is not a well-formed sdd ref fails the sync.
func TestSync_SDDSpoofedRevisionFailsSync(t *testing.T) {
	for name, ref := range map[string]IntentRef{
		"provider": {Provider: "records", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: sddDigest},
		"id":       {Provider: "sdd", RecordID: "add-auth", ArtifactSHA256: sddDigest},
		"digest":   {Provider: "sdd", RecordID: "CI-add-auth-xuuno2", ArtifactSHA256: "bad"},
	} {
		t.Run(name, func(t *testing.T) {
			resolver := &scriptedResolver{sdd: map[string]IntentRef{"add-auth": ref}}
			if _, _, err := syncCommits(t, resolver, claimed("aaa", "sdd:add-auth")); err == nil {
				t.Fatal("spoofed sdd revision accepted")
			}
		})
	}
	if _, refs, err := syncCommits(t, addAuthResolver(), claimed("aaa", "sdd:add-auth")); err != nil || len(refs["aaa"]) != 1 {
		t.Fatalf("neutral control: a correct revision must bind: %v %+v", err, refs)
	}
}

// M3-INV-3 (INV-31): every tenant outcome is one ClaimSkip, the commit carries NO refs, other
// commits are unaffected, and the sync succeeds.
func TestSync_SDDTenantOutcomesAreClaimSkips(t *testing.T) {
	resolver := addAuthResolver()
	resolver.errs = map[string]error{
		"archived":  &ClaimUnboundError{Cause: "archived-earlier"},
		"no-intent": &ClaimUnboundError{Cause: "no-change-intent"},
		"ab":        &ClaimUnboundError{Cause: "unmappable:slug-too-short"},
		"pending":   &ClaimUnboundError{Cause: "not-ledgered"},
	}
	cases := []struct{ value, cause string }{
		{"sdd:nope", "unknown-change"},
		{"sdd:archived", "archived-earlier"},
		{"sdd:no-intent", "no-change-intent"},
		{"sdd:ab", "unmappable:slug-too-short"},
		{"sdd:pending", "not-ledgered"},
		{"SDD:add-auth", "malformed"},
		{"Sdd:add-auth", "malformed"},
		{"sdd:Add-auth", "malformed"},
		{"sdd:a", "malformed"},
		{"sdd:", "malformed"},
		{"sdd:add_auth", "malformed"},
		{"sdd:add-auth:extra", "malformed"},
		{"sdd:-lead", "malformed"},
		// D1's upper bound: a change name of 100 bytes is well formed (it reaches the resolver),
		// one of 101 is malformed (review r2-m3-impl-review-mutation-concurrency-20261119 F3).
		{"sdd:a" + strings.Repeat("b", 99), "unknown-change"},
		{"sdd:a" + strings.Repeat("b", 100), "malformed"},
	}
	for _, c := range cases {
		t.Run(c.value, func(t *testing.T) {
			result, refs, err := syncCommits(t, resolver,
				claimed("bad", c.value),
				claimed("two", "sdd:add-auth", c.value),
				claimed("mix", "records:CI-sample-item-acde12", c.value),
				claimed("good", "sdd:add-auth"))
			if err != nil {
				t.Fatalf("a tenant outcome failed the sync: %v", err)
			}
			for _, sha := range []string{"bad", "two", "mix"} {
				if refs[sha] != nil {
					t.Errorf("%s kept refs beside a skipped claim: %+v", sha, refs[sha])
				}
			}
			if len(refs["good"]) != 1 {
				t.Errorf("the clean commit lost its ref: %+v", refs["good"])
			}
			if len(result.ClaimSkips) != 3 {
				t.Fatalf("skips=%+v", result.ClaimSkips)
			}
			for i, sha := range []string{"bad", "two", "mix"} {
				if got := result.ClaimSkips[i]; got != (ClaimSkip{Commit: sha, Value: c.value, Cause: c.cause}) {
					t.Errorf("skip %d = %+v", i, got)
				}
			}
		})
	}
	if _, refs, err := syncCommits(t, resolver, claimed("ok", "sdd:add-auth")); err != nil || len(refs["ok"]) != 1 {
		t.Fatalf("control: sdd:add-auth must resolve: %v %+v", err, refs)
	}
}

// M3-INV-4: a resolver failure that is not a ClaimUnboundError fails the sync, as does an unbound
// report without a cause.
func TestSync_SDDResolverErrorsFailSync(t *testing.T) {
	for name, err := range map[string]error{
		"plain":         errors.New("reader on fire"),
		"wrapped plain": errors.Join(errors.New("cancelled"), context.Canceled),
		"no cause":      &ClaimUnboundError{},
	} {
		t.Run(name, func(t *testing.T) {
			resolver := &scriptedResolver{errs: map[string]error{"add-auth": err}}
			if _, _, syncErr := syncCommits(t, resolver, claimed("aaa", "sdd:add-auth")); syncErr == nil {
				t.Fatal("a resolver error was read as a skip")
			}
		})
	}
	if !errors.Is(&ClaimUnboundError{Cause: "x"}, ErrClaimUnbound) {
		t.Fatal("ClaimUnboundError does not match ErrClaimUnbound")
	}
}

// INV-31's detector is exactly "the first four bytes fold to sdd:". A value that differs in any one
// of those bytes, or is shorter than four bytes, is NOT an sdd claim: it keeps the records/specdir
// path, where an unknown provider or a malformed id fails the sync loudly, never a quiet ClaimSkip
// (review r2-m3-impl-review-mutation-concurrency-20261119 F3).
func TestSync_NearSDDValuesAreNotClaims(t *testing.T) {
	for _, value := range []string{"sdd", "sd", "sdx:add-auth", "xdd:add-auth", "sxd:add-auth", "abc:CI-sample-item-acde12", "sdd-x:add-auth", "sdd;add-auth"} {
		t.Run(value, func(t *testing.T) {
			result, _, err := syncCommits(t, addAuthResolver(), claimed("aaa", value))
			if err == nil {
				t.Fatalf("%q did not fail the sync", value)
			}
			if len(result.ClaimSkips) != 0 {
				t.Fatalf("%q became a claim skip: %+v", value, result.ClaimSkips)
			}
		})
	}
}
