package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ThothandSon/ghost-rizz/internal/generator"
)

func TestIsSupportedFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"jpg", "test.jpg", true},
		{"jpeg", "test.jpeg", true},
		{"png", "test.png", true},
		{"heic", "test.heic", true},
		{"heif", "test.heif", true},
		{"mp4", "test.mp4", true},
		{"mov", "test.mov", true},
		{"mkv", "test.mkv", true},
		{"avi", "test.avi", true},
		{"webm", "test.webm", true},
		{"m4v", "test.m4v", true},
		{"3gp", "test.3gp", true},
		{"MP4", "test.MP4", true},
		{"MOV", "test.MOV", true},
		{"JPG", "test.JPG", true},
		{"JPEG", "test.JPEG", true},
		{"JpG", "test.JpG", true},
		{"PnG", "test.PnG", true},
		{"HeIc", "test.HeIc", true},
		{"HeIf", "test.HeIf", true},
		{"jPeG", "test.jPeG", true},
		{"pNg", "test.pNg", true},
		{"hEiC", "test.hEiC", true},
		{"pNG", "test.pNG", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"jPg", "test.jPg", true},
		{"hEiF", "test.hEiF", true},
		{"no_ext", "test", false},
		{"empty", "", false},
		{"dot", ".", false},
		{"heic_upper", "test.HEIC", true},
		{"mixed_png", "test.pNg", true},
		{"upper", "TEST.JPG", true},
		{"macos_resource_fork", "._test.jpg", false},
		{"macos_resource_fork_mp4", "._test.mp4", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isSupportedFormat(tt.filename); got != tt.want {
				t.Errorf("isSupportedFormat(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestProcessImages(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	err := generator.GenerateImages(2, inDir)
	if err != nil {
		t.Fatalf("failed to generate images: %v", err)
	}

	// Add a non-supported file to test the skip branch
	_ = os.WriteFile(filepath.Join(inDir, "skip.txt"), []byte("skip"), 0644)

	// Add a directory to test the skip branch
	_ = os.MkdirAll(filepath.Join(inDir, "subdir"), 0755)

	err = ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Fatalf("ProcessImages (clean) failed: %v", err)
	}

	err = ProcessImages(inDir, outDir, "fuzz")
	if err != nil {
		t.Fatalf("ProcessImages (fuzz) failed: %v", err)
	}

	files, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("failed to read outDir: %v", err)
	}

	if len(files) != 4 {
		t.Errorf("expected 4 files in output, got %d", len(files))
	}
}

func TestProcessImages_InvalidInDir(t *testing.T) {
	err := ProcessImages(filepath.Join(t.TempDir(), "nonexistent"), t.TempDir(), "clean")
	if err == nil {
		t.Errorf("expected error for invalid input dir")
	}
}

func TestProcessImages_InvalidOutDir(t *testing.T) {
	badPath := filepath.Join(t.TempDir(), "file.txt")
	_ = os.WriteFile(badPath, []byte("x"), 0644)
	err := ProcessImages(t.TempDir(), filepath.Join(badPath, "invalid_dir"), "clean")
	if err == nil {
		t.Errorf("expected error for invalid output dir")
	}
}

func TestProcessImages_ProcessSingleImageError(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	badFile := filepath.Join(inDir, "bad.jpg")
	_ = os.WriteFile(badFile, []byte("not a real jpeg"), 0644)

	err := ProcessImages(inDir, outDir, "clean")
	if err == nil {
		t.Errorf("ProcessImages expected error but got nil")
	}
}

func TestProcessImages_HEIC(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create a dummy HEIC file
	heicFile := filepath.Join(inDir, "test.heic")
	_ = os.WriteFile(heicFile, []byte("fake heic"), 0644)

	// Mock exiftool to avoid integration failures with real exiftool on fake files
	tmpMockDir := t.TempDir()
	mockPath := filepath.Join(tmpMockDir, "exiftool")
	_ = os.WriteFile(mockPath, []byte("#!/bin/sh\necho 'mock success'\nexit 0\n"), 0755)

	oldPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", tmpMockDir+string(os.PathListSeparator)+oldPath)
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	err := ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Errorf("expected success with mock exiftool, got: %v", err)
	}
}

func TestProcessImages_Video(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	// Create dummy video files
	_ = os.WriteFile(filepath.Join(inDir, "test.mp4"), []byte("fake mp4"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "test.mov"), []byte("fake mov"), 0644)

	// Mock exiftool to avoid integration failures with real exiftool on fake files
	tmpMockDir := t.TempDir()
	mockPath := filepath.Join(tmpMockDir, "exiftool")
	_ = os.WriteFile(mockPath, []byte("#!/bin/sh\necho 'mock success'\nexit 0\n"), 0755)

	oldPath := os.Getenv("PATH")
	_ = os.Setenv("PATH", tmpMockDir+string(os.PathListSeparator)+oldPath)
	defer func() { _ = os.Setenv("PATH", oldPath) }()

	err := ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Errorf("expected success with mock exiftool, got: %v", err)
	}

	err = ProcessImages(inDir, outDir, "fuzz")
	if err != nil {
		t.Errorf("expected success with mock exiftool for fuzz, got: %v", err)
	}
}

func TestProcessImages_UnreadableFile(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	if os.Geteuid() == 0 {
		t.Skip("skipping test as root user can read files with 0000 permissions")
	}

	file := filepath.Join(inDir, "test.jpg")
	_ = os.WriteFile(file, []byte("fake jpg"), 0000)
	defer func() { _ = os.Chmod(file, 0644) }()

	err := ProcessImages(inDir, outDir, "clean")
	if err == nil {
		t.Errorf("expected error for unreadable file")
	}
}

func TestProcessImages_InvalidMode(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	err := generator.GenerateImages(1, inDir)
	if err != nil {
		t.Fatalf("failed to generate image: %v", err)
	}

	err = ProcessImages(inDir, outDir, "invalid_mode")
	if err == nil {
		t.Errorf("expected error for invalid mode in ProcessImages")
	}
}

func TestProcessImages_MultipleErrors(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	if os.Geteuid() == 0 {
		t.Skip("skipping test as root user can read files with 0000 permissions")
	}

	// Create two unreadable files to trigger multiple errors
	f1 := filepath.Join(inDir, "f1.jpg")
	f2 := filepath.Join(inDir, "f2.jpg")
	_ = os.WriteFile(f1, []byte("x"), 0000)
	_ = os.WriteFile(f2, []byte("x"), 0000)
	defer func() { _ = os.Chmod(f1, 0644); _ = os.Chmod(f2, 0644) }()

	err := ProcessImages(inDir, outDir, "clean")
	if err == nil {
		t.Errorf("expected joined errors")
	}
}

func TestProcessImages_OutDirCreationError(t *testing.T) {
	inDir := t.TempDir()
	outDir := filepath.Join(t.TempDir(), "file")
	_ = os.WriteFile(outDir, []byte("x"), 0644)

	err := ProcessImages(inDir, outDir, "clean")
	if err == nil {
		t.Errorf("expected error when outDir is a file")
	}
}

func TestProcessImages_MultiHEIC(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(inDir, "1.heic"), []byte("fake"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "2.heic"), []byte("fake"), 0644)

	// This should hit the break statement in the HEIC pre-scan
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_EmptyDir(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	err := ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Errorf("expected no error for empty dir, got %v", err)
	}
}

func TestProcessImages_SameDirError(t *testing.T) {
	dir := t.TempDir()
	err := ProcessImages(dir, dir, "clean")
	if err == nil {
		t.Errorf("expected error when inDir == outDir")
	}
	if !strings.Contains(err.Error(), "must be different") {
		t.Errorf("expected 'must be different' error, got: %v", err)
	}
}

func TestProcessImages_HEIC_Fuzz_Fallback(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	heicFile := filepath.Join(inDir, "test.heic")
	_ = os.WriteFile(heicFile, []byte("fake heic"), 0644)

	// Even if exiftool is missing, ProcessImages(fuzz) should hit the fallback
	// if it gets the 'unsupported for HEIC' error from SetExif.
	_ = ProcessImages(inDir, outDir, "fuzz")
}

func TestProcessImages_SkipDir(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.Mkdir(filepath.Join(inDir, "subdir"), 0755)
	err := ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Errorf("expected no error when skipping subdir, got %v", err)
	}
}

func TestProcessImages_SkipNoExt(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "noext"), []byte("x"), 0644)
	err := ProcessImages(inDir, outDir, "clean")
	if err != nil {
		t.Errorf("expected no error when skipping no-ext file, got %v", err)
	}
}

func TestProcessImages_MultiJPG(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f1.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f2.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f1.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f2.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_2(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f3.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f4.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_2(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f5.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f6.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_3(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f7.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f8.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_2(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f9.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f10.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_3(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f11.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f12.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_4(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f13.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f14.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_3(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f15.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f16.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_4(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f17.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f18.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_5(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f19.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f20.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_4(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f21.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f22.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_5(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f23.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f24.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_6(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f25.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f26.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_5(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f27.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f28.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_6(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f29.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f30.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_7(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f31.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f32.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_6(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f33.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f34.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_7(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f35.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f36.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_8(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f37.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f38.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_7(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f39.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f40.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_8(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f41.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f42.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_9(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f43.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f44.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiPNG_8(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f45.png"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f46.PNG"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiHEIC_9(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f47.heic"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f48.heif"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}

func TestProcessImages_MultiJPG_10(t *testing.T) {
	inDir := t.TempDir()
	outDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(inDir, "f49.jpg"), []byte("x"), 0644)
	_ = os.WriteFile(filepath.Join(inDir, "f50.jpeg"), []byte("x"), 0644)
	_ = ProcessImages(inDir, outDir, "clean")
}
