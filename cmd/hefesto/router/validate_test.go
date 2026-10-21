package router

import (
	"slices"
	"testing"
)

// completePlan returns a plan that satisfies every rule: every promised
// artifact is present, one scenario is covered by one verification.
func completePlan(t *testing.T) Plan {
	t.Helper()

	payload, err := Route(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskNone}, "add-widget")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	plan := payload.Plan
	for i := range plan.Artifacts {
		plan.Artifacts[i].Present = true
	}

	spec, ok := plan.Artifact("spec")
	if !ok {
		t.Fatal("the plan does not track a spec artifact")
	}
	spec.Scenarios = []Scenario{{ID: "S1", Title: "the widget is created"}}
	plan.Artifacts[indexOf(plan, "spec")] = spec

	report, ok := plan.Artifact("verify-report")
	if !ok {
		t.Fatal("the plan does not track a verify-report artifact")
	}
	report.Verifications = []Verification{{ID: "V1", Covers: []string{"S1"}, Command: "go test ./..."}}
	plan.Artifacts[indexOf(plan, "verify-report")] = report

	return plan
}

// indexOf returns the position of an artifact in a plan.
func indexOf(plan Plan, name string) int {
	for i, artifact := range plan.Artifacts {
		if artifact.Name == name {
			return i
		}
	}
	return -1
}

// setArtifact replaces one artifact of a plan.
func setArtifact(plan *Plan, name string, mutate func(*PlannedArtifact)) {
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name == name {
			mutate(&plan.Artifacts[i])
			return
		}
	}
}

// removeArtifact drops one artifact from a plan.
func removeArtifact(plan *Plan, name string) {
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name == name {
			plan.Artifacts = append(plan.Artifacts[:i], plan.Artifacts[i+1:]...)
			return
		}
	}
}

// hasProblem reports whether the report contains a problem with the code.
func hasProblem(report Report, code string) bool {
	for _, problem := range report.Problems {
		if problem.Code == code {
			return true
		}
	}
	return false
}

func TestValidateAcceptsCompletePlan(t *testing.T) {
	report := Validate(completePlan(t))

	if !report.OK() {
		t.Fatalf("Validate() reported problems: %+v", report.Problems)
	}
	for _, check := range report.Checks {
		if !check.Passed {
			t.Errorf("check %q failed: %s", check.Name, check.Detail)
		}
	}
}

func TestValidateRejectsUnknownProfile(t *testing.T) {
	plan := completePlan(t)
	plan.Decision.Profile = Profile("invented")

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan with an unknown profile")
	}
	if !hasProblem(report, CheckProfileKnown) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckProfileKnown)
	}
}

func TestValidateRequiresDecisionRecord(t *testing.T) {
	plan := completePlan(t)
	removeArtifact(&plan, "decision-record")

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan without a decision record")
	}
	if !hasProblem(report, CheckDecisionRecorded) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckDecisionRecorded)
	}
}

func TestValidateRequiresDecisionRecordToBePresent(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "decision-record", func(a *PlannedArtifact) { a.Present = false })

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan whose decision was not recorded")
	}
	if !hasProblem(report, CheckDecisionRecorded) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckDecisionRecorded)
	}
}

func TestValidateRejectsUnknownArtifact(t *testing.T) {
	plan := completePlan(t)
	plan.Artifacts = append(plan.Artifacts, PlannedArtifact{Name: "invented-artifact", Present: true})

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted an artifact that is not in the catalog")
	}
	if !hasProblem(report, CheckArtifactKnown) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckArtifactKnown)
	}
}

func TestValidateRequiresEveryPromisedArtifact(t *testing.T) {
	plan := completePlan(t)
	removeArtifact(&plan, "spec")

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan that dropped a promised artifact")
	}
	if !hasProblem(report, CheckRequiredPresent) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckRequiredPresent)
	}
}

func TestValidateRequiresPromisedArtifactToBePresent(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "spec", func(a *PlannedArtifact) { a.Present = false })

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan whose specification is missing")
	}
	if !hasProblem(report, CheckRequiredPresent) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckRequiredPresent)
	}
}

func TestValidateDetectsDependencyOrder(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "spec", func(a *PlannedArtifact) { a.Present = false })

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan that breaks its own ordering")
	}
	if !hasProblem(report, CheckDependencies) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckDependencies)
	}
}

func TestValidateDetectsUnknownDependency(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "tasks", func(a *PlannedArtifact) { a.DependsOn = []string{"ghost"} })

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan that depends on an artifact it does not track")
	}
	if !hasProblem(report, CheckDependencies) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckDependencies)
	}
}

func TestValidateRequiresVerificationForEveryScenario(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "spec", func(a *PlannedArtifact) {
		a.Scenarios = []Scenario{
			{ID: "S1", Title: "the widget is created"},
			{ID: "S2", Title: "the widget is rejected when invalid"},
		}
	})

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a scenario with no verification")
	}
	if !hasProblem(report, CheckScenarioCoverage) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckScenarioCoverage)
	}
}

func TestValidateRejectsVerificationOfUnknownScenario(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "verify-report", func(a *PlannedArtifact) {
		a.Verifications = []Verification{{ID: "V1", Covers: []string{"S9"}, Command: "go test ./..."}}
	})

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a verification of a scenario that does not exist")
	}
	if !hasProblem(report, CheckScenarioCoverage) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckScenarioCoverage)
	}
}

func TestValidateRequiresSomeVerification(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "verify-report", func(a *PlannedArtifact) { a.Verifications = nil })

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan with no verification at all")
	}
	if !hasProblem(report, CheckVerificationGate) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckVerificationGate)
	}
}

func TestValidateEnforcesIndependentVerifier(t *testing.T) {
	plan := completePlan(t)
	plan.Decision.Verification = VerificationIndependent

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted a plan that required an independent verifier without one")
	}
	if !hasProblem(report, CheckVerificationGate) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckVerificationGate)
	}
}

func TestValidateAcceptsIndependentVerifier(t *testing.T) {
	plan := completePlan(t)
	plan.Decision.Verification = VerificationIndependent
	setArtifact(&plan, "verify-report", func(a *PlannedArtifact) {
		a.Verifications = []Verification{{ID: "V1", Covers: []string{"S1"}, Command: "go test ./...", Independent: true}}
	})

	report := Validate(plan)

	if !report.OK() {
		t.Fatalf("Validate() reported problems: %+v", report.Problems)
	}
}

func TestValidateRejectsOrphanArtifacts(t *testing.T) {
	plan := completePlan(t)
	// focused-test is a cataloged artifact, but this decision never promised it.
	plan.Artifacts = append(plan.Artifacts, PlannedArtifact{Name: "focused-test", Present: true})

	report := Validate(plan)

	if report.OK() {
		t.Fatal("Validate() accepted an artifact the decision did not ask for")
	}
	if !hasProblem(report, CheckNoOrphans) {
		t.Errorf("problems = %+v, want a %q problem", report.Problems, CheckNoOrphans)
	}
}

func TestValidateReportsEveryProblemAtOnce(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "decision-record", func(a *PlannedArtifact) { a.Present = false })
	setArtifact(&plan, "spec", func(a *PlannedArtifact) {
		a.Scenarios = []Scenario{{ID: "S1"}, {ID: "S2"}}
	})
	setArtifact(&plan, "verify-report", func(a *PlannedArtifact) { a.Verifications = nil })

	report := Validate(plan)

	codes := make(map[string]bool, len(report.Problems))
	for _, problem := range report.Problems {
		codes[problem.Code] = true
	}
	for _, want := range []string{CheckDecisionRecorded, CheckScenarioCoverage, CheckVerificationGate} {
		if !codes[want] {
			t.Errorf("problems = %+v, want a %q problem", report.Problems, want)
		}
	}
}

func TestValidateChecksEveryRule(t *testing.T) {
	report := Validate(completePlan(t))

	want := []string{
		CheckProfileKnown,
		CheckDecisionRecorded,
		CheckArtifactKnown,
		CheckRequiredPresent,
		CheckDependencies,
		CheckScenarioCoverage,
		CheckVerificationGate,
		CheckNoOrphans,
	}

	got := make([]string, 0, len(report.Checks))
	for _, check := range report.Checks {
		got = append(got, check.Name)
	}

	if !slices.Equal(got, want) {
		t.Errorf("checks = %v, want %v", got, want)
	}
}

func TestValidateErrorsOnlyListsBlockingProblems(t *testing.T) {
	plan := completePlan(t)
	setArtifact(&plan, "spec", func(a *PlannedArtifact) { a.Present = false })

	report := Validate(plan)

	if len(report.Errors()) == 0 {
		t.Fatal("Errors() is empty for an invalid plan")
	}
	for _, problem := range report.Errors() {
		if problem.Severity != SeverityError {
			t.Errorf("Errors() returned a %q problem", problem.Severity)
		}
	}
}
