package coverage

import "testing"

func TestEncodeDecodePath(t *testing.T) {
	encoded := EncodePath("services/api/coverage/lcov.info")
	decoded, err := DecodePath(encoded)
	if err != nil {
		t.Fatalf("DecodePath() error = %v", err)
	}
	if decoded != "services/api/coverage/lcov.info" {
		t.Fatalf("DecodePath() = %q, want %q", decoded, "services/api/coverage/lcov.info")
	}
}

func TestStoragePath(t *testing.T) {
	path := StoragePath("coverage-data", "services/api/coverage/lcov.info")
	want := ".undercov/coverage-data/c2VydmljZXMvYXBpL2NvdmVyYWdlL2xjb3YuaW5mbw.lcov"
	if path != want {
		t.Fatalf("StoragePath() = %q, want %q", path, want)
	}
}
