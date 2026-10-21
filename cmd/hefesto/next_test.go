package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Edcko/Hefesto/cmd/hefesto/router"
	"github.com/Edcko/Hefesto/cmd/hefesto/runner"
)

// writePlanFileForTest writes a plan document and returns its path.
func writePlanFileForTest(t *testing.T, plan router.Plan) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "plan.json")
	if err := writePlanFile(path, plan); err != nil {
		t.Fatalf("writePlanFile() error = %v", err)
	}
	return path
}

// routedPlan returns a fresh routed plan.
func routedPlan(t *testing.T) router.Plan {
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

func TestBuildNextStepReturnsTheFirstArtifact(t *testing.T) {
	path := writePlanFileForTest(t, routedPlan(t))

	step, err := buildNextStep(path)
	if err != nil {
		t.Fatalf("buildNextStep() error = %v", err)
	}
	if step.Phase != runner.PhaseProduce {
		t.Errorf("Phase = %q, want %q", step.Phase, runner.PhaseProduce)
	}
	if step.Artifact != "decision-record" {
		t.Errorf("Artifact = %q, want %q", step.Artifact, "decision-record")
	}
	if step.Brief == "" {
		t.Error("the step carries no brief")
	}
}

func TestBuildNextStepRequiresAPlanPath(t *testing.T) {
	if _, err := buildNextStep(""); err == nil {
		t.Fatal("buildNextStep() accepted an empty plan path")
	}
}

func TestBuildNextStepRejectsAMissingPlan(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.json")
	if _, err := buildNextStep(missing); err == nil {
		t.Fatal("buildNextStep() accepted a missing plan")
	}
}

func TestBuildNextStepRejectsMalformedPlan(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatalf("failed to write plan: %v", err)
	}

	if _, err := buildNextStep(path); err == nil {
		t.Fatal("buildNextStep() accepted a malformed plan")
	}
}

func TestBuildNextStepReportsTheBlockersOfAnIncompleteArtifact(t *testing.T) {
	plan := routedPlan(t)
	for i := range plan.Artifacts {
		if plan.Artifacts[i].Name == "decision-record" {
			plan.Artifacts[i].Present = true
		}
	}
	path := writePlanFileForTest(t, plan)

	step, err := buildNextStep(path)
	if err != nil {
		t.Fatalf("buildNextStep() error = %v", err)
	}
	if step.Phase != runner.PhaseComplete {
		t.Errorf("Phase = %q, want %q", step.Phase, runner.PhaseComplete)
	}
	if len(step.Blockers) == 0 {
		t.Error("the step reports no blockers")
	}
}

func TestBuildNextStepClosesWhenTheContractIsSatisfied(t *testing.T) {
	path := writePlanFileForTest(t, completePlan(t))

	step, err := buildNextStep(path)
	if err != nil {
		t.Fatalf("buildNextStep() error = %v", err)
	}
	if step.Phase != runner.PhaseClose {
		t.Errorf("Phase = %q, want %q", step.Phase, runner.PhaseClose)
	}
}

func TestWritePlanFileWritesAPlanDocument(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.json")
	plan := routedPlan(t)

	if err := writePlanFile(path, plan); err != nil {
		t.Fatalf("writePlanFile() error = %v", err)
	}

	data, err := os.ReadFile(path) //nolint:gosec // G304: the path comes from the test
	if err != nil {
		t.Fatalf("failed to read the plan: %v", err)
	}

	var written router.Plan
	if err := json.Unmarshal(data, &written); err != nil {
		t.Fatalf("the written plan does not parse: %v", err)
	}
	if len(written.Artifacts) != len(plan.Artifacts) {
		t.Errorf("the written plan tracks %d artifacts, want %d", len(written.Artifacts), len(plan.Artifacts))
	}
	if written.Decision.Profile != plan.Decision.Profile {
		t.Errorf("the written plan carries profile %q, want %q", written.Decision.Profile, plan.Decision.Profile)
	}
}

func TestWritePlanFileReportsAFailingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-directory", "plan.json")
	if err := writePlanFile(path, routedPlan(t)); err == nil {
		t.Fatal("writePlanFile() accepted a path it cannot write")
	}
}
