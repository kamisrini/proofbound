package twin

import (
	"context"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fixturePostgresArchiveBase64 = "/Td6WFoAAATm1rRGAgAhARYAAAB0L+Wj4Cf/AJxdADEaShsJm5rsOLhmgD4g+GJazq7I5TRZC7YiAnxPCTsZJP/70wuC3Bzcd5GWSmXdzdhh2sI7st9sergfz5o1vca6TMD0Q9oMiYs82EcewV+riQQw4rpPpzWdEH5g85JyiT2F6nHRtZ85OoclZBI1R07Csijd1FG4JtY+6S+0hHD5ycEn3mhI65QVzVzeDr1oO5XAQJ029AERrWIvAABYNfgA8/HhTAABuAGAUAAA/vRJibHEZ/sCAAAAAARZWg=="

func TestStageVerifiedPostgresBindsArchiveInventoryAndExecutable(t *testing.T) {
	input := fixtureVerifiedPostgresInput(t)
	root := t.TempDir()
	binaries, err := stageVerifiedPostgres(context.Background(), root, input)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"bin/pg_ctl", "bin/postgres"} {
		pathName := filepath.Join(binaries, filepath.FromSlash(name))
		info, err := os.Stat(pathName)
		if err != nil {
			t.Fatal(err)
		}
		if !info.Mode().IsRegular() || info.Mode().Perm() != 0o755 {
			t.Fatalf("%s mode=%v", name, info.Mode())
		}
	}
	if got, err := os.ReadFile(filepath.Join(binaries, "bin", "postgres")); err != nil || string(got) != "postgres-fixture" {
		t.Fatalf("staged postgres bytes=%q error=%v", got, err)
	}
}

func TestStageVerifiedPostgresRejectsDigestAndInventoryFailures(t *testing.T) {
	tests := map[string]func(*VerifiedPostgresInput){
		"archive digest":              func(input *VerifiedPostgresInput) { input.ArchiveSHA256 = strings.Repeat("0", 64) },
		"unsafe inventory path":       func(input *VerifiedPostgresInput) { input.Inventory[0].Path = "../pg_ctl" },
		"wrong executable digest":     func(input *VerifiedPostgresInput) { input.PostgresBinarySHA256 = strings.Repeat("0", 64) },
		"missing required executable": func(input *VerifiedPostgresInput) { input.Inventory = input.Inventory[:1] },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			input := fixtureVerifiedPostgresInput(t)
			mutate(&input)
			_, err := stageVerifiedPostgres(context.Background(), t.TempDir(), input)
			if err == nil {
				t.Fatal("expected verified archive failure")
			}
		})
	}
}

func TestValidateVerifiedPostgresInputRejectsInventoryShape(t *testing.T) {
	tests := map[string]func(*VerifiedPostgresInput){
		"unsorted inventory": func(input *VerifiedPostgresInput) {
			input.Inventory[0], input.Inventory[1] = input.Inventory[1], input.Inventory[0]
		},
		"duplicate inventory": func(input *VerifiedPostgresInput) {
			input.Inventory[1].Path = input.Inventory[0].Path
		},
		"unsupported mode": func(input *VerifiedPostgresInput) {
			input.Inventory[0].Mode = "100600"
		},
		"non-executable pg_ctl": func(input *VerifiedPostgresInput) {
			input.Inventory[0].Mode = "100644"
		},
		"missing pg_ctl": func(input *VerifiedPostgresInput) {
			input.Inventory = input.Inventory[1:]
		},
		"missing postgres": func(input *VerifiedPostgresInput) {
			input.Inventory = input.Inventory[:1]
		},
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			input := fixtureVerifiedPostgresInput(t)
			mutate(&input)
			if err := validateVerifiedPostgresInput(input); err == nil {
				t.Fatal("expected inventory validation failure")
			}
		})
	}
}

func TestNormalizeArchivePathRejectsUnsafeForms(t *testing.T) {
	for _, test := range []struct {
		name string
		path string
	}{
		{name: "empty", path: ""},
		{name: "absolute", path: "/absolute"},
		{name: "parent", path: "../parent"},
		{name: "escape", path: "nested/../../escape"},
		{name: "windows", path: "windows\\path"},
	} {
		t.Run(test.name, func(t *testing.T) {
			name := test.path
			if _, err := normalizeArchivePath(name); err == nil {
				t.Fatalf("expected unsafe archive path rejection for %q", name)
			}
		})
	}
}

func TestStageVerifiedPostgresHonorsCancellationBeforeExtraction(t *testing.T) {
	input := fixtureVerifiedPostgresInput(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	root := t.TempDir()
	if _, err := stageVerifiedPostgres(ctx, root, input); err == nil {
		t.Fatal("expected canceled staging to fail")
	} else if !strings.Contains(err.Error(), context.Canceled.Error()) {
		t.Fatalf("error=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "pgbin")); !os.IsNotExist(err) {
		t.Fatalf("canceled staging created binaries directory: %v", err)
	}
}

func TestVerifyStagedPostgresRejectsUnexpectedFileAndModeDrift(t *testing.T) {
	t.Run("unexpected file", func(t *testing.T) {
		input := fixtureVerifiedPostgresInput(t)
		binaries, err := stageVerifiedPostgres(context.Background(), t.TempDir(), input)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(binaries, "unexpected"), []byte("fixture"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := verifyStagedPostgres(binaries, input); err == nil {
			t.Fatal("expected unexpected-file rejection")
		}
	})

	t.Run("mode drift", func(t *testing.T) {
		input := fixtureVerifiedPostgresInput(t)
		binaries, err := stageVerifiedPostgres(context.Background(), t.TempDir(), input)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(filepath.Join(binaries, "bin", "pg_ctl"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := verifyStagedPostgres(binaries, input); err == nil {
			t.Fatal("expected mode-drift rejection")
		}
	})
}

func TestReplayIsolatedWithVerifiedPostgresValidatesBeforeStore(t *testing.T) {
	_, err := ReplayIsolatedWithVerifiedPostgres(context.Background(), nil, Request{}, emptySnapshot(), VerifiedPostgresInput{})
	if err == nil || !strings.Contains(err.Error(), "archive is empty") {
		t.Fatalf("error=%v", err)
	}
}

func fixtureVerifiedPostgresInput(t *testing.T) VerifiedPostgresInput {
	t.Helper()
	archive, err := base64.StdEncoding.DecodeString(fixturePostgresArchiveBase64)
	if err != nil {
		t.Fatal(err)
	}
	return VerifiedPostgresInput{
		ArchiveBytes:         archive,
		ArchiveSHA256:        "3e21b7f6f29756b3c8ace3aa1f447a9b3c4407c12b8830cab63c2c0edd132118",
		PostgresBinarySHA256: "45da691608910bb2fa61f523292200597314b6fa87db5144aa5df372feecadb7",
		Inventory: []VerifiedArchiveEntry{
			{Path: "bin/pg_ctl", Mode: "100755", SHA256: "182bbdd773e0a9ad0f1498e2f3b3b325c36ef1f31c0d71effa5e9c60218b590b"},
			{Path: "bin/postgres", Mode: "100755", SHA256: "45da691608910bb2fa61f523292200597314b6fa87db5144aa5df372feecadb7"},
		},
	}
}
