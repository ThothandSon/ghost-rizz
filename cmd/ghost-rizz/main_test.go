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
			wantOutput: []string{"Usage:", "ghost-rizz report"},
		},
		{
			name:       "Unknown argument",
			args:       []string{"unknown"},
			wantExit:   1,
			wantOutput: []string{`unknown command "unknown"`, "Usage:"},
		},
		{
			name:       "Help",
			args:       []string{"--help"},
			wantExit:   0,
			wantOutput: []string{"Usage:", "Examples:"},
		},
		{
			name:       "Version",
			args:       []string{"version"},
			wantExit:   0,
			wantOutput: []string{"ghost-rizz dev"},
		},
		{
			name:       "Command help",
			args:       []string{"clean", "-h"},
			wantExit:   0,
			wantOutput: []string{"Remove all metadata", "-out"},
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
			name:       "Clean successfully",
			args:       []string{"clean", "-in", filepath.Join(tmpDir, "input"), "-out", filepath.Join(tmpDir, "cleaned")},
			wantExit:   0,
			wantOutput: []string{"Cleaned 1 file in", "cleaned"},
		},
		{
			name:       "Report successfully",
			args:       []string{"report", "-in", filepath.Join(tmpDir, "output")},
			wantExit:   0,
			wantOutput: []string{"Report complete.", "1 file scanned", "with metadata"},
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

func TestPartialFailureListsFilesAndFails(t *testing.T) {
	in := t.TempDir()
	if code := run([]string{"generate", "-count", "2", "-out", in}); code != 0 {
		t.Fatalf("generate = %d", code)
	}
	_ = os.WriteFile(filepath.Join(in, "broken.jpg"), []byte("not a jpeg"), 0o644)
	_ = os.WriteFile(filepath.Join(in, "notes.txt"), []byte("x"), 0o644)

	var out, errOut bytes.Buffer
	stdout, stderr = &out, &errOut
	t.Cleanup(func() { stdout, stderr = os.Stdout, os.Stderr })

	code := run([]string{"clean", "-in", in, "-out", filepath.Join(t.TempDir(), "out")})
	if code != 1 {
		t.Fatalf("exit = %d, want 1 when a file fails", code)
	}
	for _, want := range []string{"Cleaned 2 files", "1 unsupported entry skipped"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("stdout %q lacks %q", out.String(), want)
		}
	}
	if !strings.Contains(errOut.String(), "1 file failed") || !strings.Contains(errOut.String(), "broken.jpg") {
		t.Errorf("stderr %q doesn't name the broken file", errOut.String())
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
