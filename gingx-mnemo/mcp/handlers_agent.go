package mcp

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// handleAgentDispatch dispatches a task to a specialized agent persona.
// Calls the gingx-sdd Python CLI to assemble the prompt (shared engine).
func (s *Server) handleAgentDispatch(args map[string]interface{}) (string, error) {
	agent, _ := args["agent"].(string)
	task, _ := args["task"].(string)

	if agent == "" || task == "" {
		return "", fmt.Errorf("agent and task are required")
	}

	profile, _ := args["profile"].(string)
	if profile == "" {
		profile = "developer"
	}

	hduID, _ := args["hdu_id"].(string)
	techStack, _ := args["tech_stack"].(string)

	// Find project root by looking for .gingx/
	projectRoot := findProjectRoot()

	// Build gingx-sdd CLI command
	cmdArgs := []string{
		"-m", "gingx_sdd", "team", "spawn", agent,
		"--task", task,
		"--profile", profile,
		"--mode", "prompt",
	}

	if hduID != "" {
		cmdArgs = append(cmdArgs, "--hdu-id", hduID)
	}
	if techStack != "" {
		cmdArgs = append(cmdArgs, "--stack", techStack)
	}

	cmd := exec.Command("python3", cmdArgs...)
	cmd.Dir = projectRoot
	cmd.Env = append(cmd.Env,
		"PYTHONPATH="+filepath.Join(projectRoot, "gingx-sdd"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("agent dispatch failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

// findProjectRoot locates the project root by searching for .gingx/ directory.
func findProjectRoot() string {
	dir := "."
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "."
	}

	for {
		gingxDir := filepath.Join(abs, ".gingx")
		if stat, err := exec.Command("test", "-d", gingxDir).CombinedOutput(); err == nil && stat == nil {
			return abs
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return ""
}

// compactContent truncates a string for logging.
func compactContent(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " | ")
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}
