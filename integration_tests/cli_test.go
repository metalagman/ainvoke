package integration_tests

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestAinvokeExec(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	// 1. Build ainvoke binary
	ainvokeBin := filepath.Join(tmpDir, "ainvoke")
	buildCmd := exec.Command("go", "build", "-o", ainvokeBin, filepath.Join(origDir, "..", "cmd", "ainvoke"))
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build ainvoke: %v\nOutput: %s", err, string(out))
	}

	// 2. Build test agent binary (helloagent)
	agentBin := filepath.Join(tmpDir, "helloagent")
	agentSrc := filepath.Join(origDir, "..", "testdata", "helloagent", "main.go")
	buildAgentCmd := exec.Command("go", "build", "-o", agentBin, agentSrc)
	if out, err := buildAgentCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build helloagent: %v\nOutput: %s", err, string(out))
	}

	// 3. Run ainvoke exec
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}

	input := `{"name": "CLIIntegration"}`
	// The default schemas for ainvoke CLI are object-based:
	// input: {"input": "string"}
	// output: {"output": "string"}
	// But helloagent expects: {"name": "string"} -> {"result": "string"}
	// So we need to provide explicit schemas.
	
	inputSchema := `{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`
	outputSchema := `{"type":"object","properties":{"result":{"type":"string"}},"required":["result"]}`

	runCmd := exec.Command(ainvokeBin, "exec", agentBin, 
		"--input", input, 
		"--work-dir", workDir,
		"--input-schema", inputSchema,
		"--output-schema", outputSchema,
		"--tty=false",
	)
	
out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ainvoke exec failed: %v\nOutput: %s", err, string(out))
	}

	// 4. Verify output from stdout (ainvoke prints output.json to stdout)
	var result struct {
		Result string `json:"result"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("failed to unmarshal ainvoke output: %v\nRaw output: %s", err, string(out))
	}

	expected := "Hello, CLIIntegration!"
	if result.Result != expected {
		t.Errorf("got %q, want %q", result.Result, expected)
	}

	// 5. Verify output file exists in workDir
	outputFile := filepath.Join(workDir, "output.json")
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Error("output.json was not created in work-dir")
	}
}

func TestAinvokeExec_DefaultSchemas(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()

	// 1. Build ainvoke binary
	ainvokeBin := filepath.Join(tmpDir, "ainvoke")
	buildCmd := exec.Command("go", "build", "-o", ainvokeBin, filepath.Join(origDir, "..", "cmd", "ainvoke"))
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to build ainvoke: %v\nOutput: %s", err, string(out))
	}

	// 2. Create a simple script that works with default schemas
	// Default input: {"input": "..."}
	// Default output: {"output": "..."}
	scriptPath := filepath.Join(tmpDir, "agent.sh")
	scriptContent := `#!/bin/bash
NAME=$(jq -r .input input.json)
echo "{\"output\": \"Hello, $NAME\"}" > output.json
`
	if err := os.WriteFile(scriptPath, []byte(scriptContent), 0755); err != nil {
		t.Fatalf("failed to write script: %v", err)
	}

	// 3. Run ainvoke exec
	workDir := filepath.Join(tmpDir, "work")
	if err := os.MkdirAll(workDir, 0755); err != nil {
		t.Fatalf("failed to create work dir: %v", err)
	}

	// Use plain string for --input, ainvoke should wrap it in {"input": "..."} if it's not JSON
	input := "DefaultSchemaUser"

	runCmd := exec.Command(ainvokeBin, "exec", 
		"--input", input, 
		"--work-dir", workDir,
		"--tty=false",
		"--", "bash", scriptPath,
	)
	
out, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("ainvoke exec failed: %v\nOutput: %s", err, string(out))
	}

	var result struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("failed to unmarshal ainvoke output: %v\nRaw output: %s", err, string(out))
	}

	expected := "Hello, DefaultSchemaUser"
	if result.Output != expected {
		t.Errorf("got %q, want %q", result.Output, expected)
	}
}
