package taskgraph

type GroupPolicy int

const (
	HaltOnAnyError   GroupPolicy = 1
	HaltAfterTimeout GroupPolicy = 2
)

func containsPolicy(p GroupPolicy, gp GroupPolicy) bool {
	return (p & gp) == gp
}
