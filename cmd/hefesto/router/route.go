package router

import "fmt"

// Payload is everything an orchestrator needs after a routing decision: the
// decision itself, the contract of the chosen profile, the document the run
// records itself in, and the plan the run starts from.
type Payload struct {
	// Change is the name of the change being routed.
	Change string `json:"change"`
	// Decision is the routing decision.
	Decision Decision `json:"decision"`
	// Contract is the contract of the chosen profile.
	Contract ProfileContract `json:"contract"`
	// Scaffold is the markdown document the run records itself in.
	Scaffold string `json:"scaffold"`
	// Plan is the plan the run starts from.
	Plan Plan `json:"plan"`
}

// Route decides the methodology for a request and assembles everything the
// run needs to follow it.
func Route(req Request, change string) (Payload, error) {
	decision, err := Decide(req)
	if err != nil {
		return Payload{}, err
	}

	contract, ok := Contract(decision.Profile)
	if !ok {
		return Payload{}, fmt.Errorf("router: no contract for profile %q", decision.Profile)
	}

	return Payload{
		Change:   change,
		Decision: decision,
		Contract: contract,
		Scaffold: Scaffold(change, decision, contract),
		Plan:     PlanTemplate(change, decision),
	}, nil
}
