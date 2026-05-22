package openfga

// ConsistencyPreference controls read consistency for relationship queries.
type ConsistencyPreference string

const (
	ConsistencyUnspecified       ConsistencyPreference = "UNSPECIFIED"
	ConsistencyMinimizeLatency   ConsistencyPreference = "MINIMIZE_LATENCY"
	ConsistencyHigherConsistency ConsistencyPreference = "HIGHER_CONSISTENCY"
)
