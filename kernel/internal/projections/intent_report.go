package projections

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/kamisrini/proofbound/kernel/internal/store"
)

type intentReportRevision struct {
	source, id, digest, status, sponsor, eventID string
	seq                                          int64
}
type intentTargetReport struct {
	requirementSource, requirementID, requirementDigest, obligationID, relation, statement, state, eventID string
	seq                                                                                                    int64
	review, outcome                                                                                        string
}
type intentCommitReport struct {
	sha, eventID     string
	seq              int64
	deployments      []string
	deploymentProofs []string
}

func (p *Projector) ReportIntent(ctx context.Context, s *store.Store, id string, now time.Time, output io.Writer) error {
	if output == nil || strings.TrimSpace(id) == "" {
		return errors.New("intent report: id and output are required")
	}
	if err := p.Apply(ctx, s); err != nil {
		return err
	}
	var revisions []intentReportRevision
	if err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		rows, err := tx.Query(ctx, `SELECT c.source,c.intent_id,c.artifact_sha256,c.status,c.declared_sponsor,c.event_id,c.seq,e.event_id FROM change_intents_view c LEFT JOIN events e ON e.event_id=c.event_id WHERE c.intent_id=$1 ORDER BY c.seq,c.source,c.artifact_sha256`, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var r intentReportRevision
			var proof *string
			if err := rows.Scan(&r.source, &r.id, &r.digest, &r.status, &r.sponsor, &r.eventID, &r.seq, &proof); err != nil {
				return err
			}
			if proof == nil {
				return fmt.Errorf("intent %s: missing event proof %s", id, r.eventID)
			}
			revisions = append(revisions, r)
		}
		return rows.Err()
	}); err != nil {
		return err
	}
	if len(revisions) == 0 {
		return fmt.Errorf("intent report: %s not found", id)
	}
	for _, revision := range revisions {
		targets, commits, err := readIntentComponents(ctx, s, revision)
		if err != nil {
			return err
		}
		state, spec := deriveIntentState(revision, targets, commits)
		if _, err := fmt.Fprintf(output, "intent=%s source=%s artifact_sha256=%s status=%s sponsor=%s state=%s spec=%s proof=%s/%d\n", revision.id, revision.source, revision.digest, revision.status, revision.sponsor, state, spec, revision.eventID, revision.seq); err != nil {
			return err
		}
		for _, target := range targets {
			if _, err := fmt.Fprintf(output, "  target=%s:%s@%s#%s relation=%s obligation_state=%s review=%s outcome=%s statement=%q proof=%s/%d\n", target.requirementSource, target.requirementID, target.requirementDigest, target.obligationID, target.relation, target.state, target.review, target.outcome, target.statement, target.eventID, target.seq); err != nil {
				return err
			}
		}
		for _, commit := range commits {
			deployment := "missing"
			if len(commit.deployments) > 0 {
				deployment = strings.Join(commit.deployments, ",")
			}
			if _, err := fmt.Fprintf(output, "  commit=%s deployment=%s proof=%s/%d", commit.sha, deployment, commit.eventID, commit.seq); err != nil {
				return err
			}
			for _, proof := range commit.deploymentProofs {
				if _, err := fmt.Fprintf(output, " deployment_proof=%s", proof); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintln(output); err != nil {
				return err
			}
		}
	}
	_ = now
	return nil
}

func readIntentComponents(ctx context.Context, s *store.Store, revision intentReportRevision) ([]intentTargetReport, []intentCommitReport, error) {
	var targets []intentTargetReport
	var commits []intentCommitReport
	err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		rows, err := tx.Query(ctx, `SELECT t.requirement_source,t.requirement_id,t.requirement_artifact_sha256,t.obligation_id,t.relation,o.statement,o.state,t.event_id,t.seq,e.event_id,COALESCE((SELECT rr.outcome FROM requirement_reviews_view rr WHERE rr.requirement_source=t.requirement_source AND rr.requirement_id=t.requirement_id AND rr.requirement_artifact_sha256=t.requirement_artifact_sha256 AND rr.obligation_id=t.obligation_id ORDER BY rr.seq DESC LIMIT 1),'UNREVIEWED'),COALESCE((SELECT ov.outcome FROM obligation_verdicts_view ov WHERE ov.intent_source=t.intent_source AND ov.intent_id=t.intent_id AND ov.intent_artifact_sha256=t.intent_artifact_sha256 AND ov.requirement_source=t.requirement_source AND ov.requirement_id=t.requirement_id AND ov.requirement_artifact_sha256=t.requirement_artifact_sha256 AND ov.obligation_id=t.obligation_id ORDER BY ov.seq DESC LIMIT 1),'UNVERIFIED') FROM intent_targets_view t JOIN requirement_obligations_view o ON o.source=t.requirement_source AND o.requirement_id=t.requirement_id AND o.artifact_sha256=t.requirement_artifact_sha256 AND o.obligation_id=t.obligation_id LEFT JOIN events e ON e.event_id=t.event_id WHERE t.intent_source=$1 AND t.intent_id=$2 AND t.intent_artifact_sha256=$3 ORDER BY t.requirement_source,t.requirement_id,t.obligation_id`, revision.source, revision.id, revision.digest)
		if err != nil {
			return err
		}
		for rows.Next() {
			var t intentTargetReport
			var proof *string
			if err := rows.Scan(&t.requirementSource, &t.requirementID, &t.requirementDigest, &t.obligationID, &t.relation, &t.statement, &t.state, &t.eventID, &t.seq, &proof, &t.review, &t.outcome); err != nil {
				rows.Close()
				return err
			}
			if proof == nil {
				rows.Close()
				return fmt.Errorf("intent target %s: missing event proof %s", t.obligationID, t.eventID)
			}
			targets = append(targets, t)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		rows, err = tx.Query(ctx, `SELECT c.commit_sha,c.event_id,c.seq,e.event_id FROM commit_intents_view c LEFT JOIN events e ON e.event_id=c.event_id WHERE c.provider=$1 AND c.intent_id=$2 AND c.intent_artifact_sha256=$3 ORDER BY c.seq,c.commit_sha`, strings.TrimPrefix(revision.source, "intent."), revision.id, revision.digest)
		if err != nil {
			return err
		}
		for rows.Next() {
			var c intentCommitReport
			var proof *string
			if err := rows.Scan(&c.sha, &c.eventID, &c.seq, &proof); err != nil {
				rows.Close()
				return err
			}
			if proof == nil {
				rows.Close()
				return fmt.Errorf("intent commit %s: missing event proof %s", c.sha, c.eventID)
			}
			commits = append(commits, c)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		rows.Close()
		for i := range commits {
			rows, err = tx.Query(ctx, `SELECT g.environment,g.status,g.event_id,g.seq,e.event_id FROM github_delivery_view g LEFT JOIN events e ON e.event_id=g.event_id WHERE g.kind='github.deployment' AND g.commit_sha=$1 ORDER BY g.environment,g.seq`, commits[i].sha)
			if err != nil {
				return err
			}
			for rows.Next() {
				var environment, status, eventID string
				var seq int64
				var proof *string
				if err := rows.Scan(&environment, &status, &eventID, &seq, &proof); err != nil {
					rows.Close()
					return err
				}
				if proof == nil {
					rows.Close()
					return fmt.Errorf("deployment for %s: missing event proof %s", commits[i].sha, eventID)
				}
				commits[i].deployments = append(commits[i].deployments, environment+":"+status)
				commits[i].deploymentProofs = append(commits[i].deploymentProofs, fmt.Sprintf("%s/%d", eventID, seq))
			}
			if err := rows.Err(); err != nil {
				rows.Close()
				return err
			}
			rows.Close()
		}
		return nil
	})
	return targets, commits, err
}

func deriveIntentState(revision intentReportRevision, targets []intentTargetReport, commits []intentCommitReport) (string, string) {
	if revision.status == "superseded" || revision.status == "withdrawn" {
		return "SUPERSEDED", specState(targets)
	}
	if len(commits) == 0 {
		return "DECLARED", specState(targets)
	}
	allSatisfied, anyFailed, anyInconclusive, allReviewed := len(targets) > 0, false, false, true
	for _, target := range targets {
		switch target.outcome {
		case "SATISFIED":
		case "NOT_SATISFIED":
			anyFailed = true
			allSatisfied = false
		case "INCONCLUSIVE":
			anyInconclusive = true
			allSatisfied = false
		default:
			allSatisfied = false
		}
		if target.review != "VERIFIABLE" {
			allReviewed = false
		}
	}
	state := "IMPLEMENTED_UNVERIFIED"
	if anyFailed {
		state = "NOT_SATISFIED"
	} else if anyInconclusive {
		state = "INCONCLUSIVE"
	} else if allSatisfied && allReviewed {
		state = "SATISFIED"
	}
	deployed := false
	for _, commit := range commits {
		deployed = deployed || len(commit.deployments) > 0
	}
	if deployed {
		if state == "SATISFIED" {
			state = "DEPLOYED_VERIFIED"
		} else {
			state = "DEPLOYED_UNVERIFIED"
		}
	}
	return state, specState(targets)
}
func specState(targets []intentTargetReport) string {
	state := "verifiable"
	for _, t := range targets {
		switch t.review {
		case "CONTRADICTORY":
			return "contradictory"
		case "UNTESTABLE":
			if state != "contradictory" {
				state = "untestable"
			}
		case "AMBIGUOUS":
			if state == "verifiable" || state == "unreviewed" {
				state = "ambiguous"
			}
		case "UNREVIEWED":
			if state == "verifiable" {
				state = "unreviewed"
			}
		}
	}
	return state
}

func (p *Projector) ReportRequirement(ctx context.Context, s *store.Store, id string, output io.Writer) error {
	if output == nil || id == "" {
		return errors.New("requirement report: id and output are required")
	}
	if err := p.Apply(ctx, s); err != nil {
		return err
	}
	found := false
	return s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		rows, err := tx.Query(ctx, `SELECT r.source,r.requirement_id,r.artifact_sha256,r.status,r.declared_owner,r.authorization_declared,r.event_id,r.seq,e.event_id FROM requirements_view r LEFT JOIN events e ON e.event_id=r.event_id WHERE r.requirement_id=$1 ORDER BY r.seq`, id)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			found = true
			var source, rid, digest, status, owner, eventID string
			var declared bool
			var seq int64
			var proof *string
			if err := rows.Scan(&source, &rid, &digest, &status, &owner, &declared, &eventID, &seq, &proof); err != nil {
				return err
			}
			if proof == nil {
				return fmt.Errorf("requirement %s: missing event proof %s", rid, eventID)
			}
			authorization := "undeclared"
			if declared {
				authorization = "declared"
			}
			if _, err := fmt.Fprintf(output, "requirement=%s source=%s artifact_sha256=%s status=%s owner=%s authorization=%s proof=%s/%d\n", rid, source, digest, status, owner, authorization, eventID, seq); err != nil {
				return err
			}
		}
		if err := rows.Err(); err != nil {
			return err
		}
		if !found {
			return fmt.Errorf("requirement report: %s not found", id)
		}
		return nil
	})
}

func (p *Projector) CheckIntent(ctx context.Context, s *store.Store, sha string, output io.Writer) error {
	if output == nil || !gitPattern.MatchString(sha) {
		return errors.New("intent check: valid commit and output are required")
	}
	if err := p.Apply(ctx, s); err != nil {
		return err
	}
	count := 0
	err := s.WithTx(ctx, func(ctx context.Context, tx *store.Tx) error {
		rows, err := tx.Query(ctx, `SELECT c.provider,c.intent_id,c.intent_artifact_sha256,c.event_id,c.seq,e.event_id FROM commit_intents_view c LEFT JOIN events e ON e.event_id=c.event_id JOIN change_intents_view i ON i.source='intent.'||c.provider AND i.intent_id=c.intent_id AND i.artifact_sha256=c.intent_artifact_sha256 WHERE c.commit_sha=$1 ORDER BY c.provider,c.intent_id`, sha)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			count++
			var provider, id, digest, eventID string
			var seq int64
			var proof *string
			if err := rows.Scan(&provider, &id, &digest, &eventID, &seq, &proof); err != nil {
				return err
			}
			if proof == nil {
				return fmt.Errorf("intent check: missing event proof %s", eventID)
			}
			if _, err := fmt.Fprintf(output, "commit=%s intent=%s:%s@%s proof=%s/%d\n", sha, provider, id, digest, eventID, seq); err != nil {
				return err
			}
		}
		return rows.Err()
	})
	if err != nil {
		return err
	}
	if count == 0 {
		_, err := fmt.Fprintf(output, "commit=%s intent=none\n", sha)
		return err
	}
	return nil
}

func sortedStrings(values []string) []string { sort.Strings(values); return values }
