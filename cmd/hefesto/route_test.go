package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Edcko/Hefesto/cmd/hefesto/router"
)

// writePlan writes a plan document to a temporary file and returns its path.
func writePlan(t *testing.T, plan router.Plan) string {
	t.Helper()

	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal plan: %v", err)
	}

	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("failed to write plan: %v", err)
	}
	return path
}

// completePlan returns a plan that satisfies its contract.
func completePlan(t *testing.T) router.Plan {
	t.Helper()

	payload, err := router.Route(router.Request{
		Type: router.WorkKnown,
		Size: router.SizeLarge,
		Risk: router.RiskNone,
	}, "add-widget")
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	plan := payload.Plan
	for i := range plan.Artifacts {
		plan.Artifacts[i].Present = true

		spec, ok := router.Artifact(plan.Artifacts[i].Name)
		if !ok {
			continue
		}
		if len(spec.RequiredFields) > 0 {
			plan.Artifacts[i].Fields = make(map[string]string, len(spec.RequiredFields))
			for _, field := range spec.RequiredFields {
				plan.Artifacts[i].Fields[field] = "recorded for " + plan.Artifacts[i].Name
			}
		}
		if spec.RequiresTasks {
			plan.Artifacts[i].Tasks = []router.Task{{ID: "1.1", Title: "do the work", Done: true}}
		}

		switch plan.Artifacts[i].Name {
		case "spec":
			plan.Artifacts[i].Scenarios = []router.Scenario{{ID: "S1", Title: "the widget is created"}}
		case "verify-report":
			plan.Artifacts[i].Verifications = []router.Verification{
				{ID: "V1", Covers: []string{"S1"}, Command: "go test ./..."},
			}
		}
	}

	return plan
}

func TestBuildRouteRequestRequiresType(t *testing.T) {
	routeType = ""
	routeSize = "small"
	routeRisk = "none"

	if _, err := buildRouteRequest(); err == nil {
		t.Fatal("buildRouteRequest() accepted a request without a type")
	}
}

func TestBuildRouteRequestRequiresSize(t *testing.T) {
	routeType = "known"
	routeSize = ""
	routeRisk = "none"

	if _, err := buildRouteRequest(); err == nil {
		t.Fatal("buildRouteRequest() accepted a request without a size")
	}
}

func TestBuildRouteRequestReturnsTheRequest(t *testing.T) {
	routeType = "ambiguous_behavior"
	routeSize = "large"
	routeRisk = "contract"

	request, err := buildRouteRequest()
	if err != nil {
		t.Fatalf("buildRouteRequest() error = %v", err)
	}
	if request.Type != router.WorkAmbiguous {
		t.Errorf("Type = %q, want %q", request.Type, router.WorkAmbiguous)
	}
	if request.Size != router.SizeLarge {
		t.Errorf("Size = %q, want %q", request.Size, router.SizeLarge)
	}
	if request.Risk != router.RiskContract {
		t.Errorf("Risk = %q, want %q", request.Risk, router.RiskContract)
	}
}

func TestVerifyPlanFileAcceptsACompletePlan(t *testing.T) {
	path := writePlan(t, completePlan(t))

	report, err := verifyPlanFile(path)
	if err != nil {
		t.Fatalf("verifyPlanFile() error = %v", err)
	}
	if !report.OK() {
		t.Fatalf("the report reports problems: %+v", report.Problems)
	}
}

func TestVerifyPlanFileReportsAnIncompletePlan(t *testing.T) {
	plan := completePlan(t)
	plan.Artifacts[0].Present = false
	path := writePlan(t, plan)

	report, err := verifyPlanFile(path)
	if err != nil {
		t.Fatalf("verifyPlanFile() error = %v", err)
	}
	if report.OK() {
		t.Fatal("the report accepted an incomplete plan")
	}
}

func TestVerifyPlanFileRejectsAMissingFile(t *testing.T) {
	if _, err := verifyPlanFile(filepath.Join(t.TempDir(), "absent.json")); err == nil {
		t.Fatal("verifyPlanFile() accepted a missing file")
	}
}

func TestVerifyPlanFileRejectsMalformedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "plan.json")
	if err := os.WriteFile(path, []byte("not json"), 0600); err != nil {
		t.Fatalf("failed to write plan: %v", err)
	}

	if _, err := verifyPlanFile(path); err == nil {
		t.Fatal("verifyPlanFile() accepted a malformed plan")
	}
}

func TestRouteCommandProducesAPayload(t *testing.T) {
	routeType = "known"
	routeSize = "large"
	routeRisk = "contract"
	routeChange = "add-widget"
	routeJSON = true

	payload, err := router.Route(router.Request{
		Type: router.WorkType(routeType),
		Size: router.Size(routeSize),
		Risk: router.RiskItem(routeRisk),
	}, routeChange)
	if err != nil {
		t.Fatalf("Route() error = %v", err)
	}

	if payload.Decision.Profile != router.ProfileSpecFirst {
		t.Errorf("Profile = %q, want %q", payload.Decision.Profile, router.ProfileSpecFirst)
	}
	if payload.Decision.Verification != router.VerificationIndependent {
		t.Errorf("Verification = %q, want %q", payload.Decision.Verification, router.VerificationIndependent)
	}
	if len(payload.Plan.Artifacts) == 0 {
		t.Error("the payload carries no plan")
	}
}
