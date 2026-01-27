package ainvoke

// AgentConfig describes how to run an agent.
type AgentConfig struct {
	Cmd    []string `json:"cmd,omitempty"     mapstructure:"cmd"`
	UseTTY bool     `json:"use_tty,omitempty" mapstructure:"use_tty"`
}
