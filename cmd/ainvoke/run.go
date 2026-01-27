package main

import "strings"

func appendCodexFlags(argv []string, model string) []string {
	out := make([]string, 0, len(argv))
	out = append(out, argv...)

	if len(out) > 0 && out[0] == "codex" {
		if len(out) == 1 || !isCodexSubcommand(out[1]) {
			out = append(out[:1], append([]string{"exec"}, out[1:]...)...)
		}
	}

	if model != "" && !hasFlag(out, "--model") && !hasFlag(out, "-m") {
		out = append(out, "--model", model)
	}

	return out
}

func appendOpenCodeFlags(argv []string, model string) []string {
	out := make([]string, 0, len(argv))
	out = append(out, argv...)

	if len(out) > 0 && out[0] == "opencode" {
		if len(out) == 1 || out[1] == "" || strings.HasPrefix(out[1], "-") || !isOpenCodeSubcommand(out[1]) {
			out = append(out[:1], append([]string{"run"}, out[1:]...)...)
		}
	}

	if model != "" && !hasFlag(out, "--model") && !hasFlag(out, "-m") {
		out = append(out, "--model", model)
	}

	return out
}

func appendGeminiFlags(argv []string, model string) []string {
	out := make([]string, 0, len(argv))
	out = append(out, argv...)

	if model != "" && !hasFlag(out, "--model") && !hasFlag(out, "-m") {
		out = append(out, "--model", model)
	}

	if !hasFlag(out, "--output-format") {
		out = append(out, "--output-format", "text")
	}

	return out
}

func appendClaudeFlags(argv []string, model string) []string {
	out := make([]string, 0, len(argv))
	out = append(out, argv...)

	if model != "" && !hasFlag(out, "--model") && !hasFlag(out, "-m") {
		out = append(out, "--model", model)
	}

	return out
}

func hasFlag(argv []string, name string) bool {
	for _, arg := range argv {
		if arg == name {
			return true
		}
	}

	return false
}

func isCodexSubcommand(arg string) bool {
	if arg == "" || strings.HasPrefix(arg, "-") {
		return false
	}

	switch arg {
	case "exec", "review", "login", "logout", "mcp", "mcp-server", "app-server",
		"completion", "sandbox", "apply", "resume", "fork", "cloud", "features", "help":
		return true
	default:
		return false
	}
}

func isOpenCodeSubcommand(arg string) bool {
	switch arg {
	case "agent", "attach", "auth", "github", "mcp", "models", "run", "serve",
		"session", "stats", "export", "import", "web", "acp", "uninstall", "upgrade", "help":
		return true
	default:
		return false
	}
}
