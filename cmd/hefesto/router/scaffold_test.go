package router

import (
	"slices"
	"strings"
	"testing"
)

func TestScaffoldRecordsTheDecision(t *testing.T) {
	decision, err := Decide(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskContract})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	contract, ok := Contract(decision.Profile)
	if !ok {
		t.Fatalf("the profile %q has no contract", decision.Profile)
	}

	scaffold := Scaffold("add-widget", decision, contract)

	for _, want := range []string{
		"# add-widget",
		"## Decision",
		"Profile: `spec-first`",
		"Verification: `independent-verifier`",
		decision.Rationale,
		"## Entry criteria",
		"## Artifacts",
		"## Scenarios",
		"## Verification",
		"## Exit criteria",
	} {
		if !strings.Contains(scaffold, want) {
			t.Errorf("scaffold does not contain %q", want)
		}
	}
}

func TestScaffoldListsEveryPromisedArtifact(t *testing.T) {
	decision, err := Decide(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskNone})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	contract, _ := Contract(decision.Profile)

	scaffold := Scaffold("add-widget", decision, contract)

	for _, name := range decision.Artifacts {
		if !strings.Contains(scaffold, "`"+name+"`") {
			t.Errorf("scaffold does not list artifact %q", name)
		}
	}
}

func TestScaffoldNamesTheFollowUpOnlyWhenThereIsOne(t *testing.T) {
	discovery, err := Decide(Request{Type: WorkUnknownDomain, Size: SizeLarge, Risk: RiskNone})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	discoveryContract, _ := Contract(discovery.Profile)
	if !strings.Contains(Scaffold("x", discovery, discoveryContract), "Follow-up: `spec-first`") {
		t.Error("the scaffold of a discovery profile does not name its follow-up")
	}

	terminal, err := Decide(Request{Type: WorkKnown, Size: SizeSmall, Risk: RiskNone})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}
	terminalContract, _ := Contract(terminal.Profile)
	if !strings.Contains(Scaffold("x", terminal, terminalContract), "Follow-up: none") {
		t.Error("the scaffold of a terminal profile does not state that there is no follow-up")
	}
}

func TestPlanTemplateTracksEveryPromisedArtifact(t *testing.T) {
	decision, err := Decide(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskContract})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}

	plan := PlanTemplate("add-widget", decision)

	if plan.Change != "add-widget" {
		t.Errorf("Change = %q, want %q", plan.Change, "add-widget")
	}
	if len(plan.Artifacts) != len(decision.Artifacts) {
		t.Fatalf("the plan tracks %d artifacts, want %d", len(plan.Artifacts), len(decision.Artifacts))
	}
	for i, artifact := range plan.Artifacts {
		if artifact.Name != decision.Artifacts[i] {
			t.Errorf("artifact %d = %q, want %q", i, artifact.Name, decision.Artifacts[i])
		}
		if artifact.Present {
			t.Errorf("artifact %q starts out present, want it missing", artifact.Name)
		}
	}
}

func TestPlanTemplateDependenciesResolveToTrackedArtifacts(t *testing.T) {
	types := []WorkType{WorkKnown, WorkAmbiguous, WorkUnknownDomain, WorkUnvalidatedIdea}
	sizes := []Size{SizeSmall, SizeLarge}
	risks := []RiskItem{RiskNone, RiskData, RiskSecurity, RiskContract, RiskConcurrency, RiskDelivery, RiskUntestable}

	for _, workType := range types {
		for _, size := range sizes {
			for _, risk := range risks {
				decision, err := Decide(Request{Type: workType, Size: size, Risk: risk})
				if err != nil {
					t.Fatalf("Decide() error = %v", err)
				}
				plan := PlanTemplate("change", decision)
				for _, artifact := range plan.Artifacts {
					for _, dependency := range artifact.DependsOn {
						if _, ok := plan.Artifact(dependency); !ok {
							t.Errorf("artifact %q depends on %q, which the plan does not track", artifact.Name, dependency)
						}
					}
				}
			}
		}
	}
}

func TestPlanTemplateRecordsTheOrdering(t *testing.T) {
	decision, err := Decide(Request{Type: WorkKnown, Size: SizeLarge, Risk: RiskNone})
	if err != nil {
		t.Fatalf("Decide() error = %v", err)
	}

	plan := PlanTemplate("add-widget", decision)

	spec, ok := plan.Artifact("spec")
	if !ok {
		t.Fatal("the plan does not track a spec artifact")
	}
	if !slices.Equal(spec.DependsOn, []string{"proposal"}) {
		t.Errorf("the spec depends on %v, want [proposal]", spec.DependsOn)
	}

	tasks, ok := plan.Artifact("tasks")
	if !ok {
		t.Fatal("the plan does not track a tasks artifact")
	}
	if !slices.Equal(tasks.DependsOn, []string{"spec"}) {
		t.Errorf("the tasks depend on %v, want [spec]", tasks.DependsOn)
	}
}
