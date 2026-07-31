package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fajarcandraaa/m2s-vsh-platform/internal/contract"
	"github.com/fajarcandraaa/m2s-vsh-platform/internal/registry"
)

const lockTimeout = 30 * time.Second

// setup memuat validator dan registry dari akar control repository.
func setup(control string) (*contract.Validator, *registry.Registry, error) {
	v, err := contract.NewValidator(filepath.Join(control, "schemas"))
	if err != nil {
		return nil, nil, err
	}
	reg, err := registry.Open(filepath.Join(control, "control", "reservations"), v)
	if err != nil {
		return nil, nil, err
	}
	return v, reg, nil
}

// --- validate-task ---

func cmdValidateTask(args []string) int {
	fs := newFlagSet("validate-task")
	taskPath := fs.String("task", "", "path task contract (.yaml)")
	controlFlag := fs.String("control", "", "akar control repository")
	if err := fs.Parse(args); err != nil {
		return exitError
	}
	if *taskPath == "" {
		return fail(exitError, "-task wajib diisi")
	}

	control, err := controlRoot(*controlFlag)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	v, _, err := setup(control)
	if err != nil {
		return fail(exitError, "%v", err)
	}

	doc, err := v.Load(*taskPath, contract.KindTask)
	if err != nil {
		if ve, ok := err.(*contract.ValidationError); ok {
			reportViolations(ve.Error(), ve.Violations)
			return exitViolation
		}
		return fail(exitError, "%v", err)
	}

	// Batas yang tidak dapat ditegakkan schema: merge ke main hanya manusia
	// (ADR-001 #2). Schema menerima base_branch main karena ia tidak tahu
	// siapa yang menjalankan merge; runner yang menolaknya.
	if base := str(doc, "ownership.base_branch"); base == "main" {
		reportViolations("task ditolak", []string{
			"ownership.base_branch = main — agent tidak boleh menargetkan main (ADR-001 #2)",
		})
		return exitViolation
	}

	fmt.Printf("ok  %s valid — %s pada %s\n",
		str(doc, "task.id"), str(doc, "ownership.role"), str(doc, "ownership.repository"))
	return exitOK
}

// --- reserve-paths ---

func cmdReservePaths(args []string) int {
	fs := newFlagSet("reserve-paths")
	taskPath := fs.String("task", "", "path task contract (.yaml)")
	controlFlag := fs.String("control", "", "akar control repository")
	if err := fs.Parse(args); err != nil {
		return exitError
	}
	if *taskPath == "" {
		return fail(exitError, "-task wajib diisi")
	}

	control, err := controlRoot(*controlFlag)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	v, reg, err := setup(control)
	if err != nil {
		return fail(exitError, "%v", err)
	}

	task, err := v.Load(*taskPath, contract.KindTask)
	if err != nil {
		if ve, ok := err.(*contract.ValidationError); ok {
			reportViolations(ve.Error(), ve.Violations)
			return exitViolation
		}
		return fail(exitError, "%v", err)
	}

	taskID := str(task, "task.id")
	repo := str(task, "ownership.repository")
	allowed := strSlice(task, "paths.allowed")

	wtRoot, err := worktreeRoot()
	if err != nil {
		return fail(exitError, "%v", err)
	}

	// Kunci menahan seluruh siklus periksa-lalu-tulis. Tanpa ini dua runner
	// dapat memeriksa konflik terhadap keadaan yang sama lalu sama-sama
	// menulis reservasi yang beririsan.
	lock, err := reg.Acquire("reserve-paths "+taskID, lockTimeout)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	defer lock.Release()

	sharedOwners := map[string]string{}
	if raw, ok := task["shared_file_ownership"].([]any); ok {
		for _, e := range raw {
			m, ok := e.(map[string]any)
			if !ok {
				continue
			}
			p, _ := m["path"].(string)
			owner, _ := m["owner_task_id"].(string)
			if p != "" && owner != "" {
				sharedOwners[p] = owner
			}
		}
	}

	conflicts, err := reg.CheckConflicts(taskID, repo, allowed, sharedOwners)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	if len(conflicts) > 0 {
		msgs := make([]string, len(conflicts))
		for i, c := range conflicts {
			msgs[i] = c.Error()
		}
		reportViolations(fmt.Sprintf("reservasi %s ditolak", taskID), msgs)
		return exitViolation
	}

	doc := map[string]any{
		"schema_version":  "1.0",
		"task_id":         taskID,
		"repository":      repo,
		"branch":          str(task, "ownership.branch"),
		"worktree":        filepath.Join(wtRoot, repo, taskID),
		"allowed_paths":   toAny(allowed),
		"reserved_paths":  toAny(allowed),
		"forbidden_paths": toAny(strSlice(task, "paths.forbidden")),
		"status":          registry.StatusActive,
		"owner_role":      str(task, "ownership.role"),
		"created_at":      time.Now().Format(time.RFC3339),
	}
	if raw, ok := task["shared_file_ownership"]; ok {
		doc["shared_file_ownership"] = raw
	}

	// Idempotent: reservasi aktif milik task yang sama dipertahankan apa
	// adanya agar created_at tidak bergeser saat perintah diulang.
	if existing, gerr := reg.Get(taskID); gerr == nil && existing.Status() == registry.StatusActive {
		fmt.Printf("ok  %s sudah direservasi (tidak ada perubahan)\n", taskID)
		return exitOK
	}

	if err := reg.Put(doc); err != nil {
		if ve, ok := err.(*contract.ValidationError); ok {
			reportViolations(ve.Error(), ve.Violations)
			return exitViolation
		}
		return fail(exitError, "%v", err)
	}

	fmt.Printf("ok  %s direservasi — %d path pada %s\n", taskID, len(allowed), repo)
	return exitOK
}

// --- launch-task ---

func cmdLaunchTask(args []string) int {
	fs := newFlagSet("launch-task")
	taskPath := fs.String("task", "", "path task contract (.yaml)")
	repoPath := fs.String("repo", "", "path repository target")
	controlFlag := fs.String("control", "", "akar control repository")
	dryRun := fs.Bool("dry-run", false, "siapkan worktree tanpa menjalankan agent")
	if err := fs.Parse(args); err != nil {
		return exitError
	}
	if *taskPath == "" || *repoPath == "" {
		return fail(exitError, "-task dan -repo wajib diisi")
	}

	control, err := controlRoot(*controlFlag)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	v, reg, err := setup(control)
	if err != nil {
		return fail(exitError, "%v", err)
	}

	task, err := v.Load(*taskPath, contract.KindTask)
	if err != nil {
		if ve, ok := err.(*contract.ValidationError); ok {
			reportViolations(ve.Error(), ve.Violations)
			return exitViolation
		}
		return fail(exitError, "%v", err)
	}
	taskID := str(task, "task.id")

	// Prasyarat platform diperiksa sebelum reservasi dan worktree disentuh
	// (ADR-006 #3). Menolak lebih awal berarti tidak ada worktree yatim yang
	// harus dibersihkan manual ketika runner-nya memang salah mesin.
	//
	// Ini kontrak yang ditolak, bukan runner yang rusak — karena itu
	// exitViolation, bukan exitError.
	if want := str(task, "execution.platform"); want != "" && want != "any" {
		if want != runtime.GOOS {
			return fail(exitViolation,
				"%s menuntut platform %s, sedangkan runner berjalan pada %s",
				taskID, want, runtime.GOOS)
		}
	}

	// Worktree hanya dibuat bila reservasi sudah ada dan masih aktif.
	// Urutan ini berasal dari Q13: validasi contract dan reservasi mendahului
	// git worktree add.
	res, err := reg.Get(taskID)
	if err != nil {
		return fail(exitViolation, "%s belum memiliki reservasi — jalankan reserve-paths lebih dulu", taskID)
	}
	if res.Status() != registry.StatusActive {
		return fail(exitViolation, "reservasi %s berstatus %s, bukan active", taskID, res.Status())
	}

	worktree := res.Worktree()
	branch := str(task, "ownership.branch")
	base := str(task, "ownership.base_branch")

	// Jaminan Q8/A-01 yang tidak dapat diungkapkan schema: worktree harus
	// berada DI LUAR repository. Schema tidak mengetahui lokasi repository,
	// sehingga pemeriksaan ini hanya mungkin di sini.
	//
	// Worktree di dalam repo membuat akar kerja agent berada pada pohon yang
	// forbidden_paths-nya berlaku, dan berisiko ikut ter-commit.
	if inside, err := isInside(worktree, *repoPath); err != nil {
		return fail(exitError, "%v", err)
	} else if inside {
		return fail(exitViolation,
			"worktree %s berada di dalam repository %s — Q8 menuntut worktree di luar repo (A-01)",
			worktree, *repoPath)
	}

	// Runner yang menjalankan git worktree, bukan agent (Q13). Hook agent
	// tidak berlaku pada proses ini karena ia berada di luar sesi agent.
	if _, err := os.Stat(worktree); os.IsNotExist(err) {
		cmd := exec.Command("git", "-C", *repoPath, "worktree", "add", worktree, "-b", branch, base)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fail(exitError, "git worktree add gagal: %v\n%s", err, out)
		}
		fmt.Printf("ok  worktree dibuat: %s (branch %s dari %s)\n", worktree, branch, base)
	} else {
		fmt.Printf("ok  worktree sudah ada: %s\n", worktree)
	}

	// Materialisasi contract sebagai snapshot read-only (Q15). Hook membaca
	// berkas lokal ini, bukan registry lintas repository.
	dest := filepath.Join(worktree, ".task", "contract.json")
	if err := contract.MaterializeJSON(task, dest); err != nil {
		return fail(exitError, "%v", err)
	}
	fmt.Printf("ok  contract dimaterialisasi: %s\n", dest)

	// Snapshot berada di dalam worktree, sehingga `git add -A` milik agent
	// akan ikut men-stage-nya. Itu bertentangan dengan .task/** yang justru
	// ada pada forbidden_paths.
	//
	// Pengabaian ditulis ke .git/info/exclude, bukan .gitignore: berkas itu
	// milik worktree dan tidak pernah ter-commit, sehingga repo aplikasi tidak
	// perlu mengetahui detail runner.
	if err := excludeTaskDir(worktree); err != nil {
		return fail(exitError, "%v", err)
	}

	if *dryRun {
		fmt.Printf("ok  dry-run — sesi agent tidak dijalankan\n")
		return exitOK
	}

	fmt.Printf("siap  jalankan agent %s dengan cwd %s\n", str(task, "ownership.role"), worktree)
	return exitOK
}

// --- collect-result ---

func cmdCollectResult(args []string) int {
	fs := newFlagSet("collect-result")
	handoffPath := fs.String("handoff", "", "path handoff (.yaml)")
	controlFlag := fs.String("control", "", "akar control repository")
	prURL := fs.String("pr", "", "URL pull request; memindahkan reservasi ke reserved-pending-merge")
	if err := fs.Parse(args); err != nil {
		return exitError
	}
	if *handoffPath == "" {
		return fail(exitError, "-handoff wajib diisi")
	}

	control, err := controlRoot(*controlFlag)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	v, reg, err := setup(control)
	if err != nil {
		return fail(exitError, "%v", err)
	}

	doc, err := v.Load(*handoffPath, contract.KindHandoff)
	if err != nil {
		if ve, ok := err.(*contract.ValidationError); ok {
			reportViolations(ve.Error(), ve.Violations)
			return exitViolation
		}
		return fail(exitError, "%v", err)
	}

	taskID, _ := doc["task_id"].(string)
	status, _ := doc["status"].(string)
	fmt.Printf("ok  handoff %s valid — status %s\n", taskID, status)

	if *prURL == "" {
		return exitOK
	}

	// Reservasi TIDAK dilepas di sini. Q12: ia ditahan sampai merge, hanya
	// berpindah ke reserved-pending-merge. Melepasnya sekarang membuka
	// kembali celah A-05.
	lock, err := reg.Acquire("collect-result "+taskID, lockTimeout)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	defer lock.Release()

	if err := reg.Transition(taskID, registry.StatusReservedPendingMerge, map[string]any{
		"pr_url": *prURL,
	}); err != nil {
		return fail(exitError, "%v", err)
	}
	fmt.Printf("ok  reservasi %s → reserved-pending-merge (masih menahan path sampai merge)\n", taskID)
	return exitOK
}

// --- release-reservation ---

func cmdReleaseReservation(args []string) int {
	fs := newFlagSet("release-reservation")
	taskID := fs.String("task-id", "", "task ID pemilik reservasi")
	controlFlag := fs.String("control", "", "akar control repository")
	by := fs.String("by", "runner", "pelepas: runner atau human")
	cancel := fs.Bool("cancel", false, "batalkan alih-alih melepas setelah merge")
	if err := fs.Parse(args); err != nil {
		return exitError
	}
	if *taskID == "" {
		return fail(exitError, "-task-id wajib diisi")
	}
	if *by != "runner" && *by != "human" {
		return fail(exitViolation, "-by harus runner atau human — worker tidak boleh melepas reservasi (§30)")
	}

	control, err := controlRoot(*controlFlag)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	_, reg, err := setup(control)
	if err != nil {
		return fail(exitError, "%v", err)
	}

	lock, err := reg.Acquire("release-reservation "+*taskID, lockTimeout)
	if err != nil {
		return fail(exitError, "%v", err)
	}
	defer lock.Release()

	res, err := reg.Get(*taskID)
	if err != nil {
		return fail(exitError, "%v", err)
	}

	// Idempotent: reservasi yang sudah terlepas tidak dianggap kesalahan.
	if s := res.Status(); s == registry.StatusReleased || s == registry.StatusCancelled {
		fmt.Printf("ok  reservasi %s sudah %s (tidak ada perubahan)\n", *taskID, s)
		return exitOK
	}

	target := registry.StatusReleased
	if *cancel {
		target = registry.StatusCancelled
	}

	if err := reg.Transition(*taskID, target, map[string]any{"released_by": *by}); err != nil {
		return fail(exitViolation, "%v", err)
	}
	fmt.Printf("ok  reservasi %s → %s oleh %s\n", *taskID, target, *by)
	return exitOK
}

// excludeTaskDir memastikan .task/ tidak pernah masuk staging area agent.
//
// Ditulis ke .git/info/exclude milik worktree — bukan .gitignore — karena
// berkas itu tidak ter-commit, sehingga repository aplikasi tidak perlu memuat
// detail runner. Idempotent: pemanggilan berulang tidak menduplikasi entri.
//
// Pada worktree, .git adalah BERKAS berisi "gitdir: <path>", bukan direktori;
// lokasi info/exclude karena itu ditanyakan kepada git.
func excludeTaskDir(worktree string) error {
	out, err := exec.Command("git", "-C", worktree, "rev-parse", "--git-path", "info/exclude").Output()
	if err != nil {
		return fmt.Errorf("menentukan lokasi info/exclude: %w", err)
	}
	excludePath := strings.TrimSpace(string(out))
	if !filepath.IsAbs(excludePath) {
		excludePath = filepath.Join(worktree, excludePath)
	}

	const entry = ".task/"
	if b, err := os.ReadFile(excludePath); err == nil {
		for _, line := range strings.Split(string(b), "\n") {
			if strings.TrimSpace(line) == entry {
				return nil // sudah ada
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("membaca %s: %w", excludePath, err)
	}

	if err := os.MkdirAll(filepath.Dir(excludePath), 0o755); err != nil {
		return fmt.Errorf("membuat direktori %s: %w", filepath.Dir(excludePath), err)
	}
	f, err := os.OpenFile(excludePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("membuka %s: %w", excludePath, err)
	}
	defer f.Close()
	if _, err := f.WriteString("\n# snapshot task contract milik runner (Q15)\n" + entry + "\n"); err != nil {
		return fmt.Errorf("menulis %s: %w", excludePath, err)
	}
	return nil
}

func toAny(ss []string) []any {
	out := make([]any, len(ss))
	for i, s := range ss {
		out[i] = s
	}
	return out
}

// isInside melaporkan apakah target berada di dalam pohon direktori parent.
//
// Kedua path di-resolve lewat EvalSymlinks sehingga symlink tidak dapat dipakai
// menyelundupkan worktree ke dalam repository.
//
// Target biasanya BELUM ADA saat pemeriksaan — worktree dibuat setelahnya.
// EvalSymlinks gagal pada path yang belum ada, karena itu resolusi dilakukan
// pada leluhur terdekat yang sudah ada lalu sisanya disambung kembali.
//
// Tanpa penyambungan itu, penjagaan ini gagal-BUKA pada macOS: `/var` adalah
// symlink ke `/private/var`, sehingga parent yang ada teresolusi menjadi
// `/private/var/...` sementara target yang belum ada tetap `/var/...`. Keduanya
// lalu dianggap berada pada pohon berbeda dan worktree di dalam repo lolos.
// Ditangkap TestIsInside.
func isInside(target, parent string) (bool, error) {
	t, err := resolveEventual(target)
	if err != nil {
		return false, err
	}
	p, err := resolveEventual(parent)
	if err != nil {
		return false, err
	}

	rel, err := filepath.Rel(p, t)
	if err != nil {
		return false, nil // pada volume berbeda: pasti di luar
	}
	if rel == "." || rel == ".." {
		return false, nil
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator)), nil
}

// resolveEventual mengembalikan bentuk path yang seluruh symlink-nya sudah
// di-resolve, termasuk untuk path yang belum ada.
func resolveEventual(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", p, err)
	}

	// Cari leluhur terdekat yang sudah ada, kumpulkan sisa segmennya.
	remaining := ""
	cur := abs
	for {
		if real, err := filepath.EvalSymlinks(cur); err == nil {
			if remaining == "" {
				return real, nil
			}
			return filepath.Join(real, remaining), nil
		}
		next := filepath.Dir(cur)
		if next == cur {
			// Mencapai akar tanpa menemukan yang ada; pakai bentuk absolut.
			return abs, nil
		}
		remaining = filepath.Join(filepath.Base(cur), remaining)
		cur = next
	}
}
