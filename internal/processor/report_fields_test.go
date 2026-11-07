package processor

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"

	"github.com/ThothandSon/ghost-rizz/internal/generator"
)

// Regression: Make, Model, Software and DateTime live in IFD0, whose path in go-exif v3
// is "IFD". The report looked them up under "IFD0" and always left them empty.
func TestReportIncludesCameraFields(t *testing.T) {
	in := t.TempDir()
	if err := generator.GenerateImages(2, in); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "report.csv")
	summary, err := ReportDir(in, out)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Files != 2 || summary.WithMetadata != 2 {
		t.Fatalf("summary = %+v", summary)
	}

	f, _ := os.Open(out)
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows[1:] {
		if row[8] != "40/1 45/1 0/1" {
			t.Errorf("%s: GPS latitude = %q, want all three components", row[0], row[8])
		}
		if row[2] != "Ghost-Rizz Generator" || row[3] != "Dummy 1.0" || row[4] != "Ghost-Rizz OS 1.0" || row[5] == "" {
			t.Errorf("%s: make=%q model=%q software=%q datetime=%q", row[0], row[2], row[3], row[4], row[5])
		}
	}
}
