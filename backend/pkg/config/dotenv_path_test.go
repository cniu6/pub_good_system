package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeDotEnv(t *testing.T, dir, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}
	return path
}

func TestFindDotEnvPath_ExeSiblingPreferred(t *testing.T) {
	root := t.TempDir()
	exeDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(root, "workdir")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	want := writeDotEnv(t, exeDir, "PORT=9001\nCORS_ORIGINS=*\n")
	writeDotEnv(t, cwd, "PORT=1111\nCORS_ORIGINS=*\n")
	writeDotEnv(t, root, "PORT=2222\nCORS_ORIGINS=*\n")

	got, ok := findDotEnvPathFrom(exeDir, cwd)
	if !ok {
		t.Fatal("expected to find .env")
	}
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("got %q, want exe sibling %q", got, want)
	}
}

func TestFindDotEnvPath_ParentOfExe(t *testing.T) {
	root := t.TempDir()
	exeDir := filepath.Join(root, "bin")
	if err := os.MkdirAll(exeDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cwd := filepath.Join(root, "workdir")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}

	// 同级没有，上级有
	want := writeDotEnv(t, root, "PORT=9002\nCORS_ORIGINS=*\n")
	writeDotEnv(t, cwd, "PORT=1111\nCORS_ORIGINS=*\n")

	got, ok := findDotEnvPathFrom(exeDir, cwd)
	if !ok {
		t.Fatal("expected to find ../.env relative to exe")
	}
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("got %q, want parent %q", got, want)
	}
}

func TestFindDotEnvPath_SkipBackendDotEnv(t *testing.T) {
	root := t.TempDir()
	backendDir := filepath.Join(root, "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// cwd 在 backend 下：旧逻辑会先命中 backend/.env；新逻辑必须跳过并命中根
	writeDotEnv(t, backendDir, "PORT=6666\nCORS_ORIGINS=*\nJWT_SECRET=backend-wrong\n")
	want := writeDotEnv(t, root, "PORT=8085\nCORS_ORIGINS=*\nJWT_SECRET=root-ok\n")

	got, ok := findDotEnvPathFrom("", backendDir)
	if !ok {
		t.Fatal("expected to find root .env after skipping backend/.env")
	}
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("got %q, want root %q", got, want)
	}
	if isBackendDirDotEnv(got) {
		t.Fatalf("must not return backend/.env: %q", got)
	}
}

func TestFindDotEnvPath_SkipFrontendViteOnly(t *testing.T) {
	root := t.TempDir()
	frontendDir := filepath.Join(root, "frontend")
	if err := os.MkdirAll(frontendDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeDotEnv(t, frontendDir, "VITE_API_URL=http://localhost:8085\nVITE_HTTP_PROXY=Y\n")
	want := writeDotEnv(t, root, "PORT=8085\nCORS_ORIGINS=*\n")

	got, ok := findDotEnvPathFrom("", frontendDir)
	if !ok {
		t.Fatal("expected to find root .env after skipping frontend VITE-only .env")
	}
	if filepath.Clean(got) != filepath.Clean(want) {
		t.Fatalf("got %q, want root %q", got, want)
	}
}

func TestIsBackendDirDotEnv(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{filepath.Join("C:", "proj", "backend", ".env"), true},
		{filepath.Join("C:", "proj", ".env"), false},
		{filepath.Join("C:", "proj", "frontend", ".env"), false},
		{`/tmp/fst/backend/.env`, true},
		{`/tmp/fst/.env`, false},
	}
	for _, tc := range cases {
		got := isBackendDirDotEnv(tc.path)
		if got != tc.want {
			t.Fatalf("isBackendDirDotEnv(%q)=%v want %v", tc.path, got, tc.want)
		}
	}
}

func TestIsUsableBackendDotEnv_RejectsBackendAndVite(t *testing.T) {
	root := t.TempDir()
	backendDir := filepath.Join(root, "backend")
	_ = os.MkdirAll(backendDir, 0o755)
	bad := writeDotEnv(t, backendDir, "PORT=1\nCORS_ORIGINS=*\n")
	if isUsableBackendDotEnv(bad) {
		t.Fatal("backend/.env must not be usable")
	}

	fe := filepath.Join(root, "frontend")
	_ = os.MkdirAll(fe, 0o755)
	viteOnly := writeDotEnv(t, fe, "VITE_FOO=1\n")
	if isUsableBackendDotEnv(viteOnly) {
		t.Fatal("vite-only .env must not be usable")
	}

	good := writeDotEnv(t, root, "PORT=8080\nCORS_ORIGINS=*\n")
	if !isUsableBackendDotEnv(good) {
		t.Fatal("root backend .env should be usable")
	}
	if !strings.HasSuffix(filepath.Base(good), ".env") {
		t.Fatalf("unexpected path %q", good)
	}
}
