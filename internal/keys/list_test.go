package keys

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListPrivateKeys_Empty(t *testing.T) {
	sshDir := t.TempDir()
	keys, err := ListPrivateKeys(sshDir)
	if err != nil {
		t.Fatalf("ListPrivateKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(keys))
	}
}

func TestListPrivateKeys_FindsFlatKeys(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sshDir, "mykey"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "mykey.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "other"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sshDir, "other.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	keys, err := ListPrivateKeys(sshDir)
	if err != nil {
		t.Fatalf("ListPrivateKeys: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}

	names := map[string]bool{keys[0].RelPath: true, keys[1].RelPath: true}
	if !names["mykey"] || !names["other"] {
		t.Errorf("unexpected keys: %+v", keys)
	}
}

func TestListPrivateKeys_SkipsPubFiles(t *testing.T) {
	sshDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(sshDir, "id_ed25519.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	keys, err := ListPrivateKeys(sshDir)
	if err != nil {
		t.Fatalf("ListPrivateKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected no keys, got %+v", keys)
	}
}

func TestListPrivateKeys_SkipsConfigFiles(t *testing.T) {
	sshDir := t.TempDir()
	for _, name := range []string{"config", "known_hosts", "authorized_keys"} {
		if err := os.WriteFile(filepath.Join(sshDir, name), []byte("x"), 0600); err != nil {
			t.Fatal(err)
		}
	}

	keys, err := ListPrivateKeys(sshDir)
	if err != nil {
		t.Fatalf("ListPrivateKeys: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("expected no keys, got %+v", keys)
	}
}

func TestListPrivateKeys_FindsSubdirKeys(t *testing.T) {
	sshDir := t.TempDir()
	subDir := filepath.Join(sshDir, "homelab-keys")
	if err := os.Mkdir(subDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(subDir, "mykey"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}

	keys, err := ListPrivateKeys(sshDir)
	if err != nil {
		t.Fatalf("ListPrivateKeys: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0].RelPath != "homelab-keys/mykey" {
		t.Errorf("RelPath = %q, want homelab-keys/mykey", keys[0].RelPath)
	}
}

func TestListPrivateKeys_NoDeepRecurse(t *testing.T) {
	sshDir := t.TempDir()
	deepDir := filepath.Join(sshDir, "a", "b")
	if err := os.MkdirAll(deepDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(deepDir, "key"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}

	keys, err := ListPrivateKeys(sshDir)
	if err != nil {
		t.Fatalf("ListPrivateKeys: %v", err)
	}
	for _, k := range keys {
		if k.RelPath == "b/key" || k.RelPath == "a/b/key" {
			t.Errorf("deep key should not be listed: %q", k.RelPath)
		}
	}
}
