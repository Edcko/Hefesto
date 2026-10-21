package router

// Severity classifies how serious a finding is.
type Severity string

const (
	// SeverityError means the plan violates the contract and the run is not
	// finished.
	SeverityError Severity = "error"
	// SeverityWarning means the plan deviates without breaking the contract.
	SeverityWarning Severity = "warning"
)

// Problem is one finding of a validation run.
type Problem struct {
	// Code identifies the kind of problem.
	Code string `json:"code"`
	// Severity is how serious the problem is.
	Severity Severity `json:"severity"`
	// Artifact is the artifact the problem belongs to, if any.
	Artifact string `json:"artifact,omitempty"`
	// Detail explains what is wrong.
	Detail string `json:"detail"`
}

// Check is the result of one contract rule.
type Check struct {
	// Name identifies the rule.
	Name string `json:"name"`
	// Passed reports whether the plan satisfies the rule.
	Passed bool `json:"passed"`
	// Detail explains the outcome.
	Detail string `json:"detail"`
}

// Report is the outcome of validating a plan against its contract.
type Report struct {
	// Profile is the profile the plan was validated against.
	Profile Profile `json:"profile"`
	// Checks lists every rule that was evaluated.
	Checks []Check `json:"checks"`
	// Problems lists the findings.
	Problems []Problem `json:"problems"`
}

// OK reports whether the plan satisfies its contract.
func (r Report) OK() bool {
	for _, problem := range r.Problems {
		if problem.Severity == SeverityError {
			return false
		}
	}
	return true
}

// Errors returns only the findings that block the run.
func (r Report) Errors() []Problem {
	var problems []Problem
	for _, problem := range r.Problems {
		if problem.Severity == SeverityError {
			problems = append(problems, problem)
		}
	}
	return problems
}

// Check names of the contract rules the validator evaluates.
const (
	CheckProfileKnown     = "profile-known"
	CheckDecisionRecorded = "decision-recorded"
	CheckArtifactKnown    = "artifact-known"
	CheckRequiredPresent  = "required-artifacts-present"
	CheckDependencies     = "dependency-order"
	CheckScenarioCoverage = "scenario-coverage"
	CheckVerificationGate = "verification-gate"
	CheckNoOrphans        = "no-orphan-artifacts"
)

// Validate checks a plan against the contract its decision promised.
//
// Every rule is evaluated: a plan that satisfies nothing still reports which
// rules failed, so the report is a list of what to fix rather than a single
// verdict.
func Validate(plan Plan) Report {
	report := Report{Profile: plan.Decision.Profile}

	report.add(validateProfileKnown(plan))
	report.add(validateDecisionRecorded(plan))
	report.add(validateArtifactKnown(plan))
	report.add(validateRequiredPresent(plan))
	report.add(validateDependencies(plan))
	report.add(validateScenarioCoverage(plan))
	report.add(validateVerificationGate(plan))
	report.add(validateNoOrphans(plan))

	return report
}

// add appends a check and the problems it found.
func (r *Report) add(check Check, problems []Problem) {
	r.Checks = append(r.Checks, check)
	r.Problems = append(r.Problems, problems...)
}

// fail builds a failed check with its problems.
func fail(name, detail string, problems ...Problem) (Check, []Problem) {
	return Check{Name: name, Passed: false, Detail: detail}, problems
}

// pass builds a successful check.
func pass(name, detail string) (Check, []Problem) {
	return Check{Name: name, Passed: true, Detail: detail}, nil
}

// validateProfileKnown checks that the plan names a profile that exists.
func validateProfileKnown(plan Plan) (Check, []Problem) {
	contract, ok := Contract(plan.Decision.Profile)
	if !ok {
		return fail(CheckProfileKnown, "the plan names a profile that has no contract",
			Problem{
				Code:     CheckProfileKnown,
				Severity: SeverityError,
				Detail:   "unknown profile " + string(plan.Decision.Profile),
			},
		)
	}
	return pass(CheckProfileKnown, "the profile "+string(contract.Profile)+" has a contract")
}

// validateDecisionRecorded checks that the decision itself was written down.
func validateDecisionRecorded(plan Plan) (Check, []Problem) {
	artifact, ok := plan.Artifact("decision-record")
	if !ok {
		return fail(CheckDecisionRecorded, "the plan has no decision-record artifact",
			Problem{
				Code:     CheckDecisionRecorded,
				Severity: SeverityError,
				Detail:   "the plan does not track a decision-record artifact",
			},
		)
	}
	if !artifact.Present {
		return fail(CheckDecisionRecorded, "the decision was not recorded",
			Problem{
				Code:     CheckDecisionRecorded,
				Severity: SeverityError,
				Artifact: artifact.Name,
				Detail:   "the decision-record artifact is not present",
			},
		)
	}
	return pass(CheckDecisionRecorded, "the routing decision is recorded")
}

// validateArtifactKnown checks that every artifact name exists in the catalog.
func validateArtifactKnown(plan Plan) (Check, []Problem) {
	var problems []Problem

	names := append([]string{}, plan.Decision.Artifacts...)
	for _, artifact := range plan.Artifacts {
		names = append(names, artifact.Name)
	}

	seen := make(map[string]bool, len(names))
	for _, name := range names {
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, ok := Artifact(name); !ok {
			problems = append(problems, Problem{
				Code:     CheckArtifactKnown,
				Severity: SeverityError,
				Artifact: name,
				Detail:   "unknown artifact " + name,
			})
		}
	}

	if len(problems) > 0 {
		return fail(CheckArtifactKnown, "the plan names artifacts that do not exist", problems...)
	}
	return pass(CheckArtifactKnown, "every artifact name resolves in the catalog")
}

// validateRequiredPresent checks that the plan produced every artifact the
// decision promised.
func validateRequiredPresent(plan Plan) (Check, []Problem) {
	var problems []Problem

	for _, name := range plan.Decision.Artifacts {
		artifact, ok := plan.Artifact(name)
		if !ok {
			problems = append(problems, Problem{
				Code:     CheckRequiredPresent,
				Severity: SeverityError,
				Artifact: name,
				Detail:   "the decision requires " + name + " but the plan does not track it",
			})
			continue
		}
		if !artifact.Present {
			problems = append(problems, Problem{
				Code:     CheckRequiredPresent,
				Severity: SeverityError,
				Artifact: name,
				Detail:   "the decision requires " + name + " but it is not present",
			})
		}
	}

	if len(problems) > 0 {
		return fail(CheckRequiredPresent, "the plan still owes artifacts", problems...)
	}
	return pass(CheckRequiredPresent, "every promised artifact is present")
}

// validateDependencies checks that the artifacts a plan declares in order
// actually exist, and that the dependencies they name exist too.
func validateDependencies(plan Plan) (Check, []Problem) {
	var problems []Problem

	for _, artifact := range plan.Artifacts {
		for _, dependency := range artifact.DependsOn {
			if _, ok := plan.Artifact(dependency); !ok {
				problems = append(problems, Problem{
					Code:     CheckDependencies,
					Severity: SeverityError,
					Artifact: artifact.Name,
					Detail:   artifact.Name + " depends on " + dependency + ", which the plan does not track",
				})
				continue
			}
			if !artifact.Present {
				continue
			}
			dependencyArtifact, _ := plan.Artifact(dependency)
			if !dependencyArtifact.Present {
				problems = append(problems, Problem{
					Code:     CheckDependencies,
					Severity: SeverityError,
					Artifact: artifact.Name,
					Detail:   artifact.Name + " is present while its dependency " + dependency + " is not",
				})
			}
		}
	}

	if len(problems) > 0 {
		return fail(CheckDependencies, "the plan breaks its own ordering", problems...)
	}
	return pass(CheckDependencies, "the artifact ordering holds")
}

// validateScenarioCoverage checks that every scenario has a verification, and
// that no verification claims a scenario that does not exist.
func validateScenarioCoverage(plan Plan) (Check, []Problem) {
	var problems []Problem

	scenarios := plan.Scenarios()
	known := make(map[string]bool, len(scenarios))
	for _, scenario := range scenarios {
		known[scenario.ID] = true
	}

	covered := make(map[string]bool, len(scenarios))
	for _, verification := range plan.Verifications() {
		for _, scenarioID := range verification.Covers {
			if !known[scenarioID] {
				problems = append(problems, Problem{
					Code:     CheckScenarioCoverage,
					Severity: SeverityError,
					Detail:   "verification " + verification.ID + " covers unknown scenario " + scenarioID,
				})
				continue
			}
			covered[scenarioID] = true
		}
	}

	for _, scenario := range scenarios {
		if !covered[scenario.ID] {
			problems = append(problems, Problem{
				Code:     CheckScenarioCoverage,
				Severity: SeverityError,
				Detail:   "scenario " + scenario.ID + " has no verification",
			})
		}
	}

	if len(problems) > 0 {
		return fail(CheckScenarioCoverage, "the scenarios are not fully covered", problems...)
	}
	return pass(CheckScenarioCoverage, "every scenario has a verification")
}

// validateVerificationGate checks that the plan carries the verification tier
// the decision required.
func validateVerificationGate(plan Plan) (Check, []Problem) {
	verifications := plan.Verifications()

	if len(verifications) == 0 {
		return fail(CheckVerificationGate, "the plan records no verification",
			Problem{
				Code:     CheckVerificationGate,
				Severity: SeverityError,
				Detail:   "no verification is recorded in the plan",
			},
		)
	}

	if plan.Decision.Verification == VerificationIndependent {
		for _, verification := range verifications {
			if verification.Independent {
				return pass(CheckVerificationGate, "an independent verification is recorded")
			}
		}
		return fail(CheckVerificationGate, "the decision required an independent verifier",
			Problem{
				Code:     CheckVerificationGate,
				Severity: SeverityError,
				Detail:   "no verification was produced by someone other than the writer",
			},
		)
	}

	return pass(CheckVerificationGate, "the writer's own verification is recorded")
}

// validateNoOrphans checks that the plan holds nothing the decision did not
// ask for.
func validateNoOrphans(plan Plan) (Check, []Problem) {
	required := make(map[string]bool, len(plan.Decision.Artifacts))
	for _, name := range plan.Decision.Artifacts {
		required[name] = true
	}

	var problems []Problem
	for _, artifact := range plan.Artifacts {
		if !required[artifact.Name] {
			problems = append(problems, Problem{
				Code:     CheckNoOrphans,
				Severity: SeverityError,
				Artifact: artifact.Name,
				Detail:   artifact.Name + " is not part of the decision; extend the profile contract instead",
			})
		}
	}

	if len(problems) > 0 {
		return fail(CheckNoOrphans, "the plan holds artifacts outside its contract", problems...)
	}
	return pass(CheckNoOrphans, "every artifact in the plan belongs to the contract")
}
