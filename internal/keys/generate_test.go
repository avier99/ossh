package keys

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate_EmptyName(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), Name: ""})
	if !errors.Is(err, ErrEmptyName) {
		t.Errorf("expected ErrEmptyName, got %v", err)
	}
}

func TestValidate_PathSeparator(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"foo/bar", `foo\bar`} {
		err := Validate(GenerateOpts{SSHDir: dir, Name: name})
		if !errors.Is(err, ErrPathSeparator) {
			t.Errorf("name %q: expected ErrPathSeparator, got %v", name, err)
		}
	}
}

func TestValidate_SpaceInName(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), Name: "my key"})
	if !errors.Is(err, ErrSpaceInName) {
		t.Errorf("expected ErrSpaceInName, got %v", err)
	}
}

func TestValidate_LeadingDot(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), Name: ".hidden"})
	if !errors.Is(err, ErrLeadingDot) {
		t.Errorf("expected ErrLeadingDot, got %v", err)
	}
}

func TestValidate_FileAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "mykey")
	if err := os.WriteFile(keyPath, []byte("fake"), 0600); err != nil {
		t.Fatal(err)
	}

	err := Validate(GenerateOpts{SSHDir: dir, Name: "mykey"})
	if err == nil {
		t.Fatal("expected error when target file exists")
	}
}

func TestValidate_PublicKeyAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	pubPath := filepath.Join(dir, "mykey.pub")
	if err := os.WriteFile(pubPath, []byte("fake"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Validate(GenerateOpts{SSHDir: dir, Name: "mykey"})
	if err == nil {
		t.Fatal("expected error when public key already exists")
	}
}

func TestValidate_Valid(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), Name: "mykey"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestValidate_SubDirWithSlash(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), SubDir: "foo/bar", Name: "mykey"})
	if !errors.Is(err, ErrInvalidSubDir) {
		t.Errorf("expected ErrInvalidSubDir, got %v", err)
	}
}

func TestValidate_SubDirAbsolute(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), SubDir: "/etc", Name: "mykey"})
	if !errors.Is(err, ErrInvalidSubDir) {
		t.Errorf("expected ErrInvalidSubDir, got %v", err)
	}
}

func TestValidate_SubDirValid(t *testing.T) {
	err := Validate(GenerateOpts{SSHDir: t.TempDir(), SubDir: "homelab-keys", Name: "mykey"})
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestCommand_ED25519Args(t *testing.T) {
	dir := t.TempDir()
	cmd, err := Command(GenerateOpts{
		SSHDir:  dir,
		Name:    "mykey",
		KeyType: KeyTypeED25519,
	})
	if err != nil {
		t.Fatal(err)
	}

	args := cmd.Args[1:]
	if !containsArgPair(args, "-t", "ed25519") {
		t.Errorf("expected -t ed25519 in args, got %v", args)
	}
	expectedPath := filepath.Join(dir, "mykey")
	if !containsArgPair(args, "-f", expectedPath) {
		t.Errorf("expected -f %s in args, got %v", expectedPath, args)
	}
	if !containsArgPair(args, "-C", "") {
		t.Errorf("expected -C \"\" in args, got %v", args)
	}
}

func TestCommand_RSA4096Args(t *testing.T) {
	dir := t.TempDir()
	cmd, err := Command(GenerateOpts{
		SSHDir:  dir,
		Name:    "mykey",
		KeyType: KeyTypeRSA4096,
	})
	if err != nil {
		t.Fatal(err)
	}

	args := cmd.Args[1:]
	if !containsArgPair(args, "-t", "rsa") {
		t.Errorf("expected -t rsa in args, got %v", args)
	}
	if !containsArgPair(args, "-b", "4096") {
		t.Errorf("expected -b 4096 in args, got %v", args)
	}
	expectedPath := filepath.Join(dir, "mykey")
	if !containsArgPair(args, "-f", expectedPath) {
		t.Errorf("expected -f %s in args, got %v", expectedPath, args)
	}
}

func TestCommand_NoPassphraseFlag(t *testing.T) {
	dir := t.TempDir()
	cmd, err := Command(GenerateOpts{
		SSHDir:  dir,
		Name:    "mykey",
		KeyType: KeyTypeED25519,
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, arg := range cmd.Args[1:] {
		if arg == "-N" {
			t.Errorf("-N flag must not be present in args: %v", cmd.Args[1:])
		}
	}
}

func TestCommand_SubDirKeyPath(t *testing.T) {
	dir := t.TempDir()
	cmd, err := Command(GenerateOpts{
		SSHDir:  dir,
		SubDir:  "homelab-keys",
		Name:    "mykey",
		KeyType: KeyTypeED25519,
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedPath := filepath.Join(dir, "homelab-keys", "mykey")
	args := cmd.Args[1:]
	if !containsArgPair(args, "-f", expectedPath) {
		t.Errorf("expected -f %s in args, got %v", expectedPath, args)
	}

	subDirPath := filepath.Join(dir, "homelab-keys")
	info, err := os.Stat(subDirPath)
	if err != nil {
		t.Fatalf("expected subdir to be created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", subDirPath)
	}
	if info.Mode().Perm() != 0700 {
		t.Errorf("expected subdir mode 0700, got %04o", info.Mode().Perm())
	}
}

func TestVerify_BothCorrect(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mykey"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mykey.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Verify(dir, "mykey"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestVerify_MissingPrivateKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mykey.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Verify(dir, "mykey")
	if err == nil {
		t.Fatal("expected error for missing private key")
	}
	if !strings.Contains(err.Error(), "private key not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerify_MissingPublicKey(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mykey"), []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}

	err := Verify(dir, "mykey")
	if err == nil {
		t.Fatal("expected error for missing public key")
	}
	if !strings.Contains(err.Error(), "public key not found") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerify_WrongPrivateKeyPerms(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "mykey"), []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "mykey.pub"), []byte("public"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Verify(dir, "mykey")
	if err == nil {
		t.Fatal("expected error for wrong private key permissions")
	}
	if !strings.Contains(err.Error(), "0600") {
		t.Errorf("unexpected error: %v", err)
	}
}

func containsArgPair(args []string, flag, value string) bool {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag && args[i+1] == value {
			return true
		}
	}
	return false
}
