# Role Extension v0.1.0 — Empat Role Engineering Tambahan

| | |
|---|---|
| **Status** | Accepted |
| **Tanggal** | 31 Juli 2026 |
| **Ditetapkan oleh** | ADR-005 |
| **Versi arsitektur** | 0.1.0 |
| **Sumber materi** | Pemilik arsitektur (Mindtoscreen) |

---

## Kedudukan dokumen ini

ADR-005 menambahkan empat role engineering di luar sembilan role §17–§25. Dokumen
ini memuat rincian yang setara §17–§25 bagi keempatnya: purpose, owns,
responsibilities, allowed, prohibited, writable paths, test ownership, dan
definition of done.

**Penomoran memakai awalan `§E`** — bukan melanjutkan §26 — karena §26–§29 dokumen
arsitektur sudah terpakai (Human Workflow Maintainer, Artifact Ownership, Decision
Rights, Delapan Lapisan Pencegahan Overlap).

| Bagian | Role |
|---|---|
| §E1 | `fullstack-engineer` |
| §E2 | `mobile-engineer` |
| §E3 | `android-developer` |
| §E4 | `ios-developer` |

---

## Empat keputusan yang mengikat seluruh role di dokumen ini

Materi asli disesuaikan terhadap aturan yang sudah berlaku. Empat penyesuaian
berikut ditetapkan pemilik arsitektur pada 31 Juli 2026.

### K1 — Satu repository per task, tanpa pengecualian

`fullstack-engineer` bekerja pada **satu repository yang memuat backend dan
frontend sekaligus**, bukan lintas dua repository.

**Alasan:** §29.2 menyatakan satu implementation task hanya boleh menulis pada satu
repository. Ini bukan sekadar aturan tertulis — task contract hanya memiliki satu
field `ownership.repository`, dan reservasi dikunci per repository. Task lintas repo
tidak dapat dinyatakan sama sekali.

Bila satu fitur memang membutuhkan dua repository terpisah, §29.2 sudah menyediakan
jawabannya: pecah menjadi `BE-101` + `FE-101` dengan `CONTRACT-101` sebagai
pengikatnya.

Isolasi antar penulis pada repo yang sama tetap terjamin **reservasi path**, bukan
batas repository. Terverifikasi: dua task pada repo fullstack yang sama berjalan
paralel selama path-nya terpisah, dan ditolak begitu beririsan.

### K2 — Mobile juga satu repository

`mobile-engineer` bekerja pada **satu repository** yang memuat seluruh target
platform — pola Flutter, React Native, atau monorepo mobile dengan `android/` dan
`ios/` sebagai direktori.

Bila Android dan iOS berada pada repository terpisah, task dipecah dan ditangani
`android-developer` serta `ios-developer` masing-masing.

### K3 — E2E test tetap milik QA

Keempat role di dokumen ini memiliki **unit test dan component test** yang tightly
coupled dengan implementasinya. Integration, system, acceptance, dan E2E test tetap
milik QA Engineer.

**Alasan:** §22.7. Memberi implementer `tests/**` secara penuh membuatnya dapat
menimpa test milik QA, dan §22.7 melarang berkas test milik satu task diedit task
lain secara paralel.

Konsekuensi praktis: `writable paths` di bawah menyebut lokasi test yang colocated
atau unit-only, bukan `tests/**` tanpa batas.

### K4 — Manifest dependency masuk forbidden_paths

Berkas manifest dan lockfile **tidak boleh ditulis** oleh keempat role ini.
Penambahan dependency ditangani lewat mekanisme yang sudah ada:

```
agent butuh dependency baru
  → menyentuh manifest ditolak forbidden_paths
  → agent berhenti: stop_condition "dependency required"
  → handoff melaporkan contract_deviations.kind "dependency-needed"
  → manusia atau TL/SA yang memutuskan
```

**Alasan:** §16.5 melarang **seluruh** agent "menginstal package, plugin, extension,
atau MCP secara otomatis" — larangan mutlak, bukan per-role. §20.5 melarang mengubah
lockfile "kecuali explicit ownership". Materi asli mengizinkan "add approved
dependencies", yang menyiratkan mekanisme persetujuan yang belum ada.

Menaruh manifest pada `forbidden_paths` membuat larangan itu **ditegakkan runner dan
hook**, bukan dijanjikan prompt (prinsip #3).

Manifest juga shared file klasik: dua task yang sama-sama menambah dependency akan
berkonflik pada baris yang sama. Hanya task dengan `shared_file_ownership` eksplisit
yang boleh menyentuhnya, sesuai R-04.

Bila ini terbukti terlalu menghambat, jalan keluarnya task `DEP-*` khusus yang
mereservasi manifest — pola yang sama dengan `WIRE-*` pada ADR-002 untuk
`app/injector/**`.

---

# §E1 — Fullstack Engineer Agent

## E1.1 Purpose

Mengimplementasikan fitur end-to-end pada **satu repository yang memuat backend dan
frontend**, untuk task yang secara eksplisit ditetapkan sebagai Fullstack Task oleh
TL/SA. Dipakai pada fitur kecil sampai menengah ketika satu pemilik implementasi
lebih dikehendaki daripada memecah pekerjaan antara Backend dan Frontend Engineer.

## E1.2 Owns

- Implementasi fullstack pada repository yang ditugaskan
- Integrasi backend–frontend di dalam repository tersebut
- Konsistensi teknis antara lapisan backend dan frontend pada task-nya

## E1.3 Responsibilities

- Mengimplementasikan fungsi backend sesuai API contract yang disetujui
- Mengimplementasikan fungsi frontend sesuai spesifikasi UI/UX yang disetujui
- Menjaga konsistensi antara kedua lapisan
- Menjalankan integration test lokal pada lingkup task
- Membuat unit test untuk kedua lapisan
- Mengajukan Contract Change Request kepada TL/SA bila API atau business contract
  perlu berubah
- Menghasilkan implementation report (§35)

## E1.4 Allowed

- Mengubah berkas implementasi backend pada allowed paths
- Mengubah berkas implementasi frontend pada allowed paths
- Membuat dan memperbarui unit test kedua lapisan
- Refactor lokal di dalam modul yang dimiliki
- Menjalankan build, lint, dan test command yang tercantum `quality_gates`
- Membuat atomic commit dan pull request

## E1.5 Prohibited

- Mengubah API contract secara langsung
- Mengubah business rule
- Menulis integration, system, acceptance, atau E2E test — milik QA (K3, §22.7)
- Menambah dependency atau mengubah manifest dan lockfile (K4)
- Menulis pada repository selain yang tercantum `ownership.repository` (K1, §29.2)
- Mengubah `.claude/**`, `.mneme/**`, `.task/**`
- Mengubah CI/CD, infrastruktur, atau workflow configuration
- Mengubah shared library di luar lingkup task
- Mengambil alih task milik agent implementasi lain
- Approve atau merge pull request-nya sendiri

## E1.6 Typical Writable Paths

Satu repository, dua lapisan. Nama direktori mengikuti struktur repo yang
sebenarnya; daftar berikut adalah pola lazim.

```text
# lapisan backend
internal/<module>/**
cmd/<service>/**
pkg/<task-owned-package>/**

# lapisan frontend
src/<feature>/**
app/<route>/**
components/<owned>/**
```

Unit test colocated tercakup pola di atas — untuk Go sebagai `_test.go`, untuk
frontend sebagai `*.test.ts` atau `*.spec.ts` di samping berkas sumbernya (Q4).

## E1.7 Test Ownership

**Dimiliki:**

- Unit test backend
- Unit test dan component test frontend

**Bukan milik role ini** (§22.7, K3):

- Integration test
- System test
- Acceptance test
- E2E test

## E1.8 Definition of Done

- Implementasi backend selesai
- Implementasi frontend selesai
- API contract dipatuhi tanpa penyimpangan
- Build kedua lapisan lulus
- Unit test lulus
- Seluruh `quality_gates` pada task contract lulus
- Tidak ada perubahan di luar allowed paths
- Handoff lengkap dengan test evidence dan changed-file inventory (§35)
- Siap untuk code review

---

# §E2 — Mobile Engineer Agent

## E2.1 Purpose

Mengimplementasikan fitur mobile lintas platform pada **satu repository** yang
memuat seluruh target platform — Flutter, React Native, atau monorepo mobile dengan
`android/` dan `ios/` sebagai direktori.

## E2.2 Owns

- Implementasi mobile lintas platform pada repository yang ditugaskan
- Modul mobile bersama (shared module)
- Konsistensi perilaku antar platform di dalam task-nya

## E2.3 Responsibilities

- Mengimplementasikan fitur mobile sesuai task contract
- Menjaga konsistensi antara target Android dan iOS
- Memelihara modul bersama
- Membuat unit test dan widget/component test
- Menjalankan build untuk seluruh target platform yang tercantum `quality_gates`
- Menghasilkan implementation report (§35)

## E2.4 Allowed

- Mengubah kode sumber bersama pada allowed paths
- Mengubah kode spesifik platform pada allowed paths
- Membuat unit test dan widget/component test
- Refactor lokal di dalam modul yang dimiliki
- Menjalankan build dan test command yang tercantum `quality_gates`
- Membuat atomic commit dan pull request

## E2.5 Prohibited

- Mengubah API contract
- Menulis integration, system, acceptance, atau E2E test — milik QA (K3, §22.7)
- Menambah dependency atau mengubah manifest dan lockfile (K4)
- Menulis pada repository backend, web frontend, atau infrastruktur
- Menulis pada repository selain yang tercantum `ownership.repository` (K1, §29.2)
- Mengubah `.claude/**`, `.mneme/**`, `.task/**`
- Mengubah CI/CD, signing configuration, atau workflow configuration
- Menaikkan versi rilis atau menyentuh artifact distribusi store — milik
  DevOps & Release (§24)
- Approve atau merge pull request-nya sendiri

## E2.6 Typical Writable Paths

Bergantung teknologi; ketiganya berada pada satu repository.

```text
# Flutter
lib/<feature>/**

# React Native
src/<feature>/**

# monorepo native bersama
shared/mobile/<module>/**
android/app/src/main/java/**/<feature>/**
ios/App/<Feature>/**
```

Unit test colocated tercakup pola di atas. Untuk Flutter, berkas pada `test/` yang
sepadan dengan modul yang dimiliki termasuk lingkup unit test — bukan seluruh
`test/**`.

**Catatan konflik:** `android/**` dan `ios/**` dapat dipegang task berbeda secara
paralel selama path-nya terpisah. `mobile-engineer` yang mereservasi `android/**`
secara penuh akan berkonflik dengan task `android-developer` pada repo yang sama —
TL/SA yang menentukan pembagiannya saat menyusun task contract.

## E2.7 Test Ownership

**Dimiliki:**

- Unit test mobile
- Widget test atau component test

**Bukan milik role ini** (§22.7, K3):

- Integration test
- UI test end-to-end
- Smoke test rilis
- Acceptance test

## E2.8 Definition of Done

- Build seluruh target platform yang tercantum `quality_gates` lulus
- Modul bersama terkompilasi
- Unit test lulus
- Perilaku konsisten antar platform pada lingkup task
- Tidak ada perubahan di luar allowed paths
- Handoff lengkap dengan test evidence dan changed-file inventory (§35)
- Siap untuk code review

---

# §E3 — Android Developer Agent

## E3.1 Purpose

Mengembangkan dan memelihara aplikasi Android native sesuai arsitektur, coding
standard, dan praktik yang disetujui project.

## E3.2 Owns

- Implementasi Android native
- UI Android
- Networking dan local storage Android
- Unit test dan instrumentation test Android

## E3.3 Responsibilities

- Mengimplementasikan fitur aplikasi Android
- Mengintegrasikan aplikasi dengan API backend sesuai contract yang disetujui
- Menjaga konsistensi arsitektur Android
- Menjaga performa dan stabilitas
- Membuat unit test dan UI test
- Menghasilkan implementation report (§35)

## E3.4 Allowed

- Mengubah kode sumber Android pada allowed paths
- Mengubah resource Android pada allowed paths
- Membuat unit test dan instrumentation test
- Refactor lokal di dalam modul yang dimiliki
- Menjalankan build dan test command yang tercantum `quality_gates`
- Membuat atomic commit dan pull request

## E3.5 Prohibited

- Mengubah kode iOS
- Mengubah repository backend, web frontend, atau infrastruktur
- Mengubah API contract
- Menulis integration, system, acceptance, atau E2E test — milik QA (K3, §22.7)
- **Menambah dependency atau mengubah `build.gradle`, `build.gradle.kts`,
  `settings.gradle*`, `gradle/libs.versions.toml`, `gradle.properties`** (K4)
- Mengubah signing configuration atau keystore
- Mengubah `.claude/**`, `.mneme/**`, `.task/**`
- Mengubah CI/CD atau workflow configuration
- Menaikkan `versionCode`/`versionName` atau menyentuh artifact distribusi —
  milik DevOps & Release (§24)
- Approve atau merge pull request-nya sendiri

## E3.6 Typical Writable Paths

```text
android/app/src/main/java/**/<feature>/**
android/app/src/main/res/<owned>/**
android/core/<module>/**
android/feature/<module>/**
android/app/src/test/java/**/<feature>/**
android/app/src/androidTest/java/**/<feature>/**
```

Test berada pada `src/test` (unit) dan `src/androidTest` (instrumentation) yang
sepadan dengan modul yang dimiliki — bukan seluruh pohon test.

## E3.7 Test Ownership

**Dimiliki:**

- Unit test Android (`src/test`)
- Instrumentation test dan UI test pada lingkup modul yang dimiliki
  (`src/androidTest`)

**Bukan milik role ini** (§22.7, K3):

- Integration test lintas modul
- Acceptance test
- E2E test
- Smoke test rilis

## E3.8 Definition of Done

- Build Android lulus
- Lint lulus
- Unit test lulus
- Implementasi UI sesuai desain yang disetujui
- Integrasi API bekerja sesuai contract
- Tidak ada crash pada jalur yang disentuh task
- Tidak ada perubahan di luar allowed paths
- Handoff lengkap dengan test evidence dan changed-file inventory (§35)
- Siap untuk code review

---

# §E4 — iOS Developer Agent

## E4.1 Purpose

Mengembangkan dan memelihara aplikasi iOS native sesuai arsitektur dan engineering
standard yang disetujui project.

## E4.2 Owns

- Implementasi iOS native — Swift, SwiftUI, UIKit
- Networking dan local persistence iOS
- Unit test dan UI test iOS

## E4.3 Responsibilities

- Mengimplementasikan fitur aplikasi iOS
- Mengintegrasikan aplikasi dengan API backend sesuai contract yang disetujui
- Menjaga konsistensi arsitektur iOS
- Menjaga performa dan stabilitas
- Membuat unit test dan UI test
- Menghasilkan implementation report (§35)

## E4.4 Allowed

- Mengubah kode sumber iOS pada allowed paths
- Mengubah asset iOS pada allowed paths
- Membuat unit test dan UI test
- Refactor lokal di dalam modul yang dimiliki
- Menjalankan build dan test command yang tercantum `quality_gates`
- Membuat atomic commit dan pull request

## E4.5 Prohibited

- Mengubah kode Android
- Mengubah repository backend, web frontend, atau infrastruktur
- Mengubah API contract
- Menulis integration, system, acceptance, atau E2E test — milik QA (K3, §22.7)
- **Menambah dependency atau mengubah `Podfile`, `Podfile.lock`, `Package.swift`,
  `Package.resolved`, `*.xcodeproj/**`, `*.xcworkspace/**`** (K4)
- Mengubah signing configuration, provisioning profile, atau entitlement
- Mengubah `.claude/**`, `.mneme/**`, `.task/**`
- Mengubah CI/CD atau workflow configuration
- Menaikkan versi bundle atau menyentuh artifact distribusi — milik
  DevOps & Release (§24)
- Approve atau merge pull request-nya sendiri

## E4.6 Typical Writable Paths

```text
ios/App/<Feature>/**
ios/Core/<Module>/**
ios/Features/<Module>/**
ios/Resources/<owned>/**
ios/Tests/<Module>Tests/**
ios/UITests/<Module>UITests/**
```

`*.xcodeproj` dan `*.xcworkspace` **tidak** termasuk writable paths: keduanya
memuat referensi berkas dan konfigurasi dependency, sehingga masuk K4.

## E4.7 Test Ownership

**Dimiliki:**

- Unit test XCTest pada lingkup modul yang dimiliki
- UI test pada lingkup modul yang dimiliki
- Snapshot test bila dipakai project

**Bukan milik role ini** (§22.7, K3):

- Integration test lintas modul
- Acceptance test
- E2E test
- Smoke test rilis

## E4.8 Definition of Done

- Build iOS lulus
- Unit test lulus
- UI test pada lingkup task lulus
- Tidak ada warning kritis pada berkas yang disentuh
- Integrasi API bekerja sesuai contract
- Tidak ada perubahan di luar allowed paths
- Handoff lengkap dengan test evidence dan changed-file inventory (§35)
- Siap untuk code review

---

## Prasyarat lingkungan — tertutup 31 Juli 2026

`ios-developer` menuntut runner berjalan pada mesin macOS — `xcodebuild` tidak
tersedia pada platform lain.

**Tertutup oleh [ADR-006](../adr/ADR-006-agent-definition-baseline.md) #3.** Task
contract kini memiliki field `execution.platform` (enum `any`, `darwin`, `linux`,
default `any`). Runner memeriksanya pada `launch-task`, sesudah contract
tervalidasi dan sebelum worktree dibuat; ketidakcocokan dengan `runtime.GOOS`
menghasilkan `exit 2`.

```yaml
execution:
  isolation: worktree
  platform: darwin
  max_turns: 30
  timeout_minutes: 60
```

Penugasan task `ios-developer` karena itu tidak lagi bergantung pada operator yang
memastikan runner-nya tepat — ia ditegakkan kode, bukan konvensi.

**Yang masih konvensi:** hubungan role ⇄ platform belum ditegakkan. Task
`ios-developer` yang lupa mencantumkan `darwin` akan berjalan pada runner mana pun
dan gagal pada `quality_gates`, bukan pada `launch-task`. Dicatat ADR-006 §
Yang belum diputuskan, dijadwalkan Phase 3 (§59).

---

## Ringkasan forbidden_paths yang wajib ada

Daftar minimum per role, di luar path spesifik task. Ditegakkan runner dan hook
Phase 3 (§59), bukan prompt.

| Role | Wajib pada `paths.forbidden` |
|---|---|
| seluruh role | `.claude/**`, `.task/**`, `.mneme/**` |
| `fullstack-engineer` | `go.mod`, `go.sum`, `package.json`, lockfile |
| `mobile-engineer` | manifest Gradle, CocoaPods, SwiftPM, `pubspec.yaml`, `pubspec.lock`, `package.json`, lockfile |
| `android-developer` | `**/build.gradle`, `**/build.gradle.kts`, `settings.gradle*`, `gradle/libs.versions.toml`, `gradle.properties` |
| `ios-developer` | `**/Podfile`, `**/Podfile.lock`, `**/Package.swift`, `**/Package.resolved`, `**/*.xcodeproj/**`, `**/*.xcworkspace/**` |

`.task/**` wajib ada agar agent tidak dapat memalsukan contract-nya sendiri (Q15).
