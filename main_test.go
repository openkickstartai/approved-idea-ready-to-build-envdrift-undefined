package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestScanFileGo(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "app.go", "package main\nimport \"os\"\nfunc main() {\n\tdb := os.Getenv(\"DATABASE_URL\")\n\ts, _ := os.LookupEnv(\"JWT_SECRET\")\n\t_ = db + s\n}\n")
	vars, err := ScanFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vars["DATABASE_URL"]; !ok {
		t.Error("expected DATABASE_URL")
	}
	if _, ok := vars["JWT_SECRET"]; !ok {
		t.Error("expected JWT_SECRET")
	}
	if len(vars) != 2 {
		t.Errorf("expected 2 vars, got %d", len(vars))
	}
}

func TestScanFilePython(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "app.py", "import os\ndb = os.environ[\"DATABASE_URL\"]\nkey = os.environ.get('API_KEY')\nport = os.getenv('PORT')\n")
	vars, err := ScanFile(p)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"DATABASE_URL", "API_KEY", "PORT"} {
		if _, ok := vars[name]; !ok {
			t.Errorf("expected %s", name)
		}
	}
	if len(vars) != 3 {
		t.Errorf("expected 3 vars, got %d", len(vars))
	}
}

func TestScanFileJS(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, "app.js", "const db = process.env.DATABASE_URL;\nconst key = process.env[\"API_KEY\"];\n")
	vars, err := ScanFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := vars["DATABASE_URL"]; !ok {
		t.Error("expected DATABASE_URL")
	}
	if _, ok := vars["API_KEY"]; !ok {
		t.Error("expected API_KEY")
	}
}

func TestParseEnvFile(t *testing.T) {
	dir := t.TempDir()
	p := writeFile(t, dir, ".env.example", "# Database\nDATABASE_URL=postgres://localhost/db\nAPI_KEY=\nPORT=3000\n")
	vars, err := ParseEnvFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if len(vars) != 3 {
		t.Errorf("expected 3, got %d", len(vars))
	}
	if vars["PORT"] != "3000" {
		t.Errorf("PORT=%q, want 3000", vars["PORT"])
	}
	if vars["API_KEY"] != "" {
		t.Errorf("API_KEY=%q, want empty", vars["API_KEY"])
	}
}

func TestComputeDriftMissingAndUnused(t *testing.T) {
	code := map[string][]Location{
		"DATABASE_URL": {{File: "a.go", Line: 3}},
		"JWT_SECRET":   {{File: "a.go", Line: 4}},
	}
	defined := map[string]string{"DATABASE_URL": "x", "LOG_LEVEL": "debug"}
	r := ComputeDrift(code, defined)
	if len(r.Missing) != 1 || r.Missing[0].Name != "JWT_SECRET" {
		t.Errorf("missing=%+v, want [JWT_SECRET]", r.Missing)
	}
	if len(r.Unused) != 1 || r.Unused[0].Name != "LOG_LEVEL" {
		t.Errorf("unused=%+v, want [LOG_LEVEL]", r.Unused)
	}
}

func TestComputeDriftClean(t *testing.T) {
	code := map[string][]Location{"PORT": {{File: "a.go", Line: 1}}}
	defined := map[string]string{"PORT": "8080"}
	r := ComputeDrift(code, defined)
	if len(r.Missing) != 0 || len(r.Unused) != 0 {
		t.Error("expected no drift")
	}
	text := r.Text()
	if text != "✓ No drift detected.\n" {
		t.Errorf("unexpected text output: %q", text)
	}
}
