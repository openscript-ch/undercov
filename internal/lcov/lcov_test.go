package lcov

import "testing"

func TestParseAndAggregate(t *testing.T) {
	files, err := Parse([]byte("SF:foo.go\nLF:4\nLH:3\nend_of_record\nSF:bar.go\nLF:6\nLH:6\nend_of_record\n"))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(files) != 2 {
		t.Fatalf("Parse() len = %d, want 2", len(files))
	}
	if files[0].CoveragePct < 74.9 || files[0].CoveragePct > 75.1 {
		t.Fatalf("first file coverage = %v, want about 75", files[0].CoveragePct)
	}

	coverage, err := Aggregate([]byte("SF:foo.go\nLF:4\nLH:3\nend_of_record\nSF:bar.go\nLF:6\nLH:6\nend_of_record\n"))
	if err != nil {
		t.Fatalf("Aggregate() error = %v", err)
	}
	if coverage < 89.9 || coverage > 90.1 {
		t.Fatalf("Aggregate() = %v, want about 90", coverage)
	}
}
