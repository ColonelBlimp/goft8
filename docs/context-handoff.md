# goft8 — Context Handoff

**Last updated:** 2026-04-13
**Branch:** main
**Authoritative design reference:** [`docs/design.md`](design.md)

This document captures the state of the v0.1 migration so that any
subsequent session (human or agent) can pick up without re-deriving
decisions from chat history.

---

## 1. Where the project stands

### 1.1 Committed to main

| Commit    | Scope                                                                  |
|-----------|------------------------------------------------------------------------|
| `9d3c41c` | Initial commit                                                         |
| `0d71fd5` | Initialize project structure                                           |
| `661fd87` | Add NOTICE, .gitignore, and v0.1 design document                       |
| `6e79513` | Initial import: WSJT-X 2.7.0 parity port from `go-ft8/research/`       |

### 1.2 Uncommitted working tree (in progress)

The following are in the working tree but **not yet committed**:

- **Public API layer (design.md step 2):**
  - `doc.go` — package-level godoc overview
  - `decoded.go` — public `Decoded` struct, `AudioSamplesPerCycle`,
    `AudioSampleRate`, `DepthFast/Normal/Deep`, `APType*` constants
  - `decoder.go` — rewritten: `Decoder`, `NewDecoder`, `Decode`,
    `Reset`, `DecodeWAV` + inline RIFF/WAVE reader (PCM16 + float32)
  - `decoder_options.go` — `DecoderOption`, `Logger`, all `With*`
    options (`WithMyCall`, `WithDxCall`, `WithFreqRange`, `WithDepth`,
    `WithMaxPasses`, `WithAPEnabled`, `WithCQOnlyAP`,
    `WithAudioStartSeconds`, `WithLogger`)
  - `encoder.go` — public `Encoder` stub (panics per design)
  - `message_parse.go` — public `Message`, `MsgType`, `ParseMessage`
- **Internal addition:** `DecodeCandidate.Pass int` field, set by
  `DecodeIterative` to the 1-based iteration index so the public
  `Decoded.Pass` is reachable.
- **Go-idiomatic renames (all within-package, call sites updated):**
  - `Decode174_91`         → `DecodeLDPC`
  - `Decode174_91_F32`     → `DecodeLDPCF32`
  - `encode174_91NoCRC`    → `encodeLDPCNoCRC`
  - `unpacktext77`         → `unpackText77`
  - `ihashcall`            → `hashCall`
  - `stdcall`              → `isStdCall`
  - `initCsync`            → `initCSync`
- **Tooling:** `Taskfile.yml` at repo root (build/vet/test/bench/cli/…).

`go build ./...` and `go vet ./...` both exit 0 with this working set.

---

## 2. What the design says to do next

From [`docs/design.md`](design.md), "Migration plan from
go-ft8/research/", the steps beyond step 2 are:

3. **Copy test fixtures** — `testdata/ft8_cap{1,2,3}.wav` from
   `~/Development/go-ft8/research/` (copy, don't move; `go-ft8` still
   uses them until archived).
4. **Copy diagnostic programs** to `internal/fortran_test/`:
   `dump_all_passes.f90`, `crc_shim.f90`, `stdcall_shim.f90`,
   plus a `README.md` with compile recipe and license notice.
5. **Regression test** — `decode_captures_test.go` asserting
   11/14/23+2 decodes on the 3 captures (WSJT-X 2.7.0 parity).
6. **Unit tests** — LDPC generator matrix vs Fortran reference, CRC-14
   test vectors, message pack/unpack round-trip, sync8 candidate list
   stability.
7. **CLI** — `cmd/goft8/main.go`, flags `-wav foo.wav -fmin 200 -fmax 3000`.
8. **README.md** — usage example, install, decode-yield numbers, links
   to design.md.
9. **Gate** — `go build ./... && go vet ./... && go test -short ./... &&
   go test ./... && go test -bench .` all pass.
10. **Commit** as one or more follow-up commits after the initial
    import commit (which already exists).

After that, the v0.1 milestone is complete and the author archives
`go-ft8` with a final commit pointing at `goft8`.

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

### 3.3 `xbase`-calibrated SNR stub

`decode.go` has a `TODO: full xbase SNR requires s8 from DecodeSingle`
at the `DecodeIterative` SNR recomputation site. The SNR currently
returned is the tone-ratio estimate from `DecodeSingle`, not the
`xbase`-calibrated value WSJT-X displays. This is a known gap vs
bit-exact parity and should be resolved before the regression test in
step 5 can assert exact yields.

### 3.4 `indexx` duplicates `argsortAsc`

Both helpers live in `sync8.go` / `ldpc.go` and both do ascending
argsort. Worth collapsing or at least confirming the distinction —
`indexx` is the Numerical Recipes name from the Fortran port and isn't
Go-idiomatic.

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
git status
task                        # build + vet + short tests (once tests exist)
task build                  # just build
go doc . | less             # inspect the public surface

# Resume migration plan — step 3 (test fixtures) is the next unit of work
mkdir -p testdata
cp ~/Development/go-ft8/research/testdata/ft8_cap1.wav testdata/
cp ~/Development/go-ft8/research/testdata/ft8_cap2.wav testdata/
cp ~/Development/go-ft8/research/testdata/ft8_cap3.wav testdata/
```

When in doubt, re-read `docs/design.md` — it is the contract. This
handoff is a convenience, not a specification.
