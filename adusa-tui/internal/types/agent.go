package types

type AgentStatus int

const (
	AgentIdle AgentStatus = iota
	AgentRunning
	AgentDone
	AgentError
)

func (s AgentStatus) String() string {
	switch s {
	case AgentIdle:
		return "Idle"
	case AgentRunning:
		return "Running"
	case AgentDone:
		return "Done"
	case AgentError:
		return "Error"
	default:
		return "Unknown"
	}
}
