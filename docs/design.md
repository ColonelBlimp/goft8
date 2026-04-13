# goft8 — Design Document (v0.1 draft)

**Status:** Draft, awaiting sign-off on 10 open questions before v0.1
file migration begins.
**Last updated:** 2026-04-13

## Purpose

`goft8` is a clean-room Go implementation of the FT8 digital-mode
decoder and encoder, embedded in amateur-radio station software as a
pure-Go (later optional-CGO) library dependency. Its primary consumer
is the author's station-manager project, which cannot link against
WSJT-X or JTDX due to hamlib vs custom-serial-port-lib conflict.

The library is **not** a fork or binding of WSJT-X — it's a clean-room
re-implementation of published FT8 algorithms, derived from study of
the GPLv3 WSJT-X 2.7.0 Fortran source as reference. The Go code is
original and licensed under MIT.

See `NOTICE` for the clean-room attribution statement.

## Provenance — why goft8 exists

1. Station-manager is a personal ham-radio logging and control app
   using a custom serial-port library.
2. It needs to run FT8 itself, not spawn WSJT-X / JTDX, because those
   use hamlib and the two can't coexist in the same process.
3. The FT8 port started embedded in station-manager, grew too large,
   and was broken out as a standalone library.
4. goft8 replaces the author's earlier `go-ft8` prototype (which will
   be archived). The lessons learned from `go-ft8` — a working
   WSJT-X 2.7.0 port with bit-exact Fortran parity — directly inform
   this design.

## Strategic framing

The project is a **practical receiver library** embedded in a personal
station manager. It is not primarily:

- a reference implementation (we don't care about 1:1 upstream code
  matching, only decode yield parity)
- a research platform (we don't need to expose decoder variants for
  experimentation, though the API leaves room)

Success = station-manager can decode FT8 at weak-signal performance
approaching the best open-source decoder, with a stable, idiomatic
Go API.

## The hybrid approach (approved roadmap)

Three public decoders were benchmarked against the author's 3 WSJT-X
test captures:

| Capture       | WSJT-X 3.0.0 display | JTDX 2.2.159 | Our WSJT-X 2.7.0 port |
|---------------|----------------------|--------------|------------------------|
| ft8_cap1.wav  | 13                   | 11           | 11                     |
| ft8_cap2.wav  | 15                   | 17           | 14                     |
| ft8_cap3.wav  | 21                   | 20           | 23                     |

**WSJT-X 3.0.0 has no public source** (only 2.7.0 is published), so
targeting 3.0.0's yield via clean-room porting is impossible. Three
signals unique to 3.0.0 are unreachable.

**JTDX is a substantially different decoder architecture** from
WSJT-X 2.7.0 — not an incremental enhancement. Its `ft8b.f90` is 2034
lines vs WSJT-X's 503. JTDX adds ~2700 lines of Fortran across nine
new FT8-specific files, a 27-entry AP dispatch table, pre-computed
message-template banks, matched-filter decoders, cross-cycle state
machine, AGC preprocessing, time-reversed decode paths, and
adaptive sync-quality gating. It reaches 7 signals on our captures
that nothing else does, at the cost of missing ~8 signals WSJT-X 2.7.0
decodes. Neither decoder is strictly better.

**Decision:** ship v0.1 at WSJT-X 2.7.0 parity with a clean API
designed to accommodate JTDX features in later minor versions.
Port JTDX enhancements incrementally:

| Version | Feature                                              | Est. LOC |
|---------|------------------------------------------------------|----------|
| v0.1    | WSJT-X 2.7.0 parity, public API, Message parser,     | (port    |
|         | DecodeWAV, Encoder type stub, CLI, tests, benchmarks | existing)|
| v0.2    | AGC preprocessing (agccft8.f90)                      | ~200     |
| v0.3    | 27-type AP system + ft8apset rewrite                 | ~500     |
| v0.4    | Matched-filter decoder (ft8s.f90, ft8mf1/mfcq)       | ~1000    |
| v0.5    | Multi-pass with time reversal + audio decimation     | ~500     |
| v0.6    | Cross-cycle state machine + virtual decode attempts  | ~1500    |

Each milestone ships independently. The v0.1 API is designed to
accommodate every milestone without breakage.

## Package layout

Single primary package `goft8` at module root:

```
github.com/ColonelBlimp/goft8/
├── go.mod                   module github.com/ColonelBlimp/goft8
├── LICENSE                  MIT, © 2026 ColonelBlimp (7Q5MLV)
├── NOTICE                   clean-room attribution text
├── README.md
├── doc.go                   package-level godoc overview
│
├── decoder.go               Decoder, NewDecoder, Decode, Reset
├── decoder_options.go       DecoderOption functional options
├── decoded.go               Decoded result type
├── encoder.go               Encoder (v0.1 stub, v0.2 impl)
├── message.go               Message, ParseMessage (public helpers)
│
├── sync8.go                 Costas sync candidate search (private)
├── sync_d.go                Fine-sync (private)
├── metrics.go               Symbol spectra + bmet extraction (private)
├── ldpc.go                  BP+OSD LDPC decoder (private)
├── ldpc_parity.go           parity/generator data (private)
├── crc.go                   CRC-14 (private)
├── ap.go                    A-priori injection (private)
├── pack28.go                callsign hashing (private)
├── downsample.go            12k → 200 Hz downsampler (private)
├── subtract.go              signal subtraction (private)
├── encode_tones.go          tone generation for Encoder (private)
├── waveform.go              GFSK pulse generation (private)
├── fft.go                   pure-Go mixed-radix FFT (private)
├── constants.go             shared constants (private)
│
├── decoder_test.go          public API regression tests
├── decode_captures_test.go  3-capture end-to-end regression
├── ldpc_test.go             LDPC unit tests
├── message_test.go          message parsing tests
├── encoder_test.go          encode/decode round-trip (v0.2)
├── benchmarks_test.go       performance benchmarks
│
├── testdata/
│   ├── ft8_cap1.wav
│   ├── ft8_cap2.wav
│   ├── ft8_cap3.wav
│   └── README.md
│
├── cmd/
│   └── goft8/               CLI tool
│       └── main.go
│
├── internal/
│   └── fortran_test/        local-only developer diagnostic tools
│       ├── dump_all_passes.f90
│       ├── crc_shim.f90
│       ├── stdcall_shim.f90
│       └── README.md        compile recipe, license notice
│
└── docs/
    └── design.md            (this file, living reference)
```

Rationale:
- Flat public surface; single import `github.com/ColonelBlimp/goft8`.
- Unexported symbols via Go case-based visibility — no `internal/`
  ceremony for private code within the same package.
- `cmd/goft8/` — standard Go pattern for tools bundled with a library.
- `internal/fortran_test/` — build-time-excluded diagnostic programs
  that link GPLv3 Fortran at compile time. Sources are our own
  (parameter setup + subroutine calls + I/O, no copied WSJT-X code);
  compiled binaries are never distributed.

## Public API (v0.1 draft)

### Core types

```go
// Package goft8 is a clean-room Go implementation of the FT8 digital
// mode decoder and encoder.
//
// Typical receive usage:
//
//	dec := goft8.NewDecoder(goft8.WithMyCall("W1ABC"))
//	for audio := range cycleCh {  // 180000 samples, 12 kHz mono
//	    decodes, err := dec.Decode(audio)
//	    if err != nil { log.Println(err); continue }
//	    for _, d := range decodes {
//	        fmt.Printf("%+d %4.1f %6.1f %s\n", d.SNR, d.DT, d.Freq, d.Message)
//	    }
//	}
//
// A Decoder is stateful and NOT safe for concurrent use. Each receive
// stream should have its own Decoder.
package goft8

// Decoded is a single FT8 signal successfully decoded from one
// 15-second audio window.
type Decoded struct {
    // Message is the 37-character-max decoded text,
    // e.g. "CQ W1ABC FN31".
    Message string

    // Freq is the refined carrier frequency in Hz (sub-Hz precision).
    Freq float64

    // DT is the time offset relative to the nominal 0.5 s TX start
    // in seconds. Matches WSJT-X display convention; a signal
    // arriving on time has DT ≈ 0.
    DT float64

    // SNR is the estimated signal-to-noise ratio in dB (2500 Hz
    // bandwidth), clamped to [-24, +40].
    SNR int

    // Pass is the iterative-decode pass this signal was recovered on
    // (1, 2, or 3).
    Pass int

    // APType is the a-priori decoding type used, 0 if none.
    APType int

    // NHardErrors is the number of hard-decision bit errors corrected
    // by the LDPC decoder. 0 means clean; higher values indicate a
    // weaker decode that was still CRC-validated.
    NHardErrors int

    // Tones holds the 79 channel tone indices (0..7) representing
    // this signal, useful for subtraction or re-encoding.
    Tones [79]int
}
```

### Decoder

```go
// Decoder runs the FT8 decode pipeline on 15-second audio cycles.
// Construct with NewDecoder, reuse across multiple Decode calls so
// cross-cycle state (used by future matched-filter passes) persists.
//
// Decoder is NOT safe for concurrent use by multiple goroutines.
type Decoder struct {
    // unexported
}

// NewDecoder creates a Decoder configured by the given options.
// With no options it uses defaults equivalent to WSJT-X's "normal"
// mode with AP disabled and a full 200-3000 Hz audio search.
func NewDecoder(opts ...DecoderOption) *Decoder

// Decode runs the full decode pipeline on one 15-second audio window.
//
// audio must be exactly AudioSamplesPerCycle samples of 12 kHz mono
// float32 PCM in the approximate range [-1, +1]. The decoder
// re-normalizes internally and tolerates modest normalization
// differences.
//
// Returns all distinct decoded signals ordered by pass then by sync
// power. A nil slice means no decodes; error is non-nil only for
// unrecoverable input errors (wrong length, etc.).
func (d *Decoder) Decode(audio []float32) ([]Decoded, error)

// Reset clears cross-cycle state maintained by the Decoder,
// equivalent to starting a fresh receive session. Call this when
// changing band, after a long idle, or to discard stale callsign
// history. Does not reset Decoder configuration.
func (d *Decoder) Reset()
```

### Options (functional options pattern)

```go
// DecoderOption configures a Decoder at construction time.
type DecoderOption func(*decoderConfig)

// WithMyCall sets the operator's callsign, used for AP decoding of
// messages addressed to the user. Standard ITU form, or compound
// (e.g. "W1ABC/P"). If empty, AP types 2+ are disabled.
func WithMyCall(call string) DecoderOption

// WithDxCall sets the DX station's callsign for AP type 3+ decoding
// during an active QSO.
func WithDxCall(call string) DecoderOption

// WithFreqRange restricts the audio frequency search to [min, max] Hz.
// Defaults to 200..3000 Hz.
func WithFreqRange(min, max int) DecoderOption

// WithDepth selects LDPC decoder depth:
//   DepthFast   = 1  // BP only
//   DepthNormal = 2  // BP + OSD(0) (default)
//   DepthDeep   = 3  // BP + OSD(2)
func WithDepth(depth int) DecoderOption

// WithMaxPasses sets the number of iterative subtraction passes.
// Defaults to 3.
func WithMaxPasses(n int) DecoderOption

// WithAPEnabled turns a-priori decoding on or off. Defaults to true
// when MyCall is set, false otherwise.
func WithAPEnabled(enabled bool) DecoderOption

// WithCQOnlyAP restricts AP to the CQ prior only. Defaults to false
// when MyCall is set.
func WithCQOnlyAP(cqOnly bool) DecoderOption

// WithAudioStartSeconds lets the caller specify how many seconds of
// silence to expect before the nominal TX start. Defaults to 0.5 s.
// Advanced users only.
func WithAudioStartSeconds(s float64) DecoderOption

// WithLogger attaches an optional logger for internal diagnostics.
// Default: no logging.
func WithLogger(logger Logger) DecoderOption
```

### Constants

```go
const (
    AudioSamplesPerCycle = 180000 // 15 s × 12 kHz
    AudioSampleRate      = 12000  // Hz
)

const (
    DepthFast   = 1
    DepthNormal = 2
    DepthDeep   = 3
)

const (
    APTypeNone     = 0
    APTypeCQ       = 1 // CQ ??? ???
    APTypeMyCall   = 2 // MyCall ??? ???
    APTypeMyDx     = 3 // MyCall DxCall ???
    APTypeMyDxRRR  = 4 // MyCall DxCall RRR
    APTypeMyDx73   = 5 // MyCall DxCall 73
    APTypeMyDxRR73 = 6 // MyCall DxCall RR73
)
```

### Convenience

```go
// DecodeWAV is a one-shot convenience: loads a 12 kHz mono 15-second
// WAV file, runs the decoder once, and returns the decoded signals.
// For a receive loop, create a Decoder once and reuse it.
func DecodeWAV(path string, opts ...DecoderOption) ([]Decoded, error)
```

### Encoder (v0.1 stub, v0.2 implementation)

```go
// Encoder generates the 12 kHz audio waveform for an FT8 message,
// for TX-side use.
type Encoder struct {
    // unexported
}

// NewEncoder creates an Encoder with the given options.
func NewEncoder(opts ...EncoderOption) *Encoder

// Encode generates the GFSK-modulated waveform for a single FT8
// message. Returns NTXSamples samples of 12 kHz mono PCM.
//
// v0.1: this method panics with a "not yet implemented" message.
// v0.2: full implementation.
func (e *Encoder) Encode(msg string) ([]float32, error)

// EncoderOption configures an Encoder.
type EncoderOption func(*encoderConfig)

// WithTxFreq sets the carrier frequency in Hz. Defaults to 1500.
func WithTxFreq(hz float64) EncoderOption

const NTXSamples = 151680 // 12.64 s × 12 kHz
```

### Message parsing

```go
// Message holds the parsed fields of a decoded FT8 message.
type Message struct {
    Raw    string  // original text
    Type   MsgType
    Call1  string  // first callsign or "CQ" marker
    Call2  string  // second callsign
    Grid   string  // 4-char grid, if present
    Report string  // signal report, if present
}

type MsgType int

const (
    MsgUnknown MsgType = iota
    MsgCQ               // "CQ [call] [grid]"
    MsgStandard         // "CALL1 CALL2 GRID"
    MsgReport           // "CALL1 CALL2 [-NN | R-NN]"
    MsgRRR              // "CALL1 CALL2 RRR"
    MsgRR73             // "CALL1 CALL2 RR73"
    Msg73               // "CALL1 CALL2 73"
    MsgFreeText         // 13-character free text
    MsgTelemetry        // 18-hex-character telemetry
)

// ParseMessage attempts to parse a decoded FT8 message into fields.
// Returns nil and a non-nil error on malformed input.
func ParseMessage(raw string) (*Message, error)
```

### Logger

```go
// Logger is the minimal interface goft8 uses for diagnostic output.
// Satisfied by stdlib log.Logger and most logging libraries.
type Logger interface {
    Printf(format string, args ...any)
}
```

## Design principles (the "why" behind the choices)

1. **Stateful Decoder, single-goroutine.** Matches JTDX's
   architecture (cross-cycle state is load-bearing for future
   matched-filter decodes). Station-manager naturally has one
   goroutine per receive chain — simpler than mutex-protecting
   internal state. Users who need parallel decoding create multiple
   Decoders.

2. **Functional options.** Extensible, typed, defaults sensibly.
   Avoids the "options struct with many fields" anti-pattern.

3. **`[]float32` at 12 kHz as input format.** Station-manager already
   handles SDR resampling; expecting pre-resampled audio is the most
   flexible contract. No io.Reader abstraction because audio cycles
   are bounded and fit in memory trivially (720 KB per cycle).

4. **`([]Decoded, error)` not `*DecodeResult`.** Simpler, avoids
   intermediate struct allocation for the zero-decode case, standard
   Go "collection + error" pattern.

5. **No `context.Context` in v0.1.** A 4–6 second decode on 15-second
   cycles isn't worth context plumbing unless cancellation is a real
   need. `DecodeWithContext(ctx, audio)` can be added later.

6. **Minimal Logger interface.** `Printf` only — satisfied by stdlib
   `log.Logger`, `*slog.Logger` with a small wrapper, or any
   third-party logger. No dependency on a specific logging library.

7. **Message parser is public in v0.1.** Station-manager needs to
   extract callsigns and grids from decoded messages for QSO logging.
   Providing this in the library avoids duplicate implementations.

8. **Encoder type stub in v0.1.** API shape is locked so station-
   manager can import and reference `goft8.Encoder` immediately,
   without being able to TX yet. Implementation in v0.2.

## Migration plan from go-ft8/research/

Source material lives in `~/Development/go-ft8/research/` (17 Go
files, ~4200 lines, bit-exact parity with WSJT-X 2.7.0). The
migration is primarily a package rename, file reorganization, and
the addition of a public API wrapper layer. The private internals
carry over with minimal changes.

Migration steps:

1. **Copy over the core ports** as unexported implementation:
   - `sync8.go`, `sync_d.go`, `metrics.go`, `ldpc.go`, `ldpc_parity.go`,
     `crc.go`, `downsample.go`, `subtract.go`, `encode.go`, `ap.go`,
     `pack28.go`, `message.go` (as internal helpers), `fft.go`,
     `constants.go`, `params.go` → merged into `constants.go`
   - Rename package declaration from `research` to `goft8`.
   - Convert existing exported symbols used only internally to
     unexported (e.g., `research.DecodeSingle` → `decodeSingle`).

2. **Write the public API layer:**
   - `decoder.go` — `Decoder`, `NewDecoder`, `Decode`, `Reset`
     wrapping the internal `decodeSingle`/`decodeIterative` calls.
   - `decoded.go` — `Decoded` struct (replaces
     `research.DecodeCandidate`).
   - `decoder_options.go` — functional options.
   - `encoder.go` — `Encoder` stub with panic.
   - `message.go` — public `Message` + `ParseMessage` using the
     existing `Unpack77` internals.
   - `doc.go` — package-level godoc overview.

3. **Copy test fixtures:**
   - `testdata/ft8_cap{1,2,3}.wav` (copy, don't move; go-ft8 still
     uses them until archived).

4. **Copy diagnostic programs:**
   - `internal/fortran_test/dump_all_passes.f90`
   - `internal/fortran_test/crc_shim.f90`
   - `internal/fortran_test/stdcall_shim.f90`
   - `internal/fortran_test/README.md` (with compile recipe)

5. **Write a single regression test:**
   - `decode_captures_test.go` — runs the 3 captures and asserts
     11/14/23+2 decodes matching `dump_all_passes` output.

6. **Write unit tests** carried from research/:
   - LDPC generator matrix vs Fortran reference
   - CRC-14 test vectors
   - Message pack/unpack round-trip
   - Sync8 candidate list stability

7. **Write the CLI:**
   - `cmd/goft8/main.go` — `goft8 -wav foo.wav -fmin 200 -fmax 3000`

8. **Write README.md** with usage example, install instructions,
   decode-yield numbers, and links to design.md.

9. **Run `go build ./...`, `go vet ./...`, `go test -short ./...`,
   `go test ./...`, `go test -bench .`** — all must pass before
   first commit.

10. **Commit as "Initial import: WSJT-X 2.7.0 parity port from
    go-ft8/research/"**.

After this commit, `go-ft8` gets a final archive commit pointing at
`goft8` and is frozen.

## Locked decisions (approved 2026-04-13)

All 10 design questions resolved. Migration begins after user commits
their initial scaffolding state.

1. **Stateful Decoder, NOT concurrency-safe.** Each receive stream
   owns its own `Decoder`. No internal mutex; state lives on the
   struct and is accessed single-threaded.
2. **`Decode()` returns `([]Decoded, error)` directly.** No
   intermediate `*DecodeResult`. Standard Go collection+error shape.
3. **Audio input is `[]float32` at 12 kHz mono.** Caller handles
   resampling. Length must be exactly `AudioSamplesPerCycle` (180000).
4. **`Decoded.SNR` is `int` dB.** Matches WSJT-X display convention;
   no pretense of sub-dB precision.
5. **No `context.Context` on `Decode()` in v0.1.** A future
   `DecodeWithContext(ctx, audio)` can be added without breaking the
   existing API if cancellation becomes a real need.
6. **Minimal `Logger` interface** (`Printf(format, args...)` only).
   Satisfied by `*log.Logger`, `*slog.Logger` with a small wrapper,
   or any third-party logger.
7. **Public `Message` + `ParseMessage` in v0.1.** Station-manager
   needs these for QSO logging; providing them in the library avoids
   duplicate implementations.
8. **`Encoder` type stub in v0.1.** The `Encoder` struct, `NewEncoder`,
   `Encode`, `EncoderOption`, `WithTxFreq`, and `NTXSamples` are all
   declared. `Encoder.Encode()` panics with
   `"goft8.Encoder.Encode: not yet implemented, scheduled for v0.2"`.
   This locks the API shape so station-manager can reference the
   types immediately.
9. **v0.1 scope** is exactly the milestone-table row above:
   RX pipeline at WSJT-X 2.7.0 parity, full public API, Message
   parser, DecodeWAV convenience, Encoder stub, `cmd/goft8` CLI,
   3-capture regression test, LDPC/CRC/message unit tests,
   benchmarks, README + godoc.
10. **Design doc lives here**, at
    `github.com/ColonelBlimp/goft8/docs/design.md`. This file is the
    living reference; updates land alongside code changes.

Migration from `go-ft8/research/` begins after the user's initial
scaffolding state is committed (to give a clean `git status`).
