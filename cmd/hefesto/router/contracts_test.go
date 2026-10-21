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
