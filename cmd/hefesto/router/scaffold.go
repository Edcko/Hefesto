package router

import (
	"fmt"
	"strings"
)

// artifactDependencies records the order the artifacts of a profile are
// produced in. An artifact may only exist once the artifacts it depends on do.
var artifactDependencies = map[string][]string{
	"one-red-test":       {"clarification-note"},
	"spec":               {"proposal"},
	"tasks":              {"spec"},
	"verify-report":      {"tasks"},
	"options-table":      {"exploration-map"},
	"open-decisions":     {"options-table"},
	"feedback-log":       {"prototype"},
	"rollback-plan":      nil,
	"focused-test":       nil,
	"decision-record":    nil,
	"exploration-map":    nil,
	"proposal":           nil,
	"prototype":          nil,
	"clarification-note": nil,
}

// scenarioHolder returns the artifact that carries the scenarios of a
// decision, if the profile states any.
func scenarioHolder(artifacts []string) string {
	for _, name := range artifacts {
		if name == "spec" {
			return name
		}
	}
	return ""
}

// verificationHolder returns the artifact that carries the verifications of a
// decision: the last artifact the profile owes.
func verificationHolder(artifacts []string) string {
	if len(artifacts) == 0 {
		return ""
	}
	return artifacts[len(artifacts)-1]
}

// PlanTemplate returns the plan a run starts from: every artifact the
// decision promised is tracked and still missing.
func PlanTemplate(change string, decision Decision) Plan {
	plan := Plan{Change: change, Decision: decision}

	for _, name := range decision.Artifacts {
		plan.Artifacts = append(plan.Artifacts, PlannedArtifact{
			Name:      name,
			Present:   false,
			DependsOn: artifactDependencies[name],
		})
	}

	return plan
}

// Scaffold returns the markdown document a run records itself in: the
// decision, the criteria that opened and close the profile, the artifacts it
// owes, and the tables that hold the scenarios and their verifications.
func Scaffold(change string, decision Decision, contract ProfileContract) string {
	var out strings.Builder

	title := change
	if title == "" {
		title = "change"
	}

	fmt.Fprintf(&out, "# %s\n\n", title)
	out.WriteString("## Decision\n\n")
	fmt.Fprintf(&out, "- Profile: `%s`\n", decision.Profile)
	fmt.Fprintf(&out, "- Verification: `%s`\n", decision.Verification)
	if decision.FollowUp != "" {
		fmt.Fprintf(&out, "- Follow-up: `%s`\n", decision.FollowUp)
	} else {
		out.WriteString("- Follow-up: none\n")
	}
	fmt.Fprintf(&out, "- Rationale: %s\n\n", decision.Rationale)

	out.WriteString("## Entry criteria\n\n")
	for _, criterion := range contract.Entry {
		fmt.Fprintf(&out, "- [ ] %s\n", criterion)
	}
	out.WriteString("\n")

	out.WriteString("## Artifacts\n\n")
	for _, name := range decision.Artifacts {
		spec, _ := Artifact(name)
		fmt.Fprintf(&out, "- [ ] `%s` (%s) — %s\n", name, spec.Kind, spec.Purpose)
	}
	out.WriteString("\n")

	holder := scenarioHolder(decision.Artifacts)
	if holder != "" {
		out.WriteString("## Scenarios\n\n")
		out.WriteString("| ID | Scenario | Verified by | Status |\n")
		out.WriteString("|----|----------|-------------|--------|\n")
		out.WriteString("| S1 | | | pending |\n\n")
		fmt.Fprintf(&out, "Scenarios belong to the `%s` artifact.\n\n", holder)
	}

	out.WriteString("## Verification\n\n")
	out.WriteString("| ID | Covers | Command | Independent | Status |\n")
	out.WriteString("|----|--------|---------|------------|--------|\n")
	out.WriteString("| V1 | | | | pending |\n\n")

	out.WriteString("## Exit criteria\n\n")
	for _, criterion := range contract.Exit {
		fmt.Fprintf(&out, "- [ ] %s\n", criterion)
	}
	out.WriteString("\n")

	return out.String()
}
