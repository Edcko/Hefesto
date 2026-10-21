// Package runner drives a run through the contract of its profile.
//
// The router decides which methodology fits a request. The runner turns that
// decision into a sequence: what to produce next, what to read first, and
// what "done" means for the artifact in hand.
//
// A runner does not write the artifacts. A model or a person does. What the
// runner guarantees is the order and the completion criteria, so "the work is
// done" stops being a claim and becomes a check.
package runner

import (
	"fmt"
	"strings"

	"github.com/Edcko/Hefesto/cmd/hefesto/router"
)

// Phase is the kind of action a step asks for.
type Phase string

const (
	// PhaseProduce means the artifact does not exist yet.
	PhaseProduce Phase = "produce"
	// PhaseComplete means the artifact exists but is missing its content.
	PhaseComplete Phase = "complete"
	// PhaseClose means everything the contract requires is present.
	PhaseClose Phase = "close"
)

// Step is the next action of a run.
type Step struct {
	// Phase is the kind of action this step asks for.
	Phase Phase `json:"phase"`
	// Artifact is the artifact the action is about, empty when the run closes.
	Artifact string `json:"artifact"`
	// Kind is the kind of that artifact.
	Kind router.ArtifactKind `json:"kind"`
	// Purpose states why that artifact exists.
	Purpose string `json:"purpose"`
	// Brief is the instruction for whoever produces the artifact.
	Brief string `json:"brief"`
	// Inputs are the artifacts that come before it in the order.
	Inputs []string `json:"inputs"`
	// Blockers are the reasons the step cannot advance.
	Blockers []string `json:"blockers"`
}

// Next returns the next action the run must take.
//
// It walks the artifacts the decision promised, in the order their
// dependencies allow, and stops at the first one that is not complete.
func Next(plan router.Plan) (Step, error) {
	if _, ok := router.Contract(plan.Decision.Profile); !ok {
		return Step{}, fmt.Errorf("runner: the profile %q has no contract", plan.Decision.Profile)
	}

	order, err := router.OrderArtifacts(plan.Decision.Artifacts)
	if err != nil {
		return Step{}, err
	}

	for index, name := range order {
		artifact, ok := plan.Artifact(name)
		if !ok {
			return Step{}, fmt.Errorf("runner: the decision promises %q, which the plan does not track", name)
		}

		problems := Check(plan, name)
		if len(problems) == 0 {
			continue
		}

		spec, _ := router.Artifact(name)
		brief, briefErr := Brief(plan, name)
		if briefErr != nil {
			return Step{}, briefErr
		}

		phase := PhaseComplete
		if !artifact.Present {
			phase = PhaseProduce
		}

		return Step{
			Phase:    phase,
			Artifact: name,
			Kind:     spec.Kind,
			Purpose:  spec.Purpose,
			Brief:    brief,
			Inputs:   order[:index],
			Blockers: problems,
		}, nil
	}

	return Step{Phase: PhaseClose}, nil
}

// Check returns what is missing for one artifact to be complete. An empty
// result means the artifact satisfies its contract.
func Check(plan router.Plan, name string) []string {
	artifact, ok := plan.Artifact(name)
	if !ok {
		return []string{"the plan does not track " + name}
	}

	var problems []string

	if !artifact.Present {
		problems = append(problems, "the artifact is not present")
	}

	spec, ok := router.Artifact(name)
	if !ok {
		return problems
	}

	for _, field := range spec.RequiredFields {
		if strings.TrimSpace(artifact.Fields[field]) == "" {
			problems = append(problems, "the required field "+field+" is empty")
		}
	}

	if spec.RequiresScenarios {
		if len(artifact.Scenarios) == 0 {
			problems = append(problems, "no scenario is recorded")
		}
		for _, scenario := range artifact.Scenarios {
			if scenario.ID == "" || scenario.Title == "" {
				problems = append(problems, "a scenario has no id or no title")
			}
		}
	}

	if spec.RequiresTasks {
		if len(artifact.Tasks) == 0 {
			problems = append(problems, "no task is recorded")
		}
		for _, task := range artifact.Tasks {
			if task.ID == "" || task.Title == "" {
				problems = append(problems, "a task has no id or no title")
			}
		}
	}

	if spec.RequiresVerifications && len(artifact.Verifications) == 0 {
		problems = append(problems, "no verification is recorded")
	}

	if spec.AllTasksDone {
		for _, task := range plan.Tasks() {
			if !task.Done {
				problems = append(problems, "the task "+task.ID+" is not done")
			}
		}
	}

	return problems
}

// Brief returns the instruction for whoever produces one artifact: what to
// read, what to write, and what "done" means.
func Brief(plan router.Plan, name string) (string, error) {
	if _, ok := plan.Artifact(name); !ok {
		return "", fmt.Errorf("runner: the plan does not track %q", name)
	}

	spec, ok := router.Artifact(name)
	if !ok {
		return "", fmt.Errorf("runner: unknown artifact %q", name)
	}

	order, err := router.OrderArtifacts(plan.Decision.Artifacts)
	if err != nil {
		return "", err
	}

	var out strings.Builder

	fmt.Fprintf(&out, "## `%s` (%s)\n\n", name, spec.Kind)
	fmt.Fprintf(&out, "%s\n\n", spec.Purpose)

	out.WriteString("### Read first\n\n")
	read := false
	for _, candidate := range order {
		if candidate == name {
			break
		}
		read = true
		other, _ := router.Artifact(candidate)
		state := "not produced yet"
		if artifact, ok := plan.Artifact(candidate); ok && artifact.Present {
			state = "produced"
		}
		fmt.Fprintf(&out, "- `%s` (%s) — %s\n", candidate, state, other.Purpose)
	}
	if !read {
		out.WriteString("- nothing: this is the first artifact of the run\n")
	}
	out.WriteString("\n")

	out.WriteString("### Produce\n\n")
	switch {
	case spec.RequiresScenarios:
		out.WriteString("Write the requirements and the scenarios that prove them. Every scenario gets an id and a title.\n\n")
	case spec.RequiresTasks:
		out.WriteString("Break the specification down into tasks. Every task gets an id, a title, and it starts not done.\n\n")
	case spec.RequiresVerifications:
		out.WriteString("Record the evidence: one verification per scenario, with the command that produced it.\n\n")
	case spec.Kind == router.ArtifactProbe:
		out.WriteString("Run the probe and record what it observed.\n\n")
	case spec.Kind == router.ArtifactTest:
		out.WriteString("Write the test and record what it proves.\n\n")
	default:
		out.WriteString("Write the artifact and record the fields below.\n\n")
	}

	out.WriteString("### Record in plan.json\n\n")
	fmt.Fprintf(&out, "- `%s.present` → true\n", name)
	for _, field := range spec.RequiredFields {
		fmt.Fprintf(&out, "- `%s.fields.%s`\n", name, field)
	}
	if spec.RequiresScenarios {
		fmt.Fprintf(&out, "- `%s.scenarios` → at least one, each with an id and a title\n", name)
	}
	if spec.RequiresTasks {
		fmt.Fprintf(&out, "- `%s.tasks` → at least one, each with an id and a title\n", name)
	}
	if spec.RequiresVerifications {
		fmt.Fprintf(&out, "- `%s.verifications` → at least one, each covering a scenario\n", name)
	}
	out.WriteString("\n")

	out.WriteString("### Done when\n\n")
	out.WriteString("- `hefesto next` moves past this artifact\n")

	return out.String(), nil
}
