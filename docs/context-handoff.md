# goft8 — Context Handoff

**Last updated:** 2026-04-13 (evening session — step 5 landed, working tree clean)
**Branch:** main
**Authoritative design reference:** [`docs/design.md`](design.md)

This document captures the state of the v0.1 migration so that any
subsequent session (human or agent) can pick up without re-deriving
decisions from chat history.

---

## 0. Standing ground rules (read first)

1. **The research Go port is frozen.** The files copied from
   `~/Development/go-ft8/research/` (`sync8.go`, `sync_d.go`,
   `metrics.go`, `ldpc.go`, `ldpc_parity.go`, `crc.go`, `decode.go`,
   `downsample.go`, `subtract.go`, `encode.go`, `ap.go`, `pack28.go`,
   `message.go`, `fft.go`, `constants.go`, `params.go`) are bit-exact
   with WSJT-X 2.7.0 main-loop output and must be treated as
   read-only reference. Do **not** refactor, dedup, rename, or "fix"
   them even when the logic looks odd — the oddness is load-bearing.
   Only `docs/design.md` Migration step 1 file list is legitimately
   ported content; everything else in the repo is new wrapper code.
2. **Only `~/Development/go-ft8/research/` is the source of truth.**
   Ignore `~/Development/go-ft8/*.go` at the root — those are the
   frozen `ft8x` production package and are on the explicit
   "do not migrate" list.
3. **New code goes in the wrapper layer only:** `decoder.go`,
   `decoded.go`, `decoder_options.go`, `doc.go`, `encoder.go`,
   `message_parse.go`, test files, `cmd/goft8/`.
4. **Additive ports from research/ are OK** — e.g. bringing over a
   helper currently confined to research's tests — but they must be
   verbatim copies, not rewrites.

Why these rules exist: the 2026-04-13 evening session burned time
chasing regressions caused by speculative "improvements" to the
research internals (§3.3 xbase SNR formula, §3.4 `indexx` dedup).
Both were reverted. See §3 for the writeup.

---

## 1. Where the project stands

### 1.1 Committed to main

| Commit    | Scope                                                                  |
|-----------|------------------------------------------------------------------------|
| `9d3c41c` | Initial commit                                                         |
| `0d71fd5` | Initialize project structure                                           |
| `661fd87` | Add NOTICE, .gitignore, and v0.1 design document                       |
| `6e79513` | Initial import: WSJT-X 2.7.0 parity port from `go-ft8/research/`       |
| `3123dab` | Public API wrapper layer + `DecodeResult.Pass` field + LDPC renames   |
| `a32e003` | Rename `sync` local to `syncPower` for clarity (power calculation)    |
| `13a9b9f` | Fix `sync` shadowing in `sync_d.go` (local accumulator → `syncPower`) |
| `d7b9ccf` | Add FT8 capture fixtures for regression testing (migration step 3)    |
| `273e442` | Add 3-capture decode regression test at research parity (step 5)      |

### 1.2 Working tree

Clean. `go build ./...`, `go vet ./...`, and
`go test -run TestDecodeCaptures ./...` all pass. Migration steps 1,
2, 3, and 5 are complete and committed.

---

## 2. What the design says to do next

From [`docs/design.md`](design.md), "Migration plan from
go-ft8/research/", the remaining steps are:

3. ~~**Copy test fixtures**~~ — done in `d7b9ccf`. Three 12 kHz mono
   PCM 15-second captures recorded by the author via station-manager's
   audio library live in `testdata/ft8_cap{1,2,3}.wav`. Copied, not
   sourced from WSJT-X samples, so the MIT licensing story is clean.
4. **Copy diagnostic programs** to `internal/fortran_test/`:
   `dump_all_passes.f90`, `crc_shim.f90`, `stdcall_shim.f90`,
   plus a `README.md` with compile recipe and license notice. These
   are developer-only GPLv3-when-linked binaries — the `.f90` sources
   are our own and MIT-licensed, but compiled outputs must never be
   committed (see NOTICE and `internal/fortran_test/.gitignore`).
5. ~~**Regression test**~~ — done in `273e442`. `decode_captures_test.go`
   pins `DecodeIterative` at **10 / 14 / 23** decodes across
   `ft8_cap{1,2,3}.wav` with AP CQ-only, depth 3, 200..2600 Hz.
   **Note the corrected yield**: earlier revisions of this doc and
   `design.md` cited "11/14/23+2". That figure came from
   `research/root_cause_all_test.go`, which has its own inline iteration
   with a `basebandTimeScan` retry fallback; research's production
   `DecodeIterative` is at 10/14/23. See §3.5 for the follow-up plan
   to lift cap1 from 10 → 11.
6. **Unit tests** — LDPC generator matrix vs Fortran reference, CRC-14
   test vectors, message pack/unpack round-trip, sync8 candidate list
   stability.
7. **CLI** — `cmd/goft8/main.go`, flags `-wav foo.wav -fmin 200 -fmax 3000`.
8. **README.md** — usage example, install, decode-yield numbers, links
   to design.md.
9. **Gate** — `go build ./... && go vet ./... && go test -short ./... &&
   go test ./... && go test -bench .` all pass.
10. **Commit** each remaining step as its own follow-up commit.

After that, the v0.1 milestone is complete and the author archives
`go-ft8` with a final commit pointing at `goft8`.

Natural next step on resume: **step 6 (unit tests)**. Step 4 (Fortran
diagnostics) is optional tooling — it's useful for future regression
investigations but doesn't block the v0.1 gate. Step 6 gives unit-level
coverage that will legitimise the step-3.1 unexport refactor later.

---

## 3. Known follow-ups not yet scheduled

These are design-compliant but deferred from the current session; pick
them up when convenient.

### 3.1 Unexport internals that leak out of the public surface

The original `research/` package exported many symbols that the design
expects to be unexported once wrapped by the public API. The wrapper
layer was added without flipping visibility, so the following symbols
are still exported and visible in `go doc`:

- `DecodeIterative`, `DecodeSingle`, `DecodeParams`,
  `DefaultDecodeParams`, `DecodeCandidate`, `CandidateFreq`,
  `DecodeResult`, `DecodeLDPC`, `DecodeLDPCF32`
- `Sync8`, `Sync8d`, `Sync8FindCandidates`, `HardSync`,
  `ComputeSpectrogramForTest`, `ComputeSync2DForTest`
- `ComputeSymbolSpectra`, `ComputeSoftMetrics`, `ComputeAPSymbols`,
  `ApplyAP`, `BitsToC77`, `Unpack77`, `GenFT8Tones`, `GenFT8CWave`,
  `SubtractFT8`, `TwkFreq1`, `FFT`, `IFFT`, `RealFFT`,
  `SpectrogramFFT3840`
- Types: `Candidate`, `Downsampler`, `Spectrogram`
- Vars/consts exported as upper-case: `GrayMap`, `GrayUnmap`, `Icos7`,
  `LDPCMn`, `LDPCNm`, `LDPCNrw`, `NP2`, `JZ`, `Fs`, …

Unexporting is a larger refactor because some are used by
existing diagnostic paths (e.g. `ComputeSync2DForTest`). The
pragmatic sequence is: add the unit tests in step 6 first (the tests
that legitimately need internals stay in `*_test.go` and retain
access), then flip visibility for the rest.

### 3.2 Taskfile assumes `cmd/goft8/` exists

`Taskfile.yml` has `cli` and `cli:run` targets that point at
`./cmd/goft8`. The directory does not exist yet — those targets will
fail until step 7 lands. They are included now so the contract is in
place.

### 3.3 `xbase`-calibrated SNR stub (**attempted, reverted — do not retry**)

`decode.go` still has the `TODO: full xbase SNR requires s8 from
DecodeSingle` at the `DecodeIterative` SNR recomputation site, and the
SNR returned is the tone-ratio estimate from `DecodeSingle`.

**Status:** attempted in the 2026-04-13 evening session. The Fortran
formula `xsnr2 = 10*log10(xsig/xbase/3e6 - 1) - 27` was implemented by
threading an unexported `xsig` field through `DecodeCandidate` and
computing the calibrated SNR in `DecodeIterative`. It produced values
so pessimistic that every decode except the 2–3 strongest on each
capture clamped to the −24 floor (research uses pure tone-ratio and
reports −5 to −20 dB for the same decodes). The formula as ported is
almost certainly wrong in either scaling or units, and the Fortran
reference is not in the repo to verify against. The change was
reverted; the tone-ratio SNR returned by research's `DecodeSingle` is
what the library ships.

**Not a parity blocker.** Research's production `DecodeIterative`
doesn't do xbase calibration either and still hits the Fortran main-loop
decode count on all three captures. Displayed SNR is a cosmetic
difference, not a yield one. Leave this TODO in place until someone
wants to match WSJT-X's displayed dB value exactly.

### 3.4 `indexx` duplicates `argsortAsc` (**attempted, reverted — leave alone**)

`sync8.go:indexx` (sort.Slice over a sub-range) and `ldpc.go:argsortAsc`
(Numerical Recipes quicksort over a whole slice) both produce ascending
argsorts. The original suggestion was to collapse them.

**Status:** attempted in the 2026-04-13 evening session. `indexx` was
renamed to `argsortAscRange` and rewritten as a wrapper around
`argsortAsc`, reusing its NR tie-breaking. This was intended to make
Sync8 normalization use the same tie-breaking as OSD (argued to better
match Fortran's own `indexx`). Running the regression test showed cap1
dropping from 10 → still 10 (no change), i.e. the rename is safe on
these captures, but the change had no benefit and adding it violates
ground rule §0.1 ("do not touch research internals"). Reverted.

**Do not revisit.** The duplication is superficial — the two functions
have different call patterns (range vs whole slice) and different
measured tie-breaking requirements. If you ever want to unify them,
do it by copying the research authors' approach, not ours.

### 3.5 `basebandTimeScan` retry fallback — additive port for +1 decode on cap1

The regression test in `273e442` pins cap1 at 10 decodes, which is what
research's production `DecodeIterative` produces. The 11th reference
decode `<...> RA6ABC KN96 @1814 Hz` is recovered in
`research/root_cause_all_test.go` only, by an inline retry loop that:

1. Reuses one `Downsampler` across all candidates in a pass with
   `newdat := (i == 0)` (so the cached FFT state is reused).
2. On `DecodeSingle` failure with `cand.SyncPower >= 2.0`, runs
   `basebandTimeScan(dd, ds, f0)` — a full-range sync scan over every
   `idt ∈ [0, NP2]` stride 4 — to find a better DT.
3. If the alt DT differs from sync8's suggestion by > 0.1 s, retries
   `DecodeSingle` at the alt DT with `newdat=false`.

`basebandTimeScan` is defined at
`research/iterative_decode_test.go:259`. It uses `Sync8d`, `Downsampler`,
`NP2`, `Dt2`, and `complex` tweaking factors that are all present in
goft8 already. The retry guard is ~7 lines in
`research/root_cause_all_test.go` ~218–231.

**Upgrade path (follow-up (b)):**
1. Copy `basebandTimeScan` verbatim into goft8 (probably
   `decode.go` near the bottom) as an unexported helper.
2. In `DecodeIterative`, switch the candidate loop to share one
   `Downsampler` per pass with `newdat := (i == 0)`, and add the retry
   guard on failure.
3. Bump `decode_captures_test.go` cap1 `want: 10` → `want: 11` and add
   `"<...> RA6ABC KN96"` to its `mustSee`.

This is an **additive port** (extends functionality by bringing more of
research/ verbatim into goft8), not a modification, so it's compatible
with ground rule §0. Do it as its own commit so bisection against
`273e442` stays clean.

---

## 4. Locked decisions (unchanged since 2026-04-13)

All 10 design questions in `docs/design.md` §"Locked decisions" are
resolved. Do not re-open them without the author's sign-off:

1. Stateful `Decoder`, NOT concurrency-safe.
2. `Decode()` returns `([]Decoded, error)` directly.
3. Audio input is `[]float32` at 12 kHz mono, exactly
   `AudioSamplesPerCycle` (180000) samples.
4. `Decoded.SNR` is `int` dB, clamped to [-24, +40].
5. No `context.Context` on `Decode()` in v0.1.
6. Minimal `Logger` interface (`Printf(format, args...)` only).
7. Public `Message` + `ParseMessage` shipped in v0.1.
8. `Encoder` type stub in v0.1; `Encode()` panics with
   `"goft8.Encoder.Encode: not yet implemented, scheduled for v0.2"`.
9. v0.1 scope is fixed per the milestone table.
10. Design doc is the single source of truth at
    `github.com/ColonelBlimp/goft8/docs/design.md`.

---

## 5. Quick-start for the next session

```bash
cd ~/Development/goft8

# Verify current state
git status                  # should be clean
git log --oneline -5        # latest: 273e442, d7b9ccf, 13a9b9f, a32e003, 3123dab
go build ./... && go vet ./...
go test -run TestDecodeCaptures -v ./...   # ~80 s, should be PASS 10/14/23
go doc . | less             # inspect the public surface
```

**Resume migration at step 6 (unit tests).** Ports to write, all from
`~/Development/go-ft8/research/`:

- `ldpc_test.go` — LDPC generator matrix vs Fortran reference, CRC-14
  test vectors (see research's `crc.go` and `ldpc.go` for the helpers
  to call and reference values to compare against).
- `message_test.go` — pack/unpack round-trip over a curated sample of
  message types (standard, CQ, report, RRR, 73, free text, telemetry).
- `sync8_test.go` — candidate list stability: decode a known WAV and
  assert the candidate list length, ordering, and top-N `(freq, DT)`
  values are stable across runs.

When in doubt, re-read `docs/design.md` — it is the contract. This
handoff is a convenience, not a specification. And re-read §0 above
before touching any `.go` file that isn't in the wrapper layer.
