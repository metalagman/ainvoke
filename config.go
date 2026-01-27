// Package ainvoke provides implementations for running different types of agents.
package ainvoke

// AgentConfig describes how to run an agent.
type AgentConfig struct {
	Type   string   `json:"type"              mapstructure:"type"`
	Cmd    []string `json:"cmd,omitempty"     mapstructure:"cmd"`
	Model  string   `json:"model,omitempty"   mapstructure:"model"`
	Path   string   `json:"path,omitempty"    mapstructure:"path"`
	UseTTY *bool    `json:"use_tty,omitempty" mapstructure:"use_tty"`
}
