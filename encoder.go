package goft8

// NTXSamples is the number of 12 kHz mono samples in a single FT8
// transmission (12.64 s × 12 kHz).
const NTXSamples = 151680

// Encoder generates the 12 kHz audio waveform for an FT8 message, for
// TX-side use. v0.1 ships the API shape only so that callers can
// import and reference goft8.Encoder immediately; the transmit path
// will be implemented in v0.2.
type Encoder struct {
	cfg encoderConfig
}

type encoderConfig struct {
	txFreq float64
}

func defaultEncoderConfig() encoderConfig {
	return encoderConfig{txFreq: 1500}
}

// EncoderOption configures an Encoder at construction time.
type EncoderOption func(*encoderConfig)

// WithTxFreq sets the carrier frequency in Hz. Defaults to 1500.
func WithTxFreq(hz float64) EncoderOption {
	return func(c *encoderConfig) { c.txFreq = hz }
}

// NewEncoder creates an Encoder with the given options.
func NewEncoder(opts ...EncoderOption) *Encoder {
	cfg := defaultEncoderConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Encoder{cfg: cfg}
}

// Encode generates the GFSK-modulated waveform for a single FT8
// message. Returns NTXSamples samples of 12 kHz mono PCM.
//
// v0.1: this method panics with a "not yet implemented" message.
// v0.2: full implementation.
func (e *Encoder) Encode(msg string) ([]float32, error) {
	panic("goft8.Encoder.Encode: not yet implemented, scheduled for v0.2")
}
