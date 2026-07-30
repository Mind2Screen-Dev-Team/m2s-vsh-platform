package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// controlFixture menyiapkan control repository sementara: schemas/ disalin dari
// repo nyata, control/reservations/ kosong.
//
// Menyalin schema, bukan menirunya, agar test menguji schema yang sebenarnya
// dipakai produksi.
func controlFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	srcSchemas, err := filepath.Abs(filepath.Join("..", "..", "schemas"))
	if err != nil {
		t.Fatal(err)
	}
	dstSchemas := filepath.Join(root, "schemas")
	if err := os.MkdirAll(dstSchemas, 0o755); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(srcSchemas)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(srcSchemas, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dstSchemas, e.Name()), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "control", "reservations"), 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

type taskOpts struct {
	id       string
	repo     string
	role     string
	taskType string
	base     string
	allowed  []string
	shared   string // path shared file; kosong berarti tidak ada
}

func writeTask(t *testing.T, dir string, o taskOpts) string {
	t.Helper()
	if o.repo == "" {
		o.repo = "proyek-backend"
	}
	if o.role == "" {
		o.role = "backend-engineer"
	}
	if o.taskType == "" {
		o.taskType = "backend-implementation"
	}
	if o.base == "" {
		o.base = "develop"
	}
	if len(o.allowed) == 0 {
		o.allowed = []string{"internal/payroll/**"}
	}

	var b strings.Builder
	b.WriteString("schema_version: \"1.0\"\ntask:\n")
	b.WriteString("  id: " + o.id + "\n  title: uji\n")
	b.WriteString("  type: " + o.taskType + "\n  project: uji\n  status: technical-ready\n")
	b.WriteString("ownership:\n  role: " + o.role + "\n  repository: " + o.repo + "\n")
	b.WriteString("  base_branch: " + o.base + "\n")
	b.WriteString("  branch: agent/" + o.id + "-uji\n")
	b.WriteString("execution:\n  isolation: worktree\n  max_turns: 30\n  timeout_minutes: 45\n")
	b.WriteString("paths:\n  allowed:\n")
	for _, p := range o.allowed {
		// Path dikutip: nilai yang diawali '*' adalah alias YAML dan akan
		// gagal di-parse. Pola seperti "**" karena itu wajib berkutip.
		b.WriteString("    - \"" + p + "\"\n")
	}
	b.WriteString("  forbidden:\n    - \"go.mod\"\n    - \".claude/**\"\n    - \".task/**\"\n")
	if o.shared != "" {
		b.WriteString("shared_file_ownership:\n  - path: \"" + o.shared + "\"\n")
		b.WriteString("    owner_task_id: " + o.id + "\n    owner_role: " + o.role + "\n")
	}
	b.WriteString("acceptance_criteria:\n  - uji\nquality_gates:\n  - make test\n")
	b.WriteString("outputs:\n  - code\nstop_conditions:\n  - contract change required\n")

	path := filepath.Join(dir, o.id+".yaml")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeHandoff(t *testing.T, dir, taskID string) string {
	t.Helper()
	y := `schema_version: "1.0"
task_id: ` + taskID + `
role: backend-engineer
status: implementation-complete
summary: uji
changed_files:
  - path: internal/payroll/x.go
    purpose: uji
tests:
  executed:
    - command: make test
      result: passed
contract_deviations: []
`
	path := filepath.Join(dir, "handoff-"+taskID+".yaml")
	if err := os.WriteFile(path, []byte(y), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// silence membungkam stdout/stderr selama subcommand berjalan, agar keluaran
// test tidak tertimbun pesan runner.
func silence(t *testing.T) {
	t.Helper()
	devnull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	origOut, origErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = devnull, devnull
	t.Cleanup(func() {
		os.Stdout, os.Stderr = origOut, origErr
		devnull.Close()
	})
}

// --- validate-task ---

func TestCmdValidateTask(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	silence(t)

	valid := writeTask(t, dir, taskOpts{id: "BE-101"})
	if code := cmdValidateTask([]string{"-control", root, "-task", valid}); code != exitOK {
		t.Errorf("contract valid = exit %d, mau %d", code, exitOK)
	}

	// base_branch main hanya dapat ditolak runner, bukan schema (ADR-001 #2).
	mainBase := writeTask(t, dir, taskOpts{id: "BE-102", base: "main"})
	if code := cmdValidateTask([]string{"-control", root, "-task", mainBase}); code != exitViolation {
		t.Errorf("base_branch main = exit %d, mau %d (ADR-001 #2)", code, exitViolation)
	}

	// Pelanggaran schema.
	bad := filepath.Join(dir, "bad.yaml")
	os.WriteFile(bad, []byte("schema_version: \"1.0\"\n"), 0o644)
	if code := cmdValidateTask([]string{"-control", root, "-task", bad}); code != exitViolation {
		t.Errorf("contract tidak lengkap = exit %d, mau %d", code, exitViolation)
	}

	// Berkas tidak ada = runner gagal, BUKAN kontrak ditolak.
	if code := cmdValidateTask([]string{"-control", root, "-task", "/tidak/ada.yaml"}); code != exitError {
		t.Errorf("berkas hilang = exit %d, mau %d", code, exitError)
	}

	// Flag wajib.
	if code := cmdValidateTask([]string{"-control", root}); code != exitError {
		t.Errorf("tanpa -task = exit %d, mau %d", code, exitError)
	}
}

// --- reserve-paths ---

func TestCmdReservePaths(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	t.Setenv("M2S_WORKTREE_ROOT", filepath.Join(t.TempDir(), "wt"))
	silence(t)

	first := writeTask(t, dir, taskOpts{id: "BE-101", allowed: []string{"internal/payroll/**"}})
	if code := cmdReservePaths([]string{"-control", root, "-task", first}); code != exitOK {
		t.Fatalf("reservasi pertama = exit %d", code)
	}

	// Idempotent.
	if code := cmdReservePaths([]string{"-control", root, "-task", first}); code != exitOK {
		t.Errorf("pengulangan = exit %d, harus idempotent", code)
	}

	// Subtree bertabrakan — inti R-03.
	sub := writeTask(t, dir, taskOpts{id: "BE-102", allowed: []string{"internal/payroll/period/**"}})
	if code := cmdReservePaths([]string{"-control", root, "-task", sub}); code != exitViolation {
		t.Errorf("subtree bertabrakan = exit %d, mau %d", code, exitViolation)
	}

	// Parent juga bertabrakan.
	parent := writeTask(t, dir, taskOpts{id: "BE-103", allowed: []string{"internal/**"}})
	if code := cmdReservePaths([]string{"-control", root, "-task", parent}); code != exitViolation {
		t.Errorf("parent bertabrakan = exit %d, mau %d", code, exitViolation)
	}

	// Repository berbeda tidak bertabrakan meski path sama.
	other := writeTask(t, dir, taskOpts{
		id: "FE-101", repo: "proyek-frontend", role: "frontend-engineer",
		taskType: "frontend-implementation", allowed: []string{"internal/payroll/**"},
	})
	if code := cmdReservePaths([]string{"-control", root, "-task", other}); code != exitOK {
		t.Errorf("repo berbeda = exit %d, mau %d", code, exitOK)
	}

	// Subtree terpisah pada repo yang sama boleh — dasar repo fullstack.
	disjoint := writeTask(t, dir, taskOpts{id: "BE-104", allowed: []string{"internal/attendance/**"}})
	if code := cmdReservePaths([]string{"-control", root, "-task", disjoint}); code != exitOK {
		t.Errorf("path terpisah pada repo sama = exit %d, mau %d", code, exitOK)
	}
}

// TestCmdReservePathsSharedFile menutup R-04 pada tingkat CLI.
func TestCmdReservePathsSharedFile(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	t.Setenv("M2S_WORKTREE_ROOT", filepath.Join(t.TempDir(), "wt"))
	silence(t)

	a := writeTask(t, dir, taskOpts{
		id: "BE-201", allowed: []string{"internal/a/**"}, shared: "internal/shared/enum.go",
	})
	if code := cmdReservePaths([]string{"-control", root, "-task", a}); code != exitOK {
		t.Fatalf("reservasi pertama = exit %d", code)
	}

	// Path terpisah, tetapi mengklaim shared file yang sama dengan owner beda.
	b := writeTask(t, dir, taskOpts{
		id: "BE-202", allowed: []string{"internal/b/**"}, shared: "internal/shared/enum.go",
	})
	if code := cmdReservePaths([]string{"-control", root, "-task", b}); code != exitViolation {
		t.Errorf("shared file owner berbeda = exit %d, mau %d (§29.6)", code, exitViolation)
	}
}

// --- launch-task ---

func TestCmdLaunchTaskRequiresReservation(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	repo := t.TempDir()
	silence(t)

	task := writeTask(t, dir, taskOpts{id: "BE-101"})
	// Urutan Q13: reservasi mendahului worktree.
	code := cmdLaunchTask([]string{"-control", root, "-task", task, "-repo", repo, "-dry-run"})
	if code != exitViolation {
		t.Errorf("tanpa reservasi = exit %d, mau %d (Q13)", code, exitViolation)
	}
}

// TestCmdLaunchTaskRejectsWorktreeInsideRepo menegakkan Q8/A-01.
//
// Penjagaan ini pernah gagal-BUKA karena asimetri resolusi symlink: parent yang
// sudah ada teresolusi (`/var` → `/private/var`) sementara target yang belum ada
// tidak, sehingga keduanya dianggap berbeda pohon.
func TestCmdLaunchTaskRejectsWorktreeInsideRepo(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	repo := t.TempDir()
	silence(t)

	// Worktree diarahkan ke dalam repository, tetapi BUKAN ke
	// .claude/worktrees — pola itu sudah ditolak schema sebagai anti-pattern
	// §30. Memakai subfolder biasa memastikan yang teruji di sini adalah
	// penjagaan isInside pada runner, bukan pola schema.
	t.Setenv("M2S_WORKTREE_ROOT", filepath.Join(repo, "wt"))

	task := writeTask(t, dir, taskOpts{id: "BE-101"})
	if code := cmdReservePaths([]string{"-control", root, "-task", task}); code != exitOK {
		t.Fatalf("reservasi = exit %d", code)
	}
	code := cmdLaunchTask([]string{"-control", root, "-task", task, "-repo", repo, "-dry-run"})
	if code != exitViolation {
		t.Errorf("worktree di dalam repo = exit %d, mau %d (Q8, A-01)", code, exitViolation)
	}
}

// TestSchemaRejectsClaudeWorktreePath menegakkan lapisan pertahanan kedua:
// schema menolak pola `.claude/worktrees` yang merupakan anti-pattern §30,
// bahkan sebelum runner memeriksa apakah worktree berada di dalam repo.
//
// Kedua lapisan diperlukan karena masing-masing menangkap hal berbeda: schema
// menangkap pola yang salah di mana pun lokasinya, runner menangkap lokasi yang
// salah apa pun polanya.
func TestSchemaRejectsClaudeWorktreePath(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	silence(t)

	t.Setenv("M2S_WORKTREE_ROOT", filepath.Join(t.TempDir(), ".claude", "worktrees"))
	task := writeTask(t, dir, taskOpts{id: "BE-101"})

	if code := cmdReservePaths([]string{"-control", root, "-task", task}); code != exitViolation {
		t.Errorf(".claude/worktrees = exit %d, mau %d (anti-pattern §30)", code, exitViolation)
	}
}

// --- collect-result ---

func TestCmdCollectResult(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	t.Setenv("M2S_WORKTREE_ROOT", filepath.Join(t.TempDir(), "wt"))
	silence(t)

	task := writeTask(t, dir, taskOpts{id: "BE-101"})
	if code := cmdReservePaths([]string{"-control", root, "-task", task}); code != exitOK {
		t.Fatalf("reservasi = exit %d", code)
	}

	handoff := writeHandoff(t, dir, "BE-101")

	// Tanpa -pr: hanya memvalidasi, tidak menyentuh reservasi.
	if code := cmdCollectResult([]string{"-control", root, "-handoff", handoff}); code != exitOK {
		t.Errorf("handoff valid = exit %d", code)
	}

	// Dengan -pr: berpindah ke reserved-pending-merge, TIDAK dilepas (Q12).
	pr := "https://github.com/x/y/pull/7"
	if code := cmdCollectResult([]string{"-control", root, "-handoff", handoff, "-pr", pr}); code != exitOK {
		t.Fatalf("collect dengan -pr = exit %d", code)
	}

	// Bukti A-05 tertutup: path masih tertahan setelah PR dibuat.
	other := writeTask(t, dir, taskOpts{id: "BE-102", allowed: []string{"internal/payroll/period/**"}})
	if code := cmdReservePaths([]string{"-control", root, "-task", other}); code != exitViolation {
		t.Errorf("path harus masih tertahan saat pending-merge = exit %d, mau %d (A-05, Q12)",
			code, exitViolation)
	}

	// Handoff tidak valid ditolak.
	bad := filepath.Join(dir, "bad-handoff.yaml")
	os.WriteFile(bad, []byte("schema_version: \"1.0\"\ntask_id: BE-101\n"), 0o644)
	if code := cmdCollectResult([]string{"-control", root, "-handoff", bad}); code != exitViolation {
		t.Errorf("handoff tidak lengkap = exit %d, mau %d", code, exitViolation)
	}
}

// --- release-reservation ---

func TestCmdReleaseReservation(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	t.Setenv("M2S_WORKTREE_ROOT", filepath.Join(t.TempDir(), "wt"))
	silence(t)

	task := writeTask(t, dir, taskOpts{id: "BE-101"})
	if code := cmdReservePaths([]string{"-control", root, "-task", task}); code != exitOK {
		t.Fatalf("reservasi = exit %d", code)
	}

	// Worker tidak boleh melepas (§30).
	if code := cmdReleaseReservation([]string{"-control", root, "-task-id", "BE-101", "-by", "worker"}); code != exitViolation {
		t.Errorf("-by worker = exit %d, mau %d (§30)", code, exitViolation)
	}

	// active langsung ke released ditolak; wajib lewat pending-merge (Q12).
	if code := cmdReleaseReservation([]string{"-control", root, "-task-id", "BE-101", "-by", "runner"}); code != exitViolation {
		t.Errorf("active → released = exit %d, mau %d (Q12)", code, exitViolation)
	}

	// Lewat jalur yang benar.
	handoff := writeHandoff(t, dir, "BE-101")
	if code := cmdCollectResult([]string{"-control", root, "-handoff", handoff, "-pr", "https://github.com/x/y/pull/7"}); code != exitOK {
		t.Fatalf("collect = exit %d", code)
	}
	if code := cmdReleaseReservation([]string{"-control", root, "-task-id", "BE-101", "-by", "runner"}); code != exitOK {
		t.Errorf("release setelah pending-merge = exit %d, mau %d", code, exitOK)
	}

	// Idempotent.
	if code := cmdReleaseReservation([]string{"-control", root, "-task-id", "BE-101", "-by", "runner"}); code != exitOK {
		t.Errorf("pengulangan release = exit %d, harus idempotent", code)
	}

	// Setelah dilepas, task lain boleh memakai path itu.
	other := writeTask(t, dir, taskOpts{id: "BE-102", allowed: []string{"internal/payroll/**"}})
	if code := cmdReservePaths([]string{"-control", root, "-task", other}); code != exitOK {
		t.Errorf("path harus bebas setelah released = exit %d", code)
	}

	// Task yang tidak ada.
	if code := cmdReleaseReservation([]string{"-control", root, "-task-id", "BE-999", "-by", "runner"}); code != exitError {
		t.Errorf("task tidak ada = exit %d, mau %d", code, exitError)
	}
}
