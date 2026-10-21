package router

// Scenario is one testable behavior stated in a specification.
type Scenario struct {
	// ID identifies the scenario inside its plan.
	ID string `json:"id"`
	// Title states the behavior the scenario describes.
	Title string `json:"title"`
}

// Verification is the evidence that a scenario behaves as specified.
type Verification struct {
	// ID identifies the verification inside its plan.
	ID string `json:"id"`
	// Covers lists the scenario IDs this verification proves.
	Covers []string `json:"covers"`
	// Command is what was run.
	Command string `json:"command"`
	// Independent reports whether the verification came from someone other
	// than the writer.
	Independent bool `json:"independent"`
}

// PlannedArtifact is one artifact in a plan and its current state.
type PlannedArtifact struct {
	// Name is the cataloged artifact this entry tracks.
	Name string `json:"name"`
	// Present reports whether the artifact exists yet.
	Present bool `json:"present"`
	// DependsOn lists the artifacts that must exist first.
	DependsOn []string `json:"depends_on"`
	// Scenarios are the testable behaviors this artifact states.
	Scenarios []Scenario `json:"scenarios"`
	// Verifications are the checks that prove the scenarios.
	Verifications []Verification `json:"verifications"`
}

// Plan is the record of one routed run: the decision that opened it and the
// artifacts it produced.
type Plan struct {
	// Change is the name of the change the plan covers.
	Change string `json:"change"`
	// Decision is the routing decision this plan follows.
	Decision Decision `json:"decision"`
	// Artifacts are the artifacts the run produced or still owes.
	Artifacts []PlannedArtifact `json:"artifacts"`
}

// Artifact returns the planned artifact with the given name.
func (p Plan) Artifact(name string) (PlannedArtifact, bool) {
	for _, artifact := range p.Artifacts {
		if artifact.Name == name {
			return artifact, true
		}
	}
	return PlannedArtifact{}, false
}

// Scenarios returns every scenario stated in the plan.
func (p Plan) Scenarios() []Scenario {
	var scenarios []Scenario
	for _, artifact := range p.Artifacts {
		scenarios = append(scenarios, artifact.Scenarios...)
	}
	return scenarios
}

// Verifications returns every verification recorded in the plan.
func (p Plan) Verifications() []Verification {
	var verifications []Verification
	for _, artifact := range p.Artifacts {
		verifications = append(verifications, artifact.Verifications...)
	}
	return verifications
}
