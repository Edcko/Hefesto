package router

import (
	"slices"
	"testing"
)

func TestArtifactResolvesACatalogedName(t *testing.T) {
	spec, ok := Artifact("spec")
	if !ok {
		t.Fatal("Artifact() did not resolve a cataloged artifact")
	}
	if spec.Name != "spec" {
		t.Errorf("Name = %q, want %q", spec.Name, "spec")
	}
	if spec.Kind != ArtifactDocument {
		t.Errorf("Kind = %q, want %q", spec.Kind, ArtifactDocument)
	}
	if spec.Purpose == "" {
		t.Error("the artifact has no purpose")
	}
}

func TestArtifactRejectsAnUnknownName(t *testing.T) {
	if _, ok := Artifact("invented"); ok {
		t.Fatal("Artifact() resolved an artifact that is not in the catalog")
	}
}

func TestContractResolvesEveryProfile(t *testing.T) {
	for _, profile := range Profiles() {
		contract, ok := Contract(profile)
		if !ok {
			t.Errorf("Contract(%q) did not resolve", profile)
			continue
		}
		if contract.Profile != profile {
			t.Errorf("Contract(%q) describes %q", profile, contract.Profile)
		}
		if contract.Purpose == "" {
			t.Errorf("the contract of %q has no purpose", profile)
		}
		if len(contract.Entry) == 0 {
			t.Errorf("the contract of %q has no entry criteria", profile)
		}
		if len(contract.Exit) == 0 {
			t.Errorf("the contract of %q has no exit criteria", profile)
		}
		if contract.Gate == "" {
			t.Errorf("the contract of %q has no verification gate", profile)
		}
	}
}

func TestContractRejectsAnUnknownProfile(t *testing.T) {
	if _, ok := Contract(Profile("invented")); ok {
		t.Fatal("Contract() resolved a profile that does not exist")
	}
}

func TestProfilesAreSorted(t *testing.T) {
	profiles := Profiles()
	if !slices.IsSorted(profiles) {
		t.Errorf("Profiles() = %v, want it sorted", profiles)
	}
}

func TestArtifactNamesAreSorted(t *testing.T) {
	names := ArtifactNames()
	if !slices.IsSorted(names) {
		t.Errorf("ArtifactNames() = %v, want it sorted", names)
	}
}

func TestOrderArtifactsPutsDependenciesFirst(t *testing.T) {
	ordered, err := OrderArtifacts([]string{"spec", "proposal", "tasks"})
	if err != nil {
		t.Fatalf("OrderArtifacts() error = %v", err)
	}
	if !slices.Equal(ordered, []string{"proposal", "spec", "tasks"}) {
		t.Errorf("OrderArtifacts() = %v, want [proposal spec tasks]", ordered)
	}
}

func TestOrderArtifactsRejectsAnUnknownArtifact(t *testing.T) {
	if _, err := OrderArtifacts([]string{"invented"}); err == nil {
		t.Fatal("OrderArtifacts() accepted an artifact the catalog does not know")
	}
}

func TestOrderArtifactsRejectsADanglingDependency(t *testing.T) {
	if _, err := OrderArtifacts([]string{"spec"}); err == nil {
		t.Fatal("OrderArtifacts() accepted an artifact whose dependency the decision does not promise")
	}
}

func TestOrderArtifactsRejectsACycle(t *testing.T) {
	cyclic := func(name string) []string {
		switch name {
		case "proposal":
			return []string{"spec"}
		case "spec":
			return []string{"proposal"}
		default:
			return nil
		}
	}

	if _, err := orderArtifacts([]string{"proposal", "spec"}, cyclic); err == nil {
		t.Fatal("orderArtifacts() accepted a dependency cycle")
	}
}

func TestEveryDecisionCanBeOrdered(t *testing.T) {
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
				ordered, err := OrderArtifacts(decision.Artifacts)
				if err != nil {
					t.Errorf("OrderArtifacts(%v) error = %v", decision.Artifacts, err)
					continue
				}
				if len(ordered) != len(decision.Artifacts) {
					t.Errorf("OrderArtifacts() returned %d artifacts, want %d", len(ordered), len(decision.Artifacts))
				}
			}
		}
	}
}

func TestDependenciesAreCataloged(t *testing.T) {
	for _, name := range ArtifactNames() {
		for _, dependency := range Dependencies(name) {
			if _, ok := Artifact(dependency); !ok {
				t.Errorf("the artifact %q depends on %q, which the catalog does not know", name, dependency)
			}
		}
	}
}
