package router

import "fmt"

// ArtifactKind classifies what an artifact is.
type ArtifactKind string

const (
	// ArtifactDecision is a recorded decision.
	ArtifactDecision ArtifactKind = "decision"
	// ArtifactDocument is a written artifact such as a plan, spec or note.
	ArtifactDocument ArtifactKind = "document"
	// ArtifactTest is a test that proves behavior.
	ArtifactTest ArtifactKind = "test"
	// ArtifactProbe is a check that produces evidence about a risk item.
	ArtifactProbe ArtifactKind = "probe"
	// ArtifactReport is the evidence a run produces at its end.
	ArtifactReport ArtifactKind = "report"
)

// ArtifactSpec describes one artifact in the catalog.
type ArtifactSpec struct {
	// Name is the identifier used in decisions and plans.
	Name string
	// Kind is what the artifact is.
	Kind ArtifactKind
	// Purpose states why the artifact exists.
	Purpose string
	// RequiredFields are the content keys the artifact must carry to be
	// complete.
	RequiredFields []string
	// RequiresScenarios reports whether the artifact must state scenarios.
	RequiresScenarios bool
	// RequiresTasks reports whether the artifact must list tasks.
	RequiresTasks bool
	// AllTasksDone reports whether every task of the plan must be finished
	// for the artifact to be complete.
	AllTasksDone bool
	// RequiresVerifications reports whether the artifact must record
	// verifications.
	RequiresVerifications bool
}

// catalog holds every artifact name a decision or a plan may use. A name that
// is not here does not exist: that is what keeps a profile from promising an
// artifact no phase produces.
var catalog = map[string]ArtifactSpec{
	"decision-record": {
		Name:           "decision-record",
		Kind:           ArtifactDecision,
		Purpose:        "Records the routing decision, its rationale and the risk item it named.",
		RequiredFields: []string{"profile", "rationale"},
	},
	"clarification-note": {
		Name:           "clarification-note",
		Kind:           ArtifactDocument,
		Purpose:        "The one focused question that resolved the ambiguous point.",
		RequiredFields: []string{"question", "answer"},
	},
	"one-red-test": {
		Name:           "one-red-test",
		Kind:           ArtifactTest,
		Purpose:        "A test written before the change, observed failing.",
		RequiredFields: []string{"path", "observation"},
	},
	"focused-test": {
		Name:           "focused-test",
		Kind:           ArtifactTest,
		Purpose:        "A test that proves the change behaves as intended.",
		RequiredFields: []string{"path"},
	},
	"exploration-map": {
		Name:           "exploration-map",
		Kind:           ArtifactDocument,
		Purpose:        "A map of the affected area and its current behavior.",
		RequiredFields: []string{"area", "current-behavior"},
	},
	"options-table": {
		Name:           "options-table",
		Kind:           ArtifactDocument,
		Purpose:        "The approaches considered, with their tradeoffs and effort.",
		RequiredFields: []string{"options", "recommended"},
	},
	"open-decisions": {
		Name:           "open-decisions",
		Kind:           ArtifactDocument,
		Purpose:        "The decisions that need the human before the work continues.",
		RequiredFields: []string{"decisions"},
	},
	"prototype": {
		Name:           "prototype",
		Kind:           ArtifactDocument,
		Purpose:        "A throwaway artifact that tests the hypothesis.",
		RequiredFields: []string{"hypothesis", "shape"},
	},
	"feedback-log": {
		Name:           "feedback-log",
		Kind:           ArtifactDocument,
		Purpose:        "The feedback the prototype produced.",
		RequiredFields: []string{"feedback", "source"},
	},
	"proposal": {
		Name:           "proposal",
		Kind:           ArtifactDocument,
		Purpose:        "Intent, scope, approach, risks and rollback plan for the change.",
		RequiredFields: []string{"intent", "scope", "approach", "risks", "rollback"},
	},
	"spec": {
		Name:              "spec",
		Kind:              ArtifactDocument,
		Purpose:           "The requirements and their testable scenarios.",
		RequiresScenarios: true,
	},
	"tasks": {
		Name:          "tasks",
		Kind:          ArtifactDocument,
		Purpose:       "The implementation breakdown derived from the specification.",
		RequiresTasks: true,
	},
	"verify-report": {
		Name:                  "verify-report",
		Kind:                  ArtifactReport,
		Purpose:               "The evidence that the implementation matches the specification.",
		RequiresVerifications: true,
		AllTasksDone:          true,
	},
	"rollback-plan": {
		Name:           "rollback-plan",
		Kind:           ArtifactDocument,
		Purpose:        "How the change is reverted if it lands wrong.",
		RequiredFields: []string{"steps", "trigger"},
	},
	"dry-run-probe": {
		Name:           "dry-run-probe",
		Kind:           ArtifactProbe,
		Purpose:        "Probe that shows an irreversible data effect before it happens.",
		RequiredFields: []string{"method", "observation"},
	},
	"permission-diff": {
		Name:           "permission-diff",
		Kind:           ArtifactProbe,
		Purpose:        "Probe that shows the permissions, guards and secrets a change touches.",
		RequiredFields: []string{"method", "observation"},
	},
	"compat-matrix": {
		Name:           "compat-matrix",
		Kind:           ArtifactProbe,
		Purpose:        "Probe that shows the consumers of a contract a change alters.",
		RequiredFields: []string{"method", "observation"},
	},
	"race-probe": {
		Name:           "race-probe",
		Kind:           ArtifactProbe,
		Purpose:        "Probe that exercises concurrent behavior.",
		RequiredFields: []string{"method", "observation"},
	},
	"staged-install-probe": {
		Name:           "staged-install-probe",
		Kind:           ArtifactProbe,
		Purpose:        "Probe that exercises installers, release, CI and dependencies.",
		RequiredFields: []string{"method", "observation"},
	},
	"manual-acceptance-probe": {
		Name:           "manual-acceptance-probe",
		Kind:           ArtifactProbe,
		Purpose:        "Manual check for a change no test would catch.",
		RequiredFields: []string{"method", "observation"},
	},
}

// Artifact returns the specification of a cataloged artifact.
func Artifact(name string) (ArtifactSpec, bool) {
	spec, ok := catalog[name]
	return spec, ok
}

// ArtifactNames returns every cataloged artifact name, sorted.
func ArtifactNames() []string {
	names := make([]string, 0, len(catalog))
	for name := range catalog {
		names = append(names, name)
	}
	sortStrings(names)
	return names
}

// riskProbe maps a risk item to the probe artifact that evidences it.
var riskProbe = map[RiskItem]string{
	RiskData:        "dry-run-probe",
	RiskSecurity:    "permission-diff",
	RiskContract:    "compat-matrix",
	RiskConcurrency: "race-probe",
	RiskDelivery:    "staged-install-probe",
	RiskUntestable:  "manual-acceptance-probe",
}

// ProfileContract is the verifiable contract of a methodology profile: what
// the profile is for, when it applies, and what a run must satisfy before it
// is finished.
type ProfileContract struct {
	// Profile is the profile this contract describes.
	Profile Profile `json:"profile"`
	// Purpose states what the profile is for.
	Purpose string `json:"purpose"`
	// Entry lists the conditions that justify the profile.
	Entry []string `json:"entry"`
	// Exit lists the conditions that close the profile.
	Exit []string `json:"exit"`
	// Gate states the verification the profile demands.
	Gate string `json:"gate"`
}

// profileContracts holds the contract of every profile.
var profileContracts = map[Profile]ProfileContract{
	ProfileInlineLight: {
		Profile: ProfileInlineLight,
		Purpose: "Keep a small, understood change in one session, with one test as its evidence.",
		Entry: []string{
			"The outcome and the files to touch are known.",
			"No high-risk item is named.",
		},
		Exit: []string{
			"The focused test passes.",
			"The change and its evidence are reported.",
		},
		Gate: "The writer runs the focused check and reports the observed result.",
	},
	ProfileExploreFirst: {
		Profile: ProfileExploreFirst,
		Purpose: "Understand the affected area before choosing an approach.",
		Entry: []string{
			"The affected area or its current behavior is not understood yet.",
		},
		Exit: []string{
			"An exploration map of the affected area exists.",
			"The open decisions are listed for the human.",
			"The follow-up profile is chosen and recorded.",
		},
		Gate: "The writer reports the map, the compared approaches and every decision that still needs the human.",
	},
	ProfilePrototypeFirst: {
		Profile: ProfilePrototypeFirst,
		Purpose: "Validate the hypothesis before any production code exists.",
		Entry: []string{
			"The outcome itself is an untested hypothesis.",
		},
		Exit: []string{
			"A prototype exists.",
			"The feedback it produced is recorded.",
			"The follow-up profile is chosen and recorded.",
		},
		Gate: "The writer reports the prototype, the feedback and the decision that follows from it.",
	},
	ProfileSpecFirst: {
		Profile: ProfileSpecFirst,
		Purpose: "Write the requirements and their testable scenarios down before implementing.",
		Entry: []string{
			"The work is large, or the expected behavior is ambiguous.",
			"The change can be tracked as one artifact set.",
		},
		Exit: []string{
			"Every scenario in the specification has a passing verification.",
			"The verification report records the observed evidence.",
		},
		Gate: "The writer runs the required checks and reports the observed result; an independent verifier re-runs them when the risk item is named.",
	},
}

// Contract returns the contract of a profile.
func Contract(profile Profile) (ProfileContract, bool) {
	contract, ok := profileContracts[profile]
	return contract, ok
}

// Profiles returns every profile, sorted.
func Profiles() []Profile {
	profiles := make([]Profile, 0, len(profileContracts))
	for profile := range profileContracts {
		profiles = append(profiles, profile)
	}
	sortProfiles(profiles)
	return profiles
}

// artifactsForProfile returns the artifacts a profile owes for a request,
// before the risk item adds its own.
func artifactsForProfile(req Request) []string {
	switch req.Type {
	case WorkUnvalidatedIdea:
		return []string{"prototype", "feedback-log"}
	case WorkUnknownDomain:
		return []string{"exploration-map", "options-table", "open-decisions"}
	case WorkAmbiguous:
		if req.Size == SizeLarge {
			return []string{"proposal", "spec", "tasks", "verify-report"}
		}
		return []string{"clarification-note", "one-red-test"}
	case WorkKnown:
		if req.Size == SizeLarge {
			return []string{"proposal", "spec", "tasks", "verify-report"}
		}
		return []string{"focused-test"}
	}
	return nil
}

// Dependencies returns the artifacts that must exist before the given one.
func Dependencies(name string) []string {
	dependencies := artifactDependencies[name]
	out := make([]string, len(dependencies))
	copy(out, dependencies)
	return out
}

// OrderArtifacts returns the given artifacts in the order they must be
// produced: an artifact never comes before the artifacts it depends on.
//
// It fails closed: an artifact the catalog does not know, a dependency the
// decision does not promise, and a dependency cycle are all errors, because
// each of them means the contract and the plan disagree.
func OrderArtifacts(names []string) ([]string, error) {
	return orderArtifacts(names, Dependencies)
}

// orderArtifacts sorts the artifacts with the given dependency lookup.
func orderArtifacts(names []string, dependencies func(string) []string) ([]string, error) {
	known := make(map[string]bool, len(names))
	for _, name := range names {
		if _, ok := Artifact(name); !ok {
			return nil, fmt.Errorf("router: unknown artifact %q", name)
		}
		known[name] = true
	}

	for _, name := range names {
		for _, dependency := range dependencies(name) {
			if !known[dependency] {
				return nil, fmt.Errorf("router: the contract requires %q, which the decision does not promise", dependency)
			}
		}
	}

	emitted := make(map[string]bool, len(names))
	ordered := make([]string, 0, len(names))

	for len(ordered) < len(names) {
		progress := false
		for _, name := range names {
			if emitted[name] {
				continue
			}
			ready := true
			for _, dependency := range dependencies(name) {
				if !emitted[dependency] {
					ready = false
					break
				}
			}
			if !ready {
				continue
			}
			emitted[name] = true
			ordered = append(ordered, name)
			progress = true
		}
		if !progress {
			return nil, fmt.Errorf("router: the artifact dependencies form a cycle")
		}
	}

	return ordered, nil
}

// sortStrings sorts a slice of strings in place.
func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}

// sortProfiles sorts a slice of profiles in place.
func sortProfiles(values []Profile) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
