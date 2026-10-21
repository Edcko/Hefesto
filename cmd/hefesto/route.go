package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Edcko/Hefesto/cmd/hefesto/router"
	"github.com/spf13/cobra"
)

var (
	routeType      string
	routeSize      string
	routeRisk      string
	routeChange    string
	routeJSON      bool
	routeWritePlan string
	verifyPlanJS   bool
)

var routeCmd = &cobra.Command{
	Use:   "route",
	Short: "Decide which development methodology profile fits a request",
	Long: `Decide the methodology profile for a piece of work and print everything
the run needs: the decision, the contract of the chosen profile, the document
the run records itself in, and the plan the run starts from.

The decision is computed, not improvised: the same inputs always produce the
same profile, the same artifacts and the same verification tier.`,
	Example: `  hefesto route --type known --size large --risk none --change add-widget
  hefesto route --type ambiguous_behavior --size small --risk contract --json
  hefesto route --type known --size large --risk none --change add-widget --write-plan plan.json`,
	RunE: runRoute,
}

var verifyPlanCmd = &cobra.Command{
	Use:   "verify-plan <plan.json>",
	Short: "Validate a plan against the contract of its profile",
	Long: `Validate a plan document against the contract of the profile it was routed to.

The plan records the routing decision, the artifacts the run produced, the
scenarios the specification states and the verifications that prove them. This
command reports every rule the plan breaks, and exits non-zero when the plan
still owes work.`,
	Example: `  hefesto verify-plan plan.json
  hefesto verify-plan plan.json --json`,
	Args: cobra.ExactArgs(1),
	RunE: runVerifyPlan,
}

func init() {
	routeCmd.Flags().StringVar(&routeType, "type", "", "dominant uncertainty: known, ambiguous_behavior, unknown_domain, unvalidated_idea")
	routeCmd.Flags().StringVar(&routeSize, "size", "", "size of the work: small or large")
	routeCmd.Flags().StringVar(&routeRisk, "risk", string(router.RiskNone), "high-risk item touched: none, data, security, contract, concurrency, delivery, untestable")
	routeCmd.Flags().StringVar(&routeChange, "change", "", "name of the change being routed")
	routeCmd.Flags().BoolVar(&routeJSON, "json", false, "Output the routing payload in JSON format")
	routeCmd.Flags().StringVar(&routeWritePlan, "write-plan", "", "write the plan template to this path")

	verifyPlanCmd.Flags().BoolVar(&verifyPlanJS, "json", false, "Output the validation report in JSON format")
}

// buildRouteRequest turns the command flags into a routing request.
func buildRouteRequest() (router.Request, error) {
	if routeType == "" {
		return router.Request{}, fmt.Errorf("--type is required (known, ambiguous_behavior, unknown_domain, unvalidated_idea)")
	}
	if routeSize == "" {
		return router.Request{}, fmt.Errorf("--size is required (small or large)")
	}
	if routeRisk == "" {
		routeRisk = string(router.RiskNone)
	}

	return router.Request{
		Type: router.WorkType(routeType),
		Size: router.Size(routeSize),
		Risk: router.RiskItem(routeRisk),
	}, nil
}

func runRoute(cmd *cobra.Command, args []string) error {
	request, err := buildRouteRequest()
	if err != nil {
		return err
	}

	payload, err := router.Route(request, routeChange)
	if err != nil {
		return err
	}

	if routeWritePlan != "" {
		if err := writePlanFile(routeWritePlan, payload.Plan); err != nil {
			return err
		}
	}

	if routeJSON {
		return printRouteJSON(payload)
	}

	printRouteText(payload)
	return nil
}

// writePlanFile writes a plan document to a path.
func writePlanFile(path string, plan router.Plan) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal the plan: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write plan %s: %w", path, err)
	}
	fmt.Printf("  Plan written to %s\n", path)
	return nil
}

// printRouteJSON outputs the routing payload as JSON.
func printRouteJSON(payload router.Payload) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printRouteText outputs the routing payload for a human reader.
func printRouteText(payload router.Payload) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Route")
	fmt.Println()

	change := payload.Change
	if change == "" {
		change = "(unnamed change)"
	}
	fmt.Printf("  Change:    %s\n", change)
	fmt.Printf("  Profile:   %s\n", payload.Decision.Profile)
	fmt.Printf("  Verify:    %s\n", payload.Decision.Verification)
	if payload.Decision.FollowUp != "" {
		fmt.Printf("  Follow-up: %s\n", payload.Decision.FollowUp)
	} else {
		fmt.Println("  Follow-up: none")
	}
	fmt.Println()
	fmt.Printf("  Rationale: %s\n", payload.Decision.Rationale)
	fmt.Println()

	fmt.Println("  Artifacts:")
	for _, name := range payload.Decision.Artifacts {
		spec, _ := router.Artifact(name)
		fmt.Printf("    • %-18s %-9s %s\n", name, spec.Kind, spec.Purpose)
	}
	fmt.Println()

	fmt.Println("  Contract:")
	for _, criterion := range payload.Contract.Entry {
		fmt.Printf("    entry: %s\n", criterion)
	}
	for _, criterion := range payload.Contract.Exit {
		fmt.Printf("    exit:  %s\n", criterion)
	}
	fmt.Printf("    gate:  %s\n", payload.Contract.Gate)
	fmt.Println()

	fmt.Println("  Scaffold:")
	for _, line := range strings.Split(strings.TrimRight(payload.Scaffold, "\n"), "\n") {
		fmt.Printf("    %s\n", line)
	}
	fmt.Println()
}

func runVerifyPlan(cmd *cobra.Command, args []string) error {
	report, err := verifyPlanFile(args[0])
	if err != nil {
		return err
	}

	if verifyPlanJS {
		if err := printVerifyReportJSON(report); err != nil {
			return err
		}
	} else {
		printVerifyReportText(report)
	}

	if !report.OK() {
		return fmt.Errorf("the plan does not satisfy the contract of profile %q", report.Profile)
	}
	return nil
}

// verifyPlanFile reads a plan document and validates it.
func verifyPlanFile(path string) (router.Report, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: the path is given by the operator running the command
	if err != nil {
		return router.Report{}, fmt.Errorf("failed to read plan %s: %w", path, err)
	}

	var plan router.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return router.Report{}, fmt.Errorf("failed to parse plan %s: %w", path, err)
	}

	return router.Validate(plan), nil
}

// printVerifyReportJSON outputs the validation report as JSON.
func printVerifyReportJSON(report router.Report) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// printVerifyReportText outputs the validation report for a human reader.
func printVerifyReportText(report router.Report) {
	fmt.Println()
	fmt.Println("🔥 Hefesto Plan Verification")
	fmt.Println()
	fmt.Printf("  Profile: %s\n", report.Profile)
	fmt.Println()

	for _, check := range report.Checks {
		mark := "✅"
		if !check.Passed {
			mark = "❌"
		}
		fmt.Printf("  %s %-28s %s\n", mark, check.Name, check.Detail)
	}
	fmt.Println()

	if len(report.Problems) == 0 {
		fmt.Println("  The plan satisfies its contract.")
		fmt.Println()
		return
	}

	for _, problem := range report.Problems {
		artifact := ""
		if problem.Artifact != "" {
			artifact = " [" + problem.Artifact + "]"
		}
		fmt.Printf("  ❌ %s%s: %s\n", problem.Code, artifact, problem.Detail)
	}
	fmt.Println()
}
