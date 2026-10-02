package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCommand(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("x"), 0644)

	tests := []struct {
		name       string
		args       []string
		wantExit   int
		wantOutput []string
	}{
		{
			name:       "No arguments",
			args:       []string{},
			wantExit:   1,
			wantOutput: []string{"expected 'generate', 'clean', 'fuzz' or 'report' subcommands"},
		},
		{
			name:       "Unknown argument",
			args:       []string{"unknown"},
			wantExit:   1,
			wantOutput: []string{"expected 'generate', 'clean', 'fuzz' or 'report' subcommands"},
		},
		{
			name:       "Generate successfully",
			args:       []string{"generate", "-count", "1", "-out", filepath.Join(tmpDir, "input")},
			wantExit:   0,
			wantOutput: []string{"Generation complete."},
		},
		{
			name:       "Fuzz successfully",
			args:       []string{"fuzz", "-in", filepath.Join(tmpDir, "input"), "-out", filepath.Join(tmpDir, "output"), "-mode", "clean"},
			wantExit:   0,
			wantOutput: []string{"Processing complete."},
		},
		{
			name:       "Report successfully",
			args:       []string{"report", "-in", filepath.Join(tmpDir, "output")},
			wantExit:   0,
			wantOutput: []string{"Report complete."},
		},
		{
			name:       "Generate error path",
			args:       []string{"generate", "-out", filepath.Join(tmpDir, "file.txt", "invalid")},
			wantExit:   1,
			wantOutput: []string{"Error generating images"},
		},
		{
			name:       "Fuzz error path",
			args:       []string{"fuzz", "-in", filepath.Join(tmpDir, "nonexistent")},
			wantExit:   1,
			wantOutput: []string{"Error processing images"},
		},
		{
			name:       "Report error path",
			args:       []string{"report", "-in", filepath.Join(tmpDir, "nonexistent")},
			wantExit:   1,
			wantOutput: []string{"Error generating report"},
		},
		{
			name:       "Generate bad flags",
			args:       []string{"generate", "-unknown_flag"},
			wantExit:   1,
			wantOutput: []string{"flag provided but not defined"},
		},
		{
			name:       "Fuzz bad flags",
			args:       []string{"fuzz", "-unknown_flag"},
			wantExit:   1,
			wantOutput: []string{"flag provided but not defined"},
		},
		{
			name:       "Report bad flags",
			args:       []string{"report", "-unknown_flag"},
			wantExit:   1,
			wantOutput: []string{"flag provided but not defined"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var errBuf bytes.Buffer
			stdout = &buf
			stderr = &errBuf
			t.Cleanup(func() {
				stdout = os.Stdout
				stderr = os.Stderr
			})

			got := run(tt.args)

			if got != tt.wantExit {
				t.Errorf("run() = %v, want %v", got, tt.wantExit)
			}
			outStr := buf.String() + errBuf.String()
			for _, w := range tt.wantOutput {
				if w != "" && !strings.Contains(outStr, w) {
					t.Errorf("run() output %q, want to contain %q", outStr, w)
				}
			}
		})
	}
}

func TestMainFunc(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ghost-rizz", "unknown"}

	oldExit := osExit
	defer func() { osExit = oldExit }()
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}

	main()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}
}
