package sdd

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/kamisrini/proofbound/kernel/internal/connector/intent"
	"github.com/kamisrini/proofbound/kernel/internal/core"
)

// buildRecord builds one requirement revision (SPEC §4.2-§4.3), or sets *skip to its cause. An
// error is a provider defect (a candidate failing ValidateRevision for an unlisted reason).
func (p *Provider) buildRecord(key string, spec, prof Artifact, pr profile, found []entry, active bool, skip *string, rev **intent.Revision, obligations *[]string) error {
	seen := map[string]bool{}
	var obls []intent.Obligation
	for _, e := range found {
		id, err := obligationID(e.Sec, e.Name)
		if err != nil {
			*skip = "empty-name-part"
			return nil
		}
		if seen[id] {
			*skip = "obligation-collision"
			return nil
		}
		seen[id] = true
		obls = append(obls, intent.Obligation{ID: id, Statement: e.Statement, State: "active"})
	}
	sort.Slice(obls, func(i, j int) bool { return obls[i].ID < obls[j].ID })
	owner, ok := firstMatch(pr.Owners, spec.Path)
	if !ok {
		*skip = "owner-uncovered"
		return nil
	}
	digest := manifest([]Artifact{spec, prof})
	id, err := requirementID(key, digest)
	if err != nil {
		*skip = "slug-too-short"
		return nil
	}
	status := "active"
	if active {
		status = "proposed"
	}
	payload, err := canonicalPayload(intent.Requirement{Schema: "proofbound.requirement.v1", RequirementID: id, Status: status, AuthorizedBy: []intent.Reference{}, DeclaredOwner: owner, Obligations: obls, Supersedes: []intent.Reference{}, ArtifactPath: spec.Path, ArtifactSHA256: digest})
	if err != nil {
		return fmt.Errorf("sdd: requirement %s: %w", spec.Path, err)
	}
	if len(payload) > core.MaxPayloadBytes {
		*skip = "payload-too-large"
		return nil
	}
	r := p.build(core.KindRequirement, id, spec.Path, digest, payload)
	if err := intent.ValidateRevision(r); err != nil {
		return fmt.Errorf("sdd: provider defect: requirement %s: %w", spec.Path, err)
	}
	*rev = &r
	for _, o := range obls {
		*obligations = append(*obligations, o.ID)
	}
	return nil
}

// buildChange builds the change intent of a name with exactly one change entry and only emitted
// record units (SPEC §5 rules 7-8). It returns a cause instead when rule 7 applies.
func (p *Provider) buildChange(n string, entry Artifact, active bool, files []Artifact, prof Artifact, pr profile, recs []*record) (*intent.Revision, string, error) {
	sponsor, ok := firstMatch(pr.Sponsors, entry.Path)
	if !ok {
		return nil, "sponsor-uncovered", nil
	}
	var targets []intent.Reference
	for _, r := range recs {
		ids := append([]string(nil), r.obligations...)
		sort.Strings(ids)
		targets = append(targets, intent.Reference{RecordKind: "requirement", Source: string(core.SourceIntentSDD), RecordID: r.rev.NativeID, ArtifactSHA256: r.rev.ArtifactSHA256, Relation: "implements", ObligationIDs: ids})
	}
	sort.Slice(targets, func(i, j int) bool {
		a, b := targets[i], targets[j]
		return a.Source+"\x00"+a.RecordKind+"\x00"+a.RecordID+"\x00"+a.ArtifactSHA256+"\x00"+a.Relation < b.Source+"\x00"+b.RecordKind+"\x00"+b.RecordID+"\x00"+b.ArtifactSHA256+"\x00"+b.Relation
	})
	status := "accepted"
	if active {
		status = pr.ChangeStatus
	}
	digest := manifest(append(append([]Artifact(nil), files...), prof))
	id, err := changeIntentID(n)
	if err != nil {
		return nil, "slug-too-short", nil
	}
	payload, err := canonicalPayload(intent.ChangeIntent{Schema: "proofbound.change-intent.v1", IntentID: id, Status: status, DeclaredSponsor: sponsor, Targets: targets, ConstrainedBy: []intent.Reference{}, Supersedes: []intent.Reference{}, ArtifactPath: entry.Path, ArtifactSHA256: digest})
	if err != nil {
		return nil, "", fmt.Errorf("sdd: change %s: %w", entry.Path, err)
	}
	if len(payload) > core.MaxPayloadBytes {
		return nil, "payload-too-large", nil
	}
	r := p.build(core.KindChangeIntent, id, entry.Path, digest, payload)
	if err := intent.ValidateRevision(r); err != nil {
		return nil, "", fmt.Errorf("sdd: provider defect: change %s: %w", entry.Path, err)
	}
	return &r, "", nil
}

// Revisions maps every distinct SDD state on the git connector's walk (SPEC §3) and returns the
// union of their revisions; Skipped then reports each unit skip once (SPEC §7).
func (p *Provider) Revisions(ctx context.Context) ([]intent.Revision, error) {
	if p == nil || p.reader == nil {
		return nil, errors.New("sdd: provider is not initialized")
	}
	commits, err := p.reader.Commits(ctx)
	if err != nil {
		return nil, fmt.Errorf("sdd: %w", err)
	}
	seenState := map[string]bool{}
	seenRev := map[string]bool{}
	seenSkip := map[unitSkip]bool{}
	var revs []intent.Revision
	var skips []intent.Skip
	for _, c := range commits {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		key, err := p.reader.StateKey(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("sdd: state key %s: %w", c, err)
		}
		if key == "" || seenState[key] {
			continue
		}
		seenState[key] = true
		arts, err := p.reader.Read(ctx, c)
		if err != nil {
			return nil, fmt.Errorf("sdd: read %s: %w", c, err)
		}
		st, err := p.mapState(arts)
		if err != nil {
			return nil, err
		}
		for _, r := range st.revisions {
			k := string(r.Kind) + "\x00" + r.NativeID + "\x00" + r.ArtifactSHA256
			if !seenRev[k] {
				seenRev[k] = true
				revs = append(revs, r)
			}
		}
		for _, s := range st.skips {
			if !seenSkip[s] {
				seenSkip[s] = true
				skips = append(skips, intent.Skip{Commit: c, Unit: s.unit, Cause: s.cause, Err: unmappable(s.cause)})
			}
		}
	}
	p.skipped = skips
	return revs, nil
}

// Resolve maps a native change name at a commit to its outcome (SPEC §6). It reads only the
// commit's state and its parents' archive listings (SDD-INV-33).
func (p *Provider) Resolve(ctx context.Context, commit, name string) (intent.Revision, error) {
	if !inDomain(name) {
		return intent.Revision{}, ErrUnknownChange
	}
	arts, err := p.reader.Read(ctx, commit)
	if err != nil {
		return intent.Revision{}, fmt.Errorf("sdd: read %s: %w", commit, err)
	}
	if len(arts) == 0 {
		return intent.Revision{}, ErrUnknownChange
	}
	st, err := p.mapState(arts)
	if err != nil {
		return intent.Revision{}, err
	}
	if st.stateSkip != "" {
		return intent.Revision{}, unmappable(st.stateSkip)
	}
	cr, ok := st.changes[name]
	if name == st.archiveBase || !ok {
		return intent.Revision{}, ErrUnknownChange
	}
	if cr.skip != "" {
		return intent.Revision{}, unmappable(cr.skip)
	}
	if cr.singleArchive {
		parents, err := p.reader.Parents(ctx, commit)
		if err != nil {
			return intent.Revision{}, fmt.Errorf("sdd: parents of %s: %w", commit, err)
		}
		for _, parent := range parents {
			entries, err := p.reader.List(ctx, parent, st.archive)
			if err != nil {
				return intent.Revision{}, fmt.Errorf("sdd: archive of %s: %w", parent, err)
			}
			for _, e := range entries {
				if m := datedRE.FindStringSubmatch(e.Name); m != nil && m[1] == name && isChangeEntryMode(e.Mode) {
					return intent.Revision{}, ErrArchivedEarlier
				}
			}
		}
	}
	if cr.absent {
		return intent.Revision{}, ErrNoChangeIntent
	}
	return *cr.rev, nil
}
