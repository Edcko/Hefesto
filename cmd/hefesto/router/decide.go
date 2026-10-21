// Package router decides which development methodology profile fits a
// request, so an orchestrator can select a discipline instead of asking the
// human to name a method every time.
//
// The router is a pure function over three observed facts: the dominant
// uncertainty of the work, its size, and the risk item it touches. It never
// reads the repository, never delegates, and never asks a model to improvise a
// method. Its output is a decision with an explicit rationale, the artifacts
// the chosen profile owes, and the verification tier it requires.
//
// The decision is only half the contract. A plan is what a run produces, and
// Validate checks that plan against the artifacts the decision promised, so a
// profile cannot drift away from what the work actually produced.
package router

import "fmt"

// WorkType describes the dominant uncertainty of a request.
type WorkType string

const (
	// WorkKnown means the expected behavior is already understood.
	WorkKnown WorkType = "known"
	// WorkAmbiguous means the expected behavior is not pinned down yet.
	WorkAmbiguous WorkType = "ambiguous_behavior"
	// WorkUnknownDomain means the affected area is not understood yet.
	WorkUnknownDomain WorkType = "unknown_domain"
	// WorkUnvalidatedIdea means the outcome itself is an untested hypothesis.
	WorkUnvalidatedIdea WorkType = "unvalidated_idea"
)

// Size is the size classification of the work.
type Size string

const (
	// SizeSmall is work whose outcome, location and risk are already known
	// and that a single diff would finish.
	SizeSmall Size = "small"
	// SizeLarge is work that needs a tracked artifact to stay resumable.
	SizeLarge Size = "large"
)

// RiskItem names a high-risk area a change touches. The values mirror the
// orchestrator's high-risk list: irreversible data effects, security,
// contracts other consumers rely on, concurrency, delivery and environment,
// and changes no test would catch.
type RiskItem string

const (
	// RiskNone means the change touches no high-risk area.
	RiskNone RiskItem = "none"
	// RiskData means the change writes, rewrites or deletes stored data.
	RiskData RiskItem = "data"
	// RiskSecurity means the change touches auth, permissions or secrets.
	RiskSecurity RiskItem = "security"
	// RiskContract means the change alters a contract others already consume.
	RiskContract RiskItem = "contract"
	// RiskConcurrency means the change affects concurrent behavior.
	RiskConcurrency RiskItem = "concurrency"
	// RiskDelivery means the change touches installers, release, CI or dependencies.
	RiskDelivery RiskItem = "delivery"
	// RiskUntestable means no test would catch a regression of the change.
	RiskUntestable RiskItem = "untestable"
)

// Profile is a methodology profile: a named way of working that an
// orchestrator can hand to a session.
type Profile string

const (
	// ProfileInlineLight keeps the work in the parent session: a focused
	// change, a test, and a short report.
	ProfileInlineLight Profile = "inline-light"
	// ProfileExploreFirst maps the affected area before any decision.
	ProfileExploreFirst Profile = "explore-first"
	// ProfilePrototypeFirst builds a throwaway artifact to validate a
	// hypothesis before production code exists.
	ProfilePrototypeFirst Profile = "prototype-first"
	// ProfileSpecFirst writes the requirements and the testable scenarios
	// down before implementation starts.
	ProfileSpecFirst Profile = "spec-first"
)

// VerificationTier is the verification tier a decision requires.
type VerificationTier string

const (
	// VerificationSelf is the writer's own verification of its work.
	VerificationSelf VerificationTier = "writer-self"
	// VerificationIndependent adds a verifier that did not write the change.
	VerificationIndependent VerificationTier = "independent-verifier"
)

// Request is the observed input to a routing decision.
type Request struct {
	// Type is the dominant uncertainty of the work.
	Type WorkType
	// Size is the size classification of the work.
	Size Size
	// Risk is the high-risk item the change touches, or RiskNone.
	Risk RiskItem
}

// Decision is the router's output for one request.
type Decision struct {
	// Profile is the methodology profile the work should follow.
	Profile Profile `json:"profile"`
	// Verification is the verification tier the profile requires.
	Verification VerificationTier `json:"verification"`
	// FollowUp is the profile to run after a discovery profile, or empty when
	// the chosen profile is terminal for this request.
	FollowUp Profile `json:"follow_up"`
	// Rationale explains why this profile was chosen, in one sentence.
	Rationale string `json:"rationale"`
	// Artifacts are the artifacts the chosen profile owes.
	Artifacts []string `json:"artifacts"`
	// RiskProbes are the artifacts that evidence the risk item, if any.
	RiskProbes []string `json:"risk_probes"`
}

// rollbackPlanRequired reports whether a risk item demands an explicit way
// back before the change lands.
func rollbackPlanRequired(risk RiskItem) bool {
	switch risk {
	case RiskData, RiskContract, RiskDelivery, RiskUntestable:
		return true
	default:
		return false
	}
}

// Decide returns the methodology profile that fits req.
//
// It fails closed: an unknown work type, size or risk item is an error rather
// than a guess, because a wrong guess silently routes the work into the wrong
// discipline.
func Decide(req Request) (Decision, error) {
	if err := validate(req); err != nil {
		return Decision{}, err
	}

	decision := baseDecision(req)

	if req.Risk != RiskNone {
		probe, ok := riskProbe[req.Risk]
		if !ok {
			return Decision{}, fmt.Errorf("router: no probe for risk item %q", req.Risk)
		}
		decision.Verification = VerificationIndependent
		decision.RiskProbes = []string{probe}
		decision.Artifacts = append(decision.Artifacts, probe)
		if rollbackPlanRequired(req.Risk) {
			decision.Artifacts = append(decision.Artifacts, "rollback-plan")
		}
		decision.Rationale += fmt.Sprintf(
			" The change touches the %s risk item, so verification stays independent of the writer.",
			req.Risk,
		)
	}

	return decision, nil
}

// baseDecision maps the work type and size to a profile, before the risk
// item is layered on top.
func baseDecision(req Request) Decision {
	artifacts := append([]string{"decision-record"}, artifactsForProfile(req)...)

	switch req.Type {
	case WorkUnvalidatedIdea:
		return Decision{
			Profile:      ProfilePrototypeFirst,
			Verification: VerificationSelf,
			FollowUp:     ProfileSpecFirst,
			Rationale:    "The outcome is an unvalidated hypothesis, so a prototype and recorded feedback come before any production code.",
			Artifacts:    artifacts,
		}

	case WorkUnknownDomain:
		return Decision{
			Profile:      ProfileExploreFirst,
			Verification: VerificationSelf,
			FollowUp:     ProfileSpecFirst,
			Rationale:    "The affected area is not understood yet, so the work starts with an exploration map and an explicit list of open decisions.",
			Artifacts:    artifacts,
		}

	case WorkAmbiguous:
		if req.Size == SizeLarge {
			return Decision{
				Profile:      ProfileSpecFirst,
				Verification: VerificationSelf,
				Rationale:    "The expected behavior is ambiguous and the work is large, so the requirements are written down before implementation starts.",
				Artifacts:    artifacts,
			}
		}
		return Decision{
			Profile:      ProfileInlineLight,
			Verification: VerificationSelf,
			Rationale:    "The work is small, so it stays inline after one focused clarification of the ambiguous point.",
			Artifacts:    artifacts,
		}

	case WorkKnown:
		if req.Size == SizeLarge {
			return Decision{
				Profile:      ProfileSpecFirst,
				Verification: VerificationSelf,
				Rationale:    "The work is large and understood, so it is tracked as a specification with testable scenarios instead of a loose plan.",
				Artifacts:    artifacts,
			}
		}
		return Decision{
			Profile:      ProfileInlineLight,
			Verification: VerificationSelf,
			Rationale:    "The work is small, understood and low risk, so it stays inline with one focused test.",
			Artifacts:    artifacts,
		}
	}

	// Unreachable: validate rejects every other work type.
	return Decision{}
}

// validate rejects inputs the decision table does not cover.
func validate(req Request) error {
	switch req.Type {
	case WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea:
	default:
		return fmt.Errorf("router: unknown work type %q", req.Type)
	}

	switch req.Size {
	case SizeSmall, SizeLarge:
	default:
		return fmt.Errorf("router: unknown size %q", req.Size)
	}

	switch req.Risk {
	case RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable:
	default:
		return fmt.Errorf("router: unknown risk item %q", req.Risk)
	}

	return nil
}
