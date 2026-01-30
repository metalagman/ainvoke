package integration_tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestExecAgentIntegration(t *testing.T) {
	// Build the helloagent binary
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	
	binPath := filepath.Join(tmpDir, "helloagent")
	srcPath := filepath.Join(origDir, "..", "testdata", "helloagent", "main.go")

	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build helloagent: %v\nOutput: %s", err, string(out))
	}

	mainGoContent := `
package main

import (
	"context"
	"log"
	"os"

	"github.com/metalagman/ainvoke/adk"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

type mockInvocationContext struct {
	context.Context
	userContent *genai.Content
	workDir     string
}

func (m *mockInvocationContext) UserContent() *genai.Content { return m.userContent }
func (m *mockInvocationContext) InvocationID() string       { return "test-id" }
func (m *mockInvocationContext) Artifacts() agent.Artifacts  { return nil }
func (m *mockInvocationContext) Memory() agent.Memory        { return nil }
func (m *mockInvocationContext) Session() session.Session    { return nil }
func (m *mockInvocationContext) Agent() agent.Agent          { return nil }
func (m *mockInvocationContext) Branch() string             { return "" }
func (m *mockInvocationContext) RunConfig() *agent.RunConfig { return nil }
func (m *mockInvocationContext) EndInvocation()             {}
func (m *mockInvocationContext) Ended() bool                { return false }

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: agent <input> <work-dir>")
	}
	inputText := os.Args[1]
	workDir := os.Args[2]

	myAgent, err := adk.NewExecAgent(
		"TestAgent",
		"Test Description",
		[]string{"` + binPath + `"},
		adk.WithExecAgentInputSchema("{\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"}},\"required\":[\"name\"]}"),
		adk.WithExecAgentOutputSchema("{\"type\":\"object\",\"properties\":{\"result\":{\"type\":\"string\"}},\"required\":[\"result\"]}"),
		adk.WithExecAgentRunDir(workDir),
	)
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	userContent := genai.NewContentFromText(inputText, genai.RoleUser)
	ctx := &mockInvocationContext{
		Context:     context.Background(),
		userContent: userContent,
		workDir:     workDir,
	}

	for event, err := range myAgent.Run(ctx) {
		if err != nil {
			log.Fatalf("Run failed: %v", err)
		}
		if event.LLMResponse.Content != nil {
			// Success
			return
		}
	}
}
`
	mainGoPath := filepath.Join(tmpDir, "main.go")
	if err := os.WriteFile(mainGoPath, []byte(mainGoContent), 0644); err != nil {
		t.Fatalf("failed to write main.go: %v", err)
	}

	// Create a go.mod for the temporary agent
	goModContent := `
module tempagent

go 1.25.5

replace github.com/metalagman/ainvoke => ` + origDir + `/..

require (
	github.com/metalagman/ainvoke v0.0.0
	google.golang.org/adk v0.3.0
)
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Tidy the temporary agent's go.mod
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	if out, err := tidyCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to tidy integration agent: %v\nOutput: %s", err, string(out))
	}

	// Build the integration agent
	agentBinPath := filepath.Join(tmpDir, "integration_agent")
	buildCmd := exec.Command("go", "build", "-o", agentBinPath, mainGoPath)
	buildCmd.Dir = tmpDir
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build integration agent: %v\nOutput: %s", err, string(out))
	}

	// Run the integration agent using launcher flags
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}

	input := `{"name": "IntegrationTest"}`
	runCmd := exec.Command(agentBinPath, input, workDir)
	if out, err := runCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to run integration agent: %v\nOutput: %s", err, string(out))
	}

	// Verify the output
	outputPath := filepath.Join(workDir, "output.json")
	outputData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output.json: %v", err)
	}

	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(outputData, &result); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}

	expected := "Hello, IntegrationTest!"
	if result.Result != expected {
		t.Errorf("got %q, want %q", result.Result, expected)
	}
}
