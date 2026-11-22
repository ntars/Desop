// Package tokens is the single source of truth for visual and motion values.
// Components must read from here instead of defining their own radii or timing.
package tokens

import "time"

// Tokens contains the shared design values used by every Desop component.
type Tokens struct {
	RadiusSmall  int
	RadiusMedium int
	RadiusLarge  int

	SpacingXS int
	SpacingSM int
	SpacingMD int
	SpacingLG int

	BlurRadius  int
	WindowAlpha uint8
	ShadowAlpha uint8
	BorderLight uint8
	AccentR     uint8
	AccentG     uint8
	AccentB     uint8
	SurfaceR    uint8
	SurfaceG    uint8
	SurfaceB    uint8
	TextR       uint8
	TextG       uint8
	TextB       uint8

	AnimationFast time.Duration
	AnimationNorm time.Duration
	AnimationSlow time.Duration
	FrameInterval time.Duration

	FontFace       string
	FontTitleSize  int
	FontBodySize   int
	FontSmallSize  int
	FontSymbolSize int
}

// Current returns the active token set. Values are deterministic so renderers
// on different monitors can share one source of truth.
func Current() Tokens {
	return Tokens{
		RadiusSmall:  6,
		RadiusMedium: 12,
		RadiusLarge:  22,

		SpacingXS: 4,
		SpacingSM: 8,
		SpacingMD: 12,
		SpacingLG: 20,

		BlurRadius:  18,
		WindowAlpha: 235,
		ShadowAlpha: 90,
		BorderLight: 42,
		AccentR:     99,
		AccentG:     186,
		AccentB:     255,
		SurfaceR:    28,
		SurfaceG:    30,
		SurfaceB:    34,
		TextR:       244,
		TextG:       246,
		TextB:       248,

		AnimationFast: 120 * time.Millisecond,
		AnimationNorm: 220 * time.Millisecond,
		AnimationSlow: 380 * time.Millisecond,
		FrameInterval: 16 * time.Millisecond,

		FontFace:       "Segoe UI",
		FontTitleSize:  28,
		FontBodySize:   14,
		FontSmallSize:  11,
		FontSymbolSize: 20,
	}
}

// AccentRGB returns the accent colour as three 0-255 components.
func (t Tokens) AccentRGB() (uint8, uint8, uint8) { return t.AccentR, t.AccentG, t.AccentB }

// SurfaceRGB returns the surface colour as three 0-255 components.
func (t Tokens) SurfaceRGB() (uint8, uint8, uint8) { return t.SurfaceR, t.SurfaceG, t.SurfaceB }

// TextRGB returns the primary text colour as three 0-255 components.
func (t Tokens) TextRGB() (uint8, uint8, uint8) { return t.TextR, t.TextG, t.TextB }
