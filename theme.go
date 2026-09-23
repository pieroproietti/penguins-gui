package main

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

const (
	preferenceFontScale = "ui.font_scale"
	minScale            = float32(0.8)
	maxScale            = float32(2.5)
	defaultScale        = float32(1.0)
	scaleStep           = float32(0.15)
)

type scalePreset struct {
	Label string
	Scale float32
}

var scalePresets = []scalePreset{
	{Label: "Small (85%)", Scale: 0.85},
	{Label: "Default (100%)", Scale: 1.00},
	{Label: "Medium (115%)", Scale: 1.15},
	{Label: "Large (130%)", Scale: 1.30},
	{Label: "Very Large (150%)", Scale: 1.50},
	{Label: "Extra Large (175%)", Scale: 1.75},
	{Label: "Huge (200%)", Scale: 2.00},
}

type scaledTheme struct {
	base  fyne.Theme
	scale float32
}

func newScaledTheme(base fyne.Theme, scale float32) fyne.Theme {
	for {
		if st, ok := base.(*scaledTheme); ok {
			base = st.base
		} else {
			break
		}
	}
	if base == nil {
		base = theme.DefaultTheme()
	}
	if scale < minScale {
		scale = minScale
	} else if scale > maxScale {
		scale = maxScale
	}
	// Round to 2 decimal places to avoid floating point drift.
	scale = float32(math.Round(float64(scale)*100) / 100)
	return &scaledTheme{
		base:  base,
		scale: scale,
	}
}

func (s *scaledTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	return s.base.Color(name, variant)
}

func (s *scaledTheme) Font(style fyne.TextStyle) fyne.Resource {
	return s.base.Font(style)
}

func (s *scaledTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return s.base.Icon(name)
}

func (s *scaledTheme) Size(name fyne.ThemeSizeName) float32 {
	return s.base.Size(name) * s.scale
}

func clampScale(scale float32) float32 {
	if scale < minScale {
		return minScale
	}
	if scale > maxScale {
		return maxScale
	}
	return float32(math.Round(float64(scale)*100) / 100)
}
