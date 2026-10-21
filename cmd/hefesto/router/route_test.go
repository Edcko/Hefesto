package router

import "testing"

func TestRouteAssemblesThePayload(t *testing.T) {
	payload, err := Route(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskContract}, "add-widget")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if payload.Change != "add-widget" {
		t.Errorf("Change = %q, want %q", payload.Change, "add-widget")
	}
	if payload.Decision.Profile != ProfileSpecFirst {
		t.Errorf("Profile = %q, want %q", payload.Decision.Profile, ProfileSpecFirst)
	}
	if payload.Contract.Profile != payload.Decision.Profile {
		t.Errorf("the contract describes %q, want %q", payload.Contract.Profile, payload.Decision.Profile)
	}
	if payload.Scaffold == "" {
		t.Error("the payload carries no scaffold")
	}
	if len(payload.Plan.Artifacts) == 0 {
		t.Error("the payload carries no plan")
	}
}

func TestRoutePropagatesTheDecisionError(t *testing.T) {
	payload, err := Route(Request{Type: "guess", Size: SizeSmall, Risk: RiskNone}, "add-widget")
	if err == nil {
		t.Fatalf("Route() = %+v, want an error", payload)
	}
	if payload.Decision.Profile != "" {
		t.Errorf("Profile = %q, want empty on error", payload.Decision.Profile)
	}
}

func TestRoutePlanStartsUnsatisfiedAndEndsSatisfied(t *testing.T) {
	payload, err := Route(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskNone}, "add-widget")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	start := Validate(payload.Plan)
	if start.OK() {
		t.Fatal("a fresh plan satisfied the contract, want it to start with work outstanding")
	}
	if !hasProblem(start, CheckRequiredPresent) {
		t.Errorf("problems = %+v, want a %q problem", start.Problems, CheckRequiredPresent)
	}

	plan := payload.Plan
	for i := range plan.Artifacts {
		plan.Artifacts[i].Present = true
	}
	setArtifact(&plan, "spec", func(a *PlannedArtifact) {
		a.Scenarios = []Scenario{{ID: "S1", Title: "the widget is created"}}
	})
	setArtifact(&plan, "verify-report", func(a *PlannedArtifact) {
		a.Verifications = []Verification{{ID: "V1", Covers: []string{"S1"}, Command: "go test ./..."}}
	})

	end := Validate(plan)
	if !end.OK() {
		t.Fatalf("the completed plan failed validation: %+v", end.Problems)
	}
}

func TestRoutePlanCarriesTheDecision(t *testing.T) {
	payload, err := Route(Request{Type: WorkAmbiguous, Size: SizeSmall, Risk: RiskNone}, "add-widget")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if payload.Plan.Decision.Profile != payload.Decision.Profile {
		t.Errorf("the plan carries profile %q, want %q", payload.Plan.Decision.Profile, payload.Decision.Profile)
	}
	if payload.Plan.Decision.Rationale != payload.Decision.Rationale {
		t.Errorf("the plan carries a different rationale than the decision")
	}
}
