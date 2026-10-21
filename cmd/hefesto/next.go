package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Edcko/Hefesto/cmd/hefesto/router"
	"github.com/Edcko/Hefesto/cmd/hefesto/runner"
	"github.com/spf13/cobra"
)

var (
	nextPlan string
	nextJSON bool
)

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show the next action a run must take",
	Long: `Show what the next action of a run is, according to the contract of the
profile the plan was routed to.

The step names the artifact in hand, what blocks it and the brief that says
what to read, what to write and what "done" means. The run advances on a
recorded artifact, never on a promise.`,
	Example: `  hefesto next --plan plan.json
  hefesto next --plan plan.json --json`,
	Args: cobra.NoArgs,
	RunE: runNext,
}

func init() {
	nextCmd.Flags().StringVar(&nextPlan, "plan", "", "path to the plan document of the run")
	nextCmd.Flags().BoolVar(&nextJSON, "json", false, "Output the step in JSON format")
}

// buildNextStep reads a plan and computes the next action of the run.
func buildNextStep(planPath string) (runner.Step, error) {
	if planPath == "" {
		return runner.Step{}, fmt.Errorf("--plan is required")
	}

	data, err := os.ReadFile(planPath) //nolint:gosec // G304: the path is given by the operator running the command
	if err != nil {
		return runner.Step{}, fmt.Errorf("failed to read plan %s: %w", planPath, err)
	}

	var plan router.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return runner.Step{}, fmt.Errorf("failed to parse plan %s: %w", planPath, err)
	}

	return runner.Next(plan)
}

func runNext(cmd *cobra.Command, args []string) error {
	step, err := buildNextStep(nextPlan)
	if err != nil {
		return err
	}

	if nextJSON {
		return printNextJSON(step)
	}

	printNextText(step)
	return nil
}

// printNextJSON outputs the step as JSON.
func printNextJSON(step runner.Step) error {
	data, err := json.MarshalIndent(step, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printNextText outputs the step for a human reader.
func printNextText(step runner.Step) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Next")
	fmt.Println()

	switch step.Phase {
	case runner.PhaseClose:
		fmt.Println("  Phase: close — everything the contract requires is present.")
		fmt.Println("  Run `hefesto verify-plan` to close the run.")
		fmt.Println()
		return

	case runner.PhaseComplete:
		fmt.Printf("  Phase:  complete — `%s` exists but is missing content.\n", step.Artifact)
		fmt.Println()
		for _, blocker := range step.Blockers {
			fmt.Printf("  ❌ %s\n", blocker)
		}
		fmt.Println()

	case runner.PhaseProduce:
		fmt.Printf("  Phase:  produce — `%s` (%s)\n", step.Artifact, step.Kind)
		fmt.Println()
		fmt.Printf("  %s\n", step.Purpose)
		fmt.Println()
	}

	if len(step.Inputs) > 0 {
		fmt.Println("  Read first:")
		for _, input := range step.Inputs {
			fmt.Printf("    • %s\n", input)
		}
		fmt.Println()
	}

	if step.Brief != "" {
		fmt.Println("  Brief:")
		for _, line := range strings.Split(strings.TrimRight(step.Brief, "\n"), "\n") {
			fmt.Printf("    %s\n", line)
		}
		fmt.Println()
	}
}
