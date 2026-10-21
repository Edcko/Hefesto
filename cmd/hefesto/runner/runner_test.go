package runner

import (
	"strings"
	"testing"

	"github.com/Edcko/Hefesto/cmd/hefesto/router"
)

// specFirstPlan returns a routed spec-first plan with nothing produced yet.
func specFirstPlan(t *testing.T) router.Plan {
	t.Helper()

	payload, err := router.Route(router.Request{
		Type: router.WorkKnown,
		Size: router.SizeLarge,
		Risk: router.RiskNone,
	}, "add-widget")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}
	return payload.Plan
}

// fillArtifact records the content a contract requires for one artifact.
func fillArtifact(plan *router.Plan, name string, fields map[string]string) {
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name != name {
			continue
		}
		plan.Artifacts[i].Present = true
		plan.Artifacts[i].Fields = fields
		return
	}
}

// setScenarios records the scenarios of an artifact.
func setScenarios(plan *router.Plan, name string, scenarios []router.Scenario) {
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name == name {
			plan.Artifacts[i].Present = true
			plan.Artifacts[i].Scenarios = scenarios
			return
		}
	}
}

// setTasks records the tasks of an artifact.
func setTasks(plan *router.Plan, name string, tasks []router.Task) {
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name == name {
			plan.Artifacts[i].Present = true
			plan.Artifacts[i].Tasks = tasks
			return
		}
	}
}

// setVerifications records the verifications of an artifact.
func setVerifications(plan *router.Plan, name string, verifications []router.Verification) {
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name == name {
			plan.Artifacts[i].Present = true
			plan.Artifacts[i].Verifications = verifications
			return
		}
	}
}

// satisfiedPlan returns a plan where every promised artifact is complete.
func satisfiedPlan(t *testing.T) router.Plan {
	t.Helper()

	plan := specFirstPlan(t)
	fillArtifact(&plan, "decision-record", map[string]string{"profile": "spec-first", "rationale": "because"})
	fillArtifact(&plan, "proposal", map[string]string{
		"intent": "add the widget", "scope": "one module", "approach": "add a route",
		"risks": "none", "rollback": "revert the commit",
	})
	setScenarios(&plan, "spec", []router.Scenario{{ID: "S1", Title: "the widget is created"}})
	setTasks(&plan, "tasks", []router.Task{{ID: "1.1", Title: "add the route", Done: true}})
	setVerifications(&plan, "verify-report", []router.Verification{
		{ID: "V1", Covers: []string{"S1"}, Command: "go test ./..."},
	})
	return plan
}

func TestNextProducesTheFirstMissingArtifact(t *testing.T) {
	step, err := Next(specFirstPlan(t))
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if step.Phase != PhaseProduce {
		t.Errorf("Phase = %q, want %q", step.Phase, PhaseProduce)
	}
	if step.Artifact != "decision-record" {
		t.Errorf("Artifact = %q, want %q", step.Artifact, "decision-record")
	}
	if step.Brief == "" {
		t.Error("the step carries no brief")
	}
}

func TestNextReportsAnIncompleteArtifact(t *testing.T) {
	plan := specFirstPlan(t)
	fillArtifact(&plan, "decision-record", map[string]string{"profile": "spec-first"})

	step, err := Next(plan)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if step.Phase != PhaseComplete {
		t.Errorf("Phase = %q, want %q", step.Phase, PhaseComplete)
	}
	if step.Artifact != "decision-record" {
		t.Errorf("Artifact = %q, want %q", step.Artifact, "decision-record")
	}
	if len(step.Blockers) == 0 {
		t.Error("the step reports no blockers")
	}
}

func TestNextFollowsTheDependencyOrder(t *testing.T) {
	plan := specFirstPlan(t)
	fillArtifact(&plan, "decision-record", map[string]string{"profile": "spec-first", "rationale": "because"})
	fillArtifact(&plan, "proposal", map[string]string{
		"intent": "add the widget", "scope": "one module", "approach": "add a route",
		"risks": "none", "rollback": "revert the commit",
	})

	step, err := Next(plan)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if step.Artifact != "spec" {
		t.Errorf("Artifact = %q, want %q", step.Artifact, "spec")
	}
}

func TestNextNamesTheInputsOfTheStep(t *testing.T) {
	plan := specFirstPlan(t)
	fillArtifact(&plan, "decision-record", map[string]string{"profile": "spec-first", "rationale": "because"})
	fillArtifact(&plan, "proposal", map[string]string{
		"intent": "add the widget", "scope": "one module", "approach": "add a route",
		"risks": "none", "rollback": "revert the commit",
	})

	step, err := Next(plan)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if len(step.Inputs) == 0 {
		t.Fatal("the step names no inputs")
	}
	found := false
	for _, input := range step.Inputs {
		if input == "proposal" {
			found = true
		}
	}
	if !found {
		t.Errorf("Inputs = %v, want them to include the proposal", step.Inputs)
	}
}

func TestNextClosesWhenTheContractIsSatisfied(t *testing.T) {
	step, err := Next(satisfiedPlan(t))
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	if step.Phase != PhaseClose {
		t.Errorf("Phase = %q, want %q", step.Phase, PhaseClose)
	}
	if step.Artifact != "" {
		t.Errorf("Artifact = %q, want empty when the run closes", step.Artifact)
	}
}

func TestNextRejectsAnUnknownProfile(t *testing.T) {
	plan := specFirstPlan(t)
	plan.Decision.Profile = router.Profile("invented")

	if _, err := Next(plan); err == nil {
		t.Fatal("Next() accepted a plan with an unknown profile")
	}
}

func TestNextRejectsAContractWithADanglingDependency(t *testing.T) {
	plan := specFirstPlan(t)
	plan.Decision.Artifacts = append(plan.Decision.Artifacts, "ghost-artifact")

	if _, err := Next(plan); err == nil {
		t.Fatal("Next() accepted a contract that promises an artifact the catalog does not know")
	}
}

func TestNextIsDeterministic(t *testing.T) {
	plan := specFirstPlan(t)

	first, err := Next(plan)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}
	second, err := Next(plan)
	if err != nil {
		t.Fatalf("Next() error = %v", err)
	}

	if first.Artifact != second.Artifact || first.Phase != second.Phase {
		t.Errorf("Next() is not deterministic: %+v vs %+v", first, second)
	}
}

func TestCheckReportsAMissingField(t *testing.T) {
	plan := specFirstPlan(t)
	fillArtifact(&plan, "decision-record", map[string]string{"profile": "spec-first"})

	problems := Check(plan, "decision-record")

	if len(problems) == 0 {
		t.Fatal("Check() reported no problem for an artifact missing a required field")
	}
	if !strings.Contains(strings.Join(problems, " "), "rationale") {
		t.Errorf("problems = %v, want them to name the missing field", problems)
	}
}

func TestCheckReportsMissingScenarios(t *testing.T) {
	plan := specFirstPlan(t)
	fillArtifact(&plan, "decision-record", map[string]string{"profile": "spec-first", "rationale": "because"})
	fillArtifact(&plan, "proposal", map[string]string{
		"intent": "add the widget", "scope": "one module", "approach": "add a route",
		"risks": "none", "rollback": "revert the commit",
	})
	fillArtifact(&plan, "spec", map[string]string{})

	problems := Check(plan, "spec")

	if len(problems) == 0 {
		t.Fatal("Check() reported no problem for a specification with no scenarios")
	}
	if !strings.Contains(strings.Join(problems, " "), "scenario") {
		t.Errorf("problems = %v, want them to mention the missing scenarios", problems)
	}
}

func TestCheckReportsUndoneTasks(t *testing.T) {
	plan := satisfiedPlan(t)
	setTasks(&plan, "tasks", []router.Task{
		{ID: "1.1", Title: "add the route", Done: true},
		{ID: "1.2", Title: "add the test", Done: false},
	})

	problems := Check(plan, "verify-report")

	if len(problems) == 0 {
		t.Fatal("Check() reported no problem for a run with an unfinished task")
	}
	if !strings.Contains(strings.Join(problems, " "), "1.2") {
		t.Errorf("problems = %v, want them to name the unfinished task", problems)
	}
}

func TestCheckReportsAnUntrackedArtifact(t *testing.T) {
	plan := specFirstPlan(t)

	problems := Check(plan, "not-in-the-plan")

	if len(problems) == 0 {
		t.Fatal("Check() reported no problem for an artifact the plan does not track")
	}
}

func TestCheckAcceptsACompleteArtifact(t *testing.T) {
	plan := satisfiedPlan(t)

	if problems := Check(plan, "proposal"); len(problems) != 0 {
		t.Errorf("Check() reported %v for a complete artifact", problems)
	}
}

func TestBriefStatesTheCompletionRule(t *testing.T) {
	plan := specFirstPlan(t)

	brief, err := Brief(plan, "spec")
	if err != nil {
		t.Fatalf("Brief() error = %v", err)
	}

	for _, want := range []string{"spec", "scenario", "proposal"} {
		if !strings.Contains(brief, want) {
			t.Errorf("the brief does not mention %q", want)
		}
	}
}

func TestBriefRejectsAnUnknownArtifact(t *testing.T) {
	if _, err := Brief(specFirstPlan(t), "not-in-the-plan"); err == nil {
		t.Fatal("Brief() accepted an artifact that is not in the plan")
	}
}
