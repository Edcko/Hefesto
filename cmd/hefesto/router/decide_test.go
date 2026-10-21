package router

import (
	"slices"
	"strings"
	"testing"
)

func TestDecide(t *testing.T) {
	tests := []struct {
		name             string
		req              Request
		wantProfile      Profile
		wantVerification VerificationTier
		wantFollowUp     Profile
		wantArtifacts    []string
		wantProbes       []string
	}{
		{
			name:             "known behavior, small, no risk stays inline",
			req:              Request{Type: WorkKnown, Size: SizeSmall, Risk: RiskNone},
			wantProfile:      ProfileInlineLight,
			wantVerification: VerificationSelf,
			wantArtifacts:    []string{"decision-record", "focused-test"},
		},
		{
			name:             "known behavior, large work needs a spec",
			req:              Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskNone},
			wantProfile:      ProfileSpecFirst,
			wantVerification: VerificationSelf,
			wantArtifacts:    []string{"decision-record", "proposal", "spec", "tasks", "verify-report"},
		},
		{
			name:             "known behavior, small, data risk escalates verification",
			req:              Request{Type: WorkKnown, Size: SizeSmall, Risk: RiskData},
			wantProfile:      ProfileInlineLight,
			wantVerification: VerificationIndependent,
			wantArtifacts:    []string{"decision-record", "focused-test", "dry-run-probe", "rollback-plan"},
			wantProbes:       []string{"dry-run-probe"},
		},
		{
			name:             "known behavior, large, contract risk escalates verification",
			req:              Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskContract},
			wantProfile:      ProfileSpecFirst,
			wantVerification: VerificationIndependent,
			wantArtifacts:    []string{"decision-record", "proposal", "spec", "tasks", "verify-report", "compat-matrix", "rollback-plan"},
			wantProbes:       []string{"compat-matrix"},
		},
		{
			name:             "known behavior, large, security risk escalates verification",
			req:              Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskSecurity},
			wantProfile:      ProfileSpecFirst,
			wantVerification: VerificationIndependent,
			wantArtifacts:    []string{"decision-record", "proposal", "spec", "tasks", "verify-report", "permission-diff"},
			wantProbes:       []string{"permission-diff"},
		},
		{
			name:             "known behavior, large, untestable risk requires a rollback plan",
			req:              Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskUntestable},
			wantProfile:      ProfileSpecFirst,
			wantVerification: VerificationIndependent,
			wantArtifacts:    []string{"decision-record", "proposal", "spec", "tasks", "verify-report", "manual-acceptance-probe", "rollback-plan"},
			wantProbes:       []string{"manual-acceptance-probe"},
		},
		{
			name:             "known behavior, small, delivery risk escalates verification",
			req:              Request{Type: WorkKnown, Size: SizeSmall, Risk: RiskDelivery},
			wantProfile:      ProfileInlineLight,
			wantVerification: VerificationIndependent,
			wantArtifacts:    []string{"decision-record", "focused-test", "staged-install-probe", "rollback-plan"},
			wantProbes:       []string{"staged-install-probe"},
		},
		{
			name:             "ambiguous behavior, small work stays inline after one clarification",
			req:              Request{Type: WorkAmbiguous, Size: SizeSmall, Risk: RiskNone},
			wantProfile:      ProfileInlineLight,
			wantVerification: VerificationSelf,
			wantArtifacts:    []string{"decision-record", "clarification-note", "one-red-test"},
		},
		{
			name:             "ambiguous behavior, large work needs a spec",
			req:              Request{Type: WorkAmbiguous, Size: SizeLarge, Risk: RiskNone},
			wantProfile:      ProfileSpecFirst,
			wantVerification: VerificationSelf,
			wantArtifacts:    []string{"decision-record", "proposal", "spec", "tasks", "verify-report"},
		},
		{
			name:             "ambiguous behavior, small, concurrency risk escalates verification",
			req:              Request{Type: WorkAmbiguous, Size: SizeSmall, Risk: RiskConcurrency},
			wantProfile:      ProfileInlineLight,
			wantVerification: VerificationIndependent,
			wantArtifacts:    []string{"decision-record", "clarification-note", "one-red-test", "race-probe"},
			wantProbes:       []string{"race-probe"},
		},
		{
			name:             "unknown domain, small work explores first",
			req:              Request{Type: WorkUnknownDomain, Size: SizeSmall, Risk: RiskNone},
			wantProfile:      ProfileExploreFirst,
			wantVerification: VerificationSelf,
			wantFollowUp:     ProfileSpecFirst,
			wantArtifacts:    []string{"decision-record", "exploration-map", "options-table", "open-decisions"},
		},
		{
			name:             "unknown domain, large work explores first and keeps the risk",
			req:              Request{Type: WorkUnknownDomain, Size: SizeLarge, Risk: RiskContract},
			wantProfile:      ProfileExploreFirst,
			wantVerification: VerificationIndependent,
			wantFollowUp:     ProfileSpecFirst,
			wantArtifacts:    []string{"decision-record", "exploration-map", "options-table", "open-decisions", "compat-matrix", "rollback-plan"},
			wantProbes:       []string{"compat-matrix"},
		},
		{
			name:             "unvalidated idea prototypes before any production code",
			req:              Request{Type: WorkUnvalidatedIdea, Size: SizeLarge, Risk: RiskNone},
			wantProfile:      ProfilePrototypeFirst,
			wantVerification: VerificationSelf,
			wantFollowUp:     ProfileSpecFirst,
			wantArtifacts:    []string{"decision-record", "prototype", "feedback-log"},
		},
		{
			name:             "unvalidated idea keeps the risk item on the prototype",
			req:              Request{Type: WorkUnvalidatedIdea, Size: SizeSmall, Risk: RiskData},
			wantProfile:      ProfilePrototypeFirst,
			wantVerification: VerificationIndependent,
			wantFollowUp:     ProfileSpecFirst,
			wantArtifacts:    []string{"decision-record", "prototype", "feedback-log", "dry-run-probe", "rollback-plan"},
			wantProbes:       []string{"dry-run-probe"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decide(tt.req)
			if err != nil {
				t.Fatalf("Decide() error = %v", err)
			}
			if got.Profile != tt.wantProfile {
				t.Errorf("Profile = %q, want %q", got.Profile, tt.wantProfile)
			}
			if got.Verification != tt.wantVerification {
				t.Errorf("Verification = %q, want %q", got.Verification, tt.wantVerification)
			}
			if got.FollowUp != tt.wantFollowUp {
				t.Errorf("FollowUp = %q, want %q", got.FollowUp, tt.wantFollowUp)
			}
			if !slices.Equal(got.Artifacts, tt.wantArtifacts) {
				t.Errorf("Artifacts = %v, want %v", got.Artifacts, tt.wantArtifacts)
			}
			if !slices.Equal(got.RiskProbes, tt.wantProbes) {
				t.Errorf("RiskProbes = %v, want %v", got.RiskProbes, tt.wantProbes)
			}
			if got.Rationale == "" {
				t.Error("Rationale is empty, want an explanation of the decision")
			}
		})
	}
}

func TestDecideRejectsUnknownInput(t *testing.T) {
	tests := []struct {
		name string
		req  Request
	}{
		{name: "unknown work type", req: Request{Type: "guess", Size: SizeSmall, Risk: RiskNone}},
		{name: "empty work type", req: Request{Type: "", Size: SizeSmall, Risk: RiskNone}},
		{name: "unknown size", req: Request{Type: WorkKnown, Size: "medium", Risk: RiskNone}},
		{name: "empty size", req: Request{Type: WorkKnown, Size: "", Risk: RiskNone}},
		{name: "unknown risk item", req: Request{Type: WorkKnown, Size: SizeSmall, Risk: "maybe"}},
		{name: "empty risk item", req: Request{Type: WorkKnown, Size: SizeSmall, Risk: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Decide(tt.req)
			if err == nil {
				t.Fatalf("Decide() = %+v, want an error", got)
			}
			if got.Profile != "" {
				t.Errorf("Profile = %q, want empty on error", got.Profile)
			}
		})
	}
}

func TestDecideIsDeterministic(t *testing.T) {
	req := Request{Type: WorkAmbiguous, Size: SizeLarge, Risk: RiskContract}

	first, err := Decide(req)
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	second, err := Decide(req)
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}

	if first.Profile != second.Profile || first.Verification != second.Verification {
		t.Errorf("Decide() is not deterministic: %+v vs %+v", first, second)
	}
	if !slices.Equal(first.Artifacts, second.Artifacts) {
		t.Errorf("Artifacts differ between calls: %v vs %v", first.Artifacts, second.Artifacts)
	}
}

func TestEveryProfileIsReachable(t *testing.T) {
	profiles := []Profile{ProfileInlineLight, ProfileExploreFirst, ProfilePrototypeFirst, ProfileSpecFirst}
	sizes := []Size{SizeSmall, SizeLarge}
	types := []WorkType{WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea}
	risks := []RiskItem{RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable}

	reachable := make(map[Profile]bool, len(profiles))
	for _, workType := range types {
		for _, size := range sizes {
			for _, risk := range risks {
				got, err := Decide(Request{Type: workType, Size: size, Risk: risk})
				if err != nil {
					t.Fatalf("Decide(%v/%v/%v) error = %v", workType, size, risk, err)
				}
				reachable[got.Profile] = true
			}
		}
	}

	for _, profile := range profiles {
		if !reachable[profile] {
			t.Errorf("profile %q is unreachable from any input combination", profile)
		}
	}
}

func TestEveryDecisionCarriesArtifacts(t *testing.T) {
	types := []WorkType{WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea}
	sizes := []Size{SizeSmall, SizeLarge}
	risks := []RiskItem{RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable}

	for _, workType := range types {
		for _, size := range sizes {
			for _, risk := range risks {
				got, err := Decide(Request{Type: workType, Size: size, Risk: risk})
				if err != nil {
					t.Fatalf("Decide() error = %v", err)
				}
				if len(got.Artifacts) == 0 {
					t.Errorf("Decide(%v/%v/%v) returned no artifacts", workType, size, risk)
				}
				if got.Verification != VerificationSelf && got.Verification != VerificationIndependent {
					t.Errorf("Verification = %q, want a known tier", got.Verification)
				}
			}
		}
	}
}

func TestFollowUpOnlyOnDiscoveryProfiles(t *testing.T) {
	types := []WorkType{WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea}
	sizes := []Size{SizeSmall, SizeLarge}
	risks := []RiskItem{RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable}

	for _, workType := range types {
		for _, size := range sizes {
			for _, risk := range risks {
				got, err := Decide(Request{Type: workType, Size: size, Risk: risk})
				if err != nil {
					t.Fatalf("Decide() error = %v", err)
				}
				switch got.Profile {
				case ProfileExploreFirst, ProfilePrototypeFirst:
					if got.FollowUp != ProfileSpecFirst {
						t.Errorf("Decide(%v/%v/%v) FollowUp = %q, want %q", workType, size, risk, got.FollowUp, ProfileSpecFirst)
					}
				default:
					if got.FollowUp != "" {
						t.Errorf("Decide(%v/%v/%v) FollowUp = %q, want empty", workType, size, risk, got.FollowUp)
					}
				}
			}
		}
	}
}

func TestRationaleNamesTheRiskItem(t *testing.T) {
	got, err := Decide(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskSecurity})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	if !strings.Contains(got.Rationale, string(RiskSecurity)) {
		t.Errorf("Rationale = %q, want it to name the risk item %q", got.Rationale, RiskSecurity)
	}
}

func TestEveryDecidedArtifactExistsInTheCatalog(t *testing.T) {
	types := []WorkType{WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea}
	sizes := []Size{SizeSmall, SizeLarge}
	risks := []RiskItem{RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable}

	for _, workType := range types {
		for _, size := range sizes {
			for _, risk := range risks {
				got, err := Decide(Request{Type: workType, Size: size, Risk: risk})
				if err != nil {
					t.Fatalf("Decide() error = %v", err)
				}
				for _, name := range got.Artifacts {
					if _, ok := Artifact(name); !ok {
						t.Errorf("Decide(%v/%v/%v) promises unknown artifact %q", workType, size, risk, name)
					}
				}
			}
		}
	}
}

func TestEveryCatalogArtifactIsReachable(t *testing.T) {
	types := []WorkType{WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea}
	sizes := []Size{SizeSmall, SizeLarge}
	risks := []RiskItem{RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable}

	reachable := make(map[string]bool)
	for _, workType := range types {
		for _, size := range sizes {
			for _, risk := range risks {
				got, err := Decide(Request{Type: workType, Size: size, Risk: risk})
				if err != nil {
					t.Fatalf("Decide() error = %v", err)
				}
				for _, name := range got.Artifacts {
					reachable[name] = true
				}
			}
		}
	}

	for _, name := range ArtifactNames() {
		if !reachable[name] {
			t.Errorf("catalog artifact %q is never promised by a decision", name)
		}
	}
}
