# Patch human-only — Phase 8 Hardening

Berkas pada path berikut ditolak `deny` di `.claude/settings.json`, sehingga
agent tidak dapat menyuntingnya — termasuk dari worktree, karena `deny`
dievaluasi terhadap path relatif repo, bukan terhadap checkout.

| Path | Alasan human-only |
|---|---|
| `cmd/m2s/**` | penegak batas path; agent yang dapat menyuntingnya dapat melonggarkan batas yang mengikatnya |
| `Makefile` | menentukan bagaimana penegak dibangun |
| `.claude/**` | definisi agent dan settings |
| `.github/**` | workflow yang menjadi required check |

Setiap patch di bawah berdiri sendiri dan siap disalin. Terapkan berurutan;
`make verify` hijau setelah patch 1–4 diterapkan.

---

## 1. `cmd/m2s/commands.go` — H-03, H-05, H-06, H-07

Empat guard, semuanya fail-closed `exitViolation` (exit 2).

### 1a. Helper baru — sisipkan sebelum `// --- validate-task ---`

`pathmatch.IsAllowed` sudah menjadi sumber kebenaran overlap glob yang dipakai
`check-path`, `validate-changed-paths`, dan `reserve-paths`. Ia dipakai ulang di
sini supaya tidak ada matcher kedua yang dapat menyimpang.

```go
// scaffoldFiles mendaftar berkas yang scaffolding sebuah stack PASTI hasilkan.
//
// H-03 (phase-8-hardening.md): Phase 7 menulis contract yang melarang go.mod
// dan src/app/layout.tsx, padahal `go mod init` dan `create-next-app` wajib
// membuatnya. Agent lalu berada di antara menaati contract atau menyelesaikan
// task, dan keduanya salah. Daftar ini membuat contract semacam itu ditolak
// sebelum agent mulai.
var scaffoldFiles = map[string][]string{
	"go":     {"go.mod"},
	"nextjs": {"src/app/layout.tsx", "src/app/globals.css"},
}

// checkScaffoldRealism memeriksa paths contract terhadap execution.scaffold.
//
// Opt-in: task tanpa execution.scaffold bukan task scaffolding dan tidak
// diperiksa, sehingga task pada repo yang sudah berdiri tetap boleh melarang
// go.mod — itu batas yang benar di sana.
//
// Mengembalikan daftar pelanggaran; kosong berarti lolos.
func checkScaffoldRealism(task map[string]any) []string {
	stack := str(task, "execution.scaffold")
	if stack == "" {
		return nil
	}
	want, ok := scaffoldFiles[stack]
	if !ok {
		return nil // enum schema sudah membatasi nilainya
	}

	allowed := strSlice(task, "paths.allowed")
	forbidden := strSlice(task, "paths.forbidden")

	var out []string
	for _, f := range want {
		// IsAllowed sekaligus menutup dua bentuk kegagalan: berkas tidak
		// tercakup allowed, dan berkas tercakup allowed tetapi ditutup
		// forbidden (forbidden mengalahkan allowed, matriks §4.8).
		if !pathmatch.IsAllowed(f, allowed, forbidden) {
			out = append(out, fmt.Sprintf(
				"scaffolding %s wajib menghasilkan %s, tetapi paths contract tidak mengizinkannya — agent akan terjepit antara menaati contract dan menyelesaikan task (H-03, §29.6)",
				stack, f))
		}
	}
	return out
}

// checkContractIDsExist memastikan tiap contract_ids yang dirujuk benar-benar
// ada sebagai spec di control repository.
//
// H-05/H-06: task yang menunjuk contract hilang lolos launch, lalu agent
// bekerja tanpa dokumen yang seharusnya mengikatnya.
func checkContractIDsExist(task map[string]any, control string) []string {
	var out []string
	for _, id := range strSlice(task, "task.contract_ids") {
		p := filepath.Join(control, "control", "tasks", "specifications", id+".yaml")
		if _, err := os.Stat(p); err != nil {
			out = append(out, fmt.Sprintf(
				"contract_ids memuat %s, tetapi %s tidak ada — task tidak boleh mulai tanpa contract yang dirujuknya (H-06)",
				id, p))
		}
	}
	return out
}
```

Tambahkan import `pathmatch` pada blok import berkas itu:

```go
	"github.com/Mind2Screen-Dev-Team/m2s-vsh-platform/internal/contract"
	"github.com/Mind2Screen-Dev-Team/m2s-vsh-platform/internal/pathmatch"
	"github.com/Mind2Screen-Dev-Team/m2s-vsh-platform/internal/registry"
```

### 1b. `cmdValidateTask` — H-03 + H-05

Ganti blok penutup fungsi (mulai dari pemeriksaan `base_branch`) menjadi:

```go
	// Batas yang tidak dapat ditegakkan schema: merge ke main hanya manusia
	// (ADR-001 #2). Schema menerima base_branch main karena ia tidak tahu
	// siapa yang menjalankan merge; runner yang menolaknya.
	if base := str(doc, "ownership.base_branch"); base == "main" {
		reportViolations("task ditolak", []string{
			"ownership.base_branch = main — agent tidak boleh menargetkan main (ADR-001 #2)",
		})
		return exitViolation
	}

	// H-03 + H-05: pemeriksaan pra-launch. Contract yang sudah pasti membuat
	// task gagal ditolak SEBELUM agent mulai, bukan setelah kerja habis di CI.
	var violations []string
	violations = append(violations, checkScaffoldRealism(doc)...)
	violations = append(violations, checkContractIDsExist(doc, control)...)
	if len(violations) > 0 {
		reportViolations("task ditolak", violations)
		return exitViolation
	}

	fmt.Printf("ok  %s valid — %s pada %s\n",
		str(doc, "task.id"), str(doc, "ownership.role"), str(doc, "ownership.repository"))
	return exitOK
}
```

### 1c. `cmdLaunchTask` — H-06 + H-07 + H-05

Sisipkan setelah blok pemeriksaan platform (tepat sebelum komentar
`// Worktree hanya dibuat bila reservasi sudah ada dan masih aktif.`):

```go
	// ── Preflight Phase 8 ────────────────────────────────────────────────
	//
	// Seluruhnya mendahului reservasi dan `git worktree add`, mengikuti pola
	// pemeriksaan platform di atas: penolakan tidak boleh meninggalkan worktree
	// yatim yang harus dibersihkan manual.

	// H-07: gate TL/SA. `technical-ready` adalah status §33 yang menandai
	// analisis teknis selesai dan disetujui — sign-off yang blueprint sebut
	// `approved`. Nilai itu tidak ada pada enum taskStatus, dan menambahkannya
	// akan meriakkan perubahan ke seluruh spec, test enum, dan dokumen state
	// machine tanpa menambah penegakan.
	if st := str(task, "task.status"); st != "technical-ready" {
		return fail(exitViolation,
			"%s berstatus %s — launch menuntut technical-ready sebagai sign-off TL/SA (H-07, §33)",
			taskID, st)
	}

	// H-06: contract yang dirujuk wajib ada. Lapis kedua — validate-task
	// memeriksa hal yang sama, tetapi launch tidak boleh bergantung pada
	// pemanggil yang menjalankannya lebih dulu.
	if v := checkContractIDsExist(task, control); len(v) > 0 {
		reportViolations(fmt.Sprintf("launch %s ditolak", taskID), v)
		return exitViolation
	}

	// H-03 lapis kedua, alasan yang sama.
	if v := checkScaffoldRealism(task); len(v) > 0 {
		reportViolations(fmt.Sprintf("launch %s ditolak", taskID), v)
		return exitViolation
	}

	// H-05: base branch wajib ada di repo target. Diperiksa di sini, bukan di
	// validate-task, karena hanya launch yang menerima -repo — validate-task
	// tidak tahu di mana repository berada.
	//
	// Dijalankan runner, bukan agent, sehingga blocklist Bash agent tidak
	// berlaku — pola yang sama dipakai `git worktree add` di bawah.
	if base := str(task, "ownership.base_branch"); base != "" {
		cmd := exec.Command("git", "-C", *repoPath, "show-ref", "--verify",
			"refs/heads/"+base)
		if err := cmd.Run(); err != nil {
			return fail(exitViolation,
				"base_branch %s tidak ada di %s — worktree tidak dapat dicabangkan dari branch yang belum ada (H-05)",
				base, *repoPath)
		}
	}
```

`taskID`, `control`, `task`, dan `*repoPath` semuanya sudah ter-resolve pada
titik itu. `exec`, `fmt`, `os`, `filepath` sudah ter-import.

### 1d. `cmd/m2s/commands_test.go` — test baru

`taskOpts` dapat tiga field:

```go
type taskOpts struct {
	id       string
	repo     string
	role     string
	taskType string
	base     string
	allowed  []string
	shared   string // path shared file; kosong berarti tidak ada
	platform string // execution.platform; kosong berarti field tidak ditulis
	status   string // task.status; kosong berarti technical-ready
	scaffold string // execution.scaffold; kosong berarti field tidak ditulis
	contract string // satu contract_ids; kosong berarti field tidak ditulis
}
```

Pada `writeTask`, setelah default `o.base`:

```go
	if o.status == "" {
		o.status = "technical-ready"
	}
```

Ganti baris status dan sisipkan dua field baru:

```go
	b.WriteString("  type: " + o.taskType + "\n  project: uji\n  status: " + o.status + "\n")
	if o.contract != "" {
		b.WriteString("  contract_ids: [" + o.contract + "]\n")
	}
	b.WriteString("ownership:\n  role: " + o.role + "\n  repository: " + o.repo + "\n")
	b.WriteString("  base_branch: " + o.base + "\n")
	b.WriteString("  branch: agent/" + o.id + "-uji\n")
	b.WriteString("execution:\n  isolation: worktree\n")
	if o.platform != "" {
		b.WriteString("  platform: " + o.platform + "\n")
	}
	if o.scaffold != "" {
		b.WriteString("  scaffold: " + o.scaffold + "\n")
	}
```

Test baru, tambahkan di akhir berkas:

```go
// --- Phase 8 hardening ---

// TestCmdValidateTaskRejectsScaffoldForbidden menegakkan H-03.
//
// Contract Phase 7 melarang go.mod dan src/app/layout.tsx padahal scaffolding
// wajib membuatnya. Ditolak sebelum agent mulai, bukan setelah CI gagal.
func TestCmdValidateTaskRejectsScaffoldForbidden(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	silence(t)

	// writeTask selalu menulis go.mod pada forbidden, sehingga stack go
	// otomatis melanggar.
	goTask := writeTask(t, dir, taskOpts{id: "BE-201", scaffold: "go"})
	if code := cmdValidateTask([]string{"-control", root, "-task", goTask}); code != exitViolation {
		t.Errorf("scaffold go dengan go.mod forbidden = exit %d, mau %d (H-03)", code, exitViolation)
	}

	nextTask := writeTask(t, dir, taskOpts{
		id: "FE-201", repo: "proyek-frontend", role: "frontend-engineer",
		taskType: "frontend-implementation", scaffold: "nextjs",
		allowed: []string{"src/components/**"},
	})
	if code := cmdValidateTask([]string{"-control", root, "-task", nextTask}); code != exitViolation {
		t.Errorf("scaffold nextjs tanpa layout.tsx = exit %d, mau %d (H-03)", code, exitViolation)
	}
}

// TestCmdValidateTaskScaffoldCoveredByGlob adalah kontrol negatif H-03.
//
// Penjaga yang menolak segalanya sama tidak bergunanya dengan yang tidak
// menolak apa pun. Pola glob yang mencakup berkas scaffolding harus diterima,
// dan task tanpa field scaffold tidak boleh diperiksa sama sekali.
func TestCmdValidateTaskScaffoldCoveredByGlob(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	silence(t)

	// src/app/** mencakup layout.tsx dan globals.css.
	covered := writeTask(t, dir, taskOpts{
		id: "FE-202", repo: "proyek-frontend", role: "frontend-engineer",
		taskType: "frontend-implementation", scaffold: "nextjs",
		allowed: []string{"src/app/**"},
	})
	if code := cmdValidateTask([]string{"-control", root, "-task", covered}); code != exitOK {
		t.Errorf("scaffold nextjs dengan src/app/** = exit %d, mau %d", code, exitOK)
	}

	// Tanpa field scaffold: task pada repo yang sudah berdiri boleh melarang
	// go.mod. Ini yang membuat H-03 opt-in, bukan universal.
	optOut := writeTask(t, dir, taskOpts{id: "BE-202"})
	if code := cmdValidateTask([]string{"-control", root, "-task", optOut}); code != exitOK {
		t.Errorf("task tanpa scaffold = exit %d, mau %d — H-03 harus opt-in", code, exitOK)
	}
}

// TestCmdValidateTaskRejectsMissingContractID menegakkan H-05/H-06.
func TestCmdValidateTaskRejectsMissingContractID(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	silence(t)

	task := writeTask(t, dir, taskOpts{id: "BE-203", contract: "CONTRACT-999"})
	if code := cmdValidateTask([]string{"-control", root, "-task", task}); code != exitViolation {
		t.Errorf("contract_ids menunjuk berkas hilang = exit %d, mau %d (H-06)", code, exitViolation)
	}

	// Kontrol negatif: contract yang ada harus lolos.
	specs := filepath.Join(root, "control", "tasks", "specifications")
	if err := os.MkdirAll(specs, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(specs, "CONTRACT-102.yaml"), []byte("# uji\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok := writeTask(t, dir, taskOpts{id: "BE-204", contract: "CONTRACT-102"})
	if code := cmdValidateTask([]string{"-control", root, "-task", ok}); code != exitOK {
		t.Errorf("contract_ids yang ada = exit %d, mau %d", code, exitOK)
	}
}

// TestCmdLaunchTaskRejectsStatusNotTechnicalReady menegakkan H-07.
//
// Gate TL/SA berada di launch, bukan di validate-handoff: handoff berjalan pada
// SubagentStop, yaitu setelah kerja habis, dan payload-nya tidak memuat task
// spec.
func TestCmdLaunchTaskRejectsStatusNotTechnicalReady(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	repo := t.TempDir()
	wtRoot := t.TempDir()
	silence(t)
	t.Setenv("M2S_WORKTREE_ROOT", wtRoot)

	task := writeTask(t, dir, taskOpts{id: "BE-205", status: "draft"})
	if code := cmdReservePaths([]string{"-control", root, "-task", task}); code != exitOK {
		t.Fatalf("reservasi = exit %d", code)
	}

	code := cmdLaunchTask([]string{"-control", root, "-task", task, "-repo", repo, "-dry-run"})
	if code != exitViolation {
		t.Errorf("status draft = exit %d, mau %d (H-07)", code, exitViolation)
	}

	// Penolakan mendahului pembuatan worktree.
	entries, err := os.ReadDir(wtRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("penolakan H-07 meninggalkan %d entri pada worktree root", len(entries))
	}
}

// TestCmdLaunchTaskRejectsBaseBranchMissing menegakkan H-05.
//
// `git worktree add ... <base>` gagal dengan exitError bila base tidak ada,
// yang terbaca sebagai runner rusak. H-05 mengubahnya menjadi kontrak ditolak
// (exitViolation) dan memindahkannya ke sebelum worktree disentuh.
func TestCmdLaunchTaskRejectsBaseBranchMissing(t *testing.T) {
	root := controlFixture(t)
	dir := t.TempDir()
	repo := t.TempDir()
	silence(t)
	t.Setenv("M2S_WORKTREE_ROOT", t.TempDir())

	// Repository git nyata dengan satu commit pada main — `develop` tidak ada.
	// show-ref --verify gagal juga pada branch yang belum lahir, sehingga satu
	// commit diperlukan agar yang teruji adalah ketiadaan develop, bukan repo
	// tanpa HEAD.
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"-c", "user.email=uji@m2s.test", "-c", "user.name=uji", "commit",
			"--allow-empty", "-m", "awal"},
	} {
		cmd := exec.Command("git", append([]string{"-C", repo}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}

	task := writeTask(t, dir, taskOpts{id: "BE-206", base: "develop"})
	if code := cmdReservePaths([]string{"-control", root, "-task", task}); code != exitOK {
		t.Fatalf("reservasi = exit %d", code)
	}

	code := cmdLaunchTask([]string{"-control", root, "-task", task, "-repo", repo, "-dry-run"})
	if code != exitViolation {
		t.Errorf("base_branch develop tidak ada = exit %d, mau %d (H-05)", code, exitViolation)
	}
}
```

`commands_test.go` perlu import `os/exec` untuk test terakhir.

---

## 2. `.claude/agents/` — mirror H-04

`templates/agents/*.md` sudah memuat blok `## Architecture Constraints (wajib
baca sebelum kerja)` (13 berkas, sudah ter-commit). `TestDeployedAgentsMatchTemplates`
menuntut salinan aktif identik byte, sehingga keempat berkas yang ter-deploy
harus disalin ulang:

```bash
cd <akar-control-repo>
for r in frontend-engineer project-manager technical-lead-system-analyst technical-writer; do
  cp "templates/agents/$r.md" ".claude/agents/$r.md"
done
```

Sebelum patch ini diterapkan, `go test ./internal/contract/` merah pada
`TestDeployedAgentsMatchTemplates` — itu satu-satunya kegagalan, dan justru
bukti test itu bekerja.

---

## 3. `Makefile` — daftarkan test H-04 di `verify-agents`

Tambahkan `TestArchitectureConstraintsPresent` ke daftar `-run` pada target
`verify-agents` (baris 101). Tanpa patch ini test tetap berjalan lewat
`make test` di dalam `check`, tetapi tidak ikut terhitung pada `verify-agents`
yang justru menjadi bukti kriteria Done §57.

Ganti daftar `-run` menjadi:

```
	go test ./internal/contract/ -run 'TestEveryRoleHasAgentTemplate|TestAgentFrontmatterFieldsAreSupported|TestAgentNameMatchesFileName|TestNoAgentHasAgentTool|TestReadOnlyRolesHaveNoWriteTools|TestWriterRolesDeclareWorktreeIsolation|TestForbiddenPathBaselinePresent|TestEveryRoleHasEffort|TestArchitectureConstraintsPresent|TestDeployedAgentsMatchTemplates|TestAgentBoundariesAreDistinct' -v 2>&1 \
```

---

## 4. `.github/workflows/path-enforcement.yml` — salinan control repo

Template kanonik sudah memuat H-01, H-02, dan H-08. Salinan control repo
menerima H-01 dan H-02, **tidak** H-08.

**Mengapa H-08 dihapus di sini:** step H-08 gated pada `github.base_ref`, bukan
pada `steps.scope.outputs.mode`, sehingga ia berlaku juga pada PR manusia.
Control repo trigger-nya `branches: [main]` dan branch kerjanya `worktree-*`,
jadi H-08 apa adanya akan menolak **setiap** PR di control repo — termasuk PR
yang membawa patch ini. Hirarki promosi `develop→staging→main` hanya bermakna di
repo aplikasi.

### 4a. Header

Ganti empat baris pembuka setelah baris pertama:

```yaml
# Ini salinan CONTROL REPO. Template kanonik ada di
# `templates/github/workflows/path-enforcement.yml`; perubahan dimulai di sana,
# lalu diturunkan ke sini dan ke repo aplikasi.
# `tests/negative/github-workflow.test.sh` memeriksa seluruh salinan tetap
# sejalan pada aturan bentuk.
#
# ── Guard Phase 8 di salinan ini ─────────────────────────────────────────
#
# H-02  aktif — pola worktree-* dilaporkan sebagai mode planning.
# H-01  ADA tetapi tak pernah menyala: APPLICATION_REPO=false karena control
#       repo main-only, dan branch-nya worktree-* sehingga mode agent tidak
#       tercapai. Step-nya dipertahankan supaya bentuk antar-salinan seragam.
# H-08  SENGAJA TIDAK ADA — lihat penjelasan pada patch 4 dokumen ini.
```

### 4b. Step scope

Ganti seluruh step `Tentukan lingkup pemeriksaan` menjadi:

```yaml
      - name: Tentukan lingkup pemeriksaan
        id: scope
        env:
          EVENT_NAME: ${{ github.event_name }}
          HEAD_REF: ${{ github.head_ref }}
          BASE_REF: ${{ github.base_ref }}
          # Control repo main-only per konvensi, sehingga H-01 tidak berlaku.
          APPLICATION_REPO: 'false'
        run: |
          set -u
          if [ "$EVENT_NAME" = "merge_group" ]; then
            echo "mode=merge-group" >> "$GITHUB_OUTPUT"
            echo "task_id=" >> "$GITHUB_OUTPUT"
            exit 0
          fi

          case "$HEAD_REF" in
            agent/*) ;;
            worktree-*)
              # H-02: branch planning/dokumentasi — jalur normal di control repo.
              echo "mode=planning" >> "$GITHUB_OUTPUT"
              echo "task_id=" >> "$GITHUB_OUTPUT"
              exit 0
              ;;
            *)
              echo "mode=non-agent" >> "$GITHUB_OUTPUT"
              echo "task_id=" >> "$GITHUB_OUTPUT"
              exit 0
              ;;
          esac

          task_id=$(printf '%s' "$HEAD_REF" \
            | sed -n 's#^agent/\([A-Z][A-Z0-9]*-[0-9]\{1,\}\)-.*#\1#p')
          if [ -z "$task_id" ]; then
            echo "::error::branch '$HEAD_REF' tidak mengikuti pola agent/<task-id>-<slug> (§44)"
            echo "mode=malformed" >> "$GITHUB_OUTPUT"
            exit 1
          fi

          # H-01: tidak menyala di sini (APPLICATION_REPO=false).
          if [ "$APPLICATION_REPO" = "true" ] && [ "$BASE_REF" != "develop" ]; then
            echo "::error::branch agent/* wajib target develop (§44); staging/main via promosi berurutan, bukan lompati develop"
            echo "mode=wrong-base" >> "$GITHUB_OUTPUT"
            exit 1
          fi

          echo "mode=agent" >> "$GITHUB_OUTPUT"
          echo "task_id=$task_id" >> "$GITHUB_OUTPUT"
```

### 4c. Step Ringkasan

Tambahkan `BASE_REF` ke `env` dan dua arm baru ke `case`:

```yaml
        env:
          MODE: ${{ steps.scope.outputs.mode }}
          TASK_ID: ${{ steps.scope.outputs.task_id }}
          HEAD_REF: ${{ github.head_ref }}
          BASE_REF: ${{ github.base_ref }}
```

```yaml
          case "$MODE" in
            agent)       msg="branch agent '$HEAD_REF' (task $TASK_ID) — batas path divalidasi terhadap contract" ;;
            planning)    msg="branch '$HEAD_REF' pola worktree-* — planning/dokumentasi, tanpa contract, lolos (H-02)" ;;
            non-agent)   msg="branch '$HEAD_REF' bukan agent/* — validasi batas path tidak berlaku, lolos" ;;
            merge-group) msg="event merge_group — batas path sudah divalidasi pada PR untuk commit ini" ;;
            malformed)   msg="branch '$HEAD_REF' mengaku agent/* tetapi tidak mengikuti pola §44 — DITOLAK" ;;
            wrong-base)  msg="branch agent '$HEAD_REF' menargetkan '$BASE_REF', wajib develop (§44) — DITOLAK (H-01)" ;;
            *)           msg="lingkup tidak dapat ditentukan — perlakukan sebagai gagal" ;;
          esac
```

Verifikasi setelah menerapkan:

```bash
bash tests/lib/check-github-artifacts.sh workflow .github/workflows/path-enforcement.yml
bash tests/negative/github-workflow.test.sh
```

---

## 5. Repo aplikasi — `m2s-vsh-project-backend` dan `m2s-vsh-project-frontend`

Diterapkan pada `.github/workflows/path-enforcement.yml` di **kedua** repo, lewat
PR terpisah di masing-masing repo. Dua perubahan, keduanya perlu.

### 5a. Org control repo salah — perbaikan yang berdiri sendiri

Kedua salinan sekarang memuat:

```yaml
          repository: fajarcandraaa/m2s-vsh-platform
```

Itu control repo pra-migrasi. CI di kedua repo aplikasi karena itu me-resolve
control repo yang salah saat mencari contract. Ganti menjadi:

```yaml
          repository: Mind2Screen-Dev-Team/m2s-vsh-platform
```

Perbaikan ini tidak bergantung pada Phase 8 dan sebaiknya lebih dulu.

### 5b. Trigger dan guard

Trigger sekarang `branches: [develop, staging, main]`. `main` bertentangan dengan
H-01 — branch task tidak boleh menargetkan main sama sekali. Kembalikan ke
bentuk template:

```yaml
on:
  pull_request:
    branches: [develop, staging]
  merge_group:
```

Lalu salin **seluruh** isi `templates/github/workflows/path-enforcement.yml`
dari control repo, yang sudah memuat H-01 (`APPLICATION_REPO: 'true'`), H-02, dan
H-08 lengkap. Berbeda dengan salinan control repo, di sini H-08 **dipertahankan**:
repo aplikasi punya develop, staging, dan main, sehingga hirarki promosinya nyata.

### 5c. Verifikasi setelah diterapkan

`tests/negative/github-workflow.test.sh` di control repo akan berhenti melewatkan
kedua repo dan mulai memeriksanya:

```
  SKIP [backend] memuat org fajarcandraaa — terapkan patch ... dulu
```

berubah menjadi kasus yang lulus. Skip itu sengaja: `make verify` tetap hijau
sebelum patch diterapkan, tanpa berpura-pura kedua salinan sudah benar.

---

## Catatan: contract Phase 7 sengaja tidak diperbaiki

`control/tasks/specifications/{BE-102,BE-102-fix,FE-102,QA-102,CONTRACT-102}.yaml`
tidak valid terhadap `task.schema.json`:

| Berkas | Masalah |
|---|---|
| `BE-102`, `BE-102-fix` | `role: backend-developer` — enum `writerRole` hanya punya `backend-engineer` |
| `QA-102` | `role: quality-assurance` — enum hanya punya `qa-engineer` |
| `CONTRACT-102` | `status: approved` — bukan nilai `taskStatus` |
| keempat + FE-102 | `status: completed` — bukan nilai `taskStatus` |
| kelimanya | `base_branch: main` — ditolak `cmdValidateTask` (ADR-001 #2) |

Artinya kelima contract itu tidak pernah lolos `validate-task`. Keliru
memperbaikinya sekarang: `base_branch: main` adalah catatan bahwa Phase 7 memang
menargetkan main, dan itu justru kesalahan yang H-01 ada untuk mencegah.
Mengubahnya menjadi `develop` akan memalsukan riwayat, bukan memperbaiki apa pun.

Fixture valid yang dipakai test sudah ada di
`schemas/examples/task-BE-101.valid.yaml`.

Yang perlu diputuskan terpisah: apakah kelima berkas dipindahkan ke
`control/tasks/archive/` dengan penanda "tidak valid terhadap schema, disimpan
sebagai catatan", atau nama role/status-nya diperbaiki sambil `base_branch: main`
dipertahankan. Keduanya keputusan Project Manager, bukan bagian dari 8 guard.
