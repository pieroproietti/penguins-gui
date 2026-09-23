package main

import (
	"math"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"fyne.io/fyne/v2/theme"
)

func TestScaledThemeSize(t *testing.T) {
	base := theme.DefaultTheme()
	defaultTextSize := base.Size(theme.SizeNameText)

	scales := []float32{0.85, 1.0, 1.15, 1.3, 1.5, 2.0}
	for _, s := range scales {
		th := newScaledTheme(base, s)
		got := th.Size(theme.SizeNameText)
		want := float32(math.Round(float64(defaultTextSize*s)*100) / 100)
		diff := float64(got - want)
		if math.Abs(diff) > 0.05 {
			t.Errorf("scale %v: got %v, want %v", s, got, want)
		}
	}
}

func TestScaledThemeUnwrap(t *testing.T) {
	base := theme.DefaultTheme()
	th1 := newScaledTheme(base, 1.2)
	th2 := newScaledTheme(th1, 1.5)

	st, ok := th2.(*scaledTheme)
	if !ok {
		t.Fatalf("expected *scaledTheme, got %T", th2)
	}
	if _, isInnerScaled := st.base.(*scaledTheme); isInnerScaled {
		t.Errorf("base theme should be unwrapped, but got nested *scaledTheme")
	}
}

func TestScaledThemeClamping(t *testing.T) {
	thLow := newScaledTheme(nil, 0.1)
	if st, ok := thLow.(*scaledTheme); !ok || st.scale != minScale {
		t.Errorf("expected scale clamped to minScale %v, got %v", minScale, st.scale)
	}

	thHigh := newScaledTheme(nil, 5.0)
	if st, ok := thHigh.(*scaledTheme); !ok || st.scale != maxScale {
		t.Errorf("expected scale clamped to maxScale %v, got %v", maxScale, st.scale)
	}
}

func TestScaledThemeColorAndFont(t *testing.T) {
	test.NewApp()
	base := theme.DefaultTheme()
	th := newScaledTheme(base, 1.25)

	if th.Font(fyne.TextStyle{}) == nil {
		t.Errorf("expected non-nil font")
	}

	darkBg := th.Color(theme.ColorNameBackground, theme.VariantDark)
	lightBg := th.Color(theme.ColorNameBackground, theme.VariantLight)
	if darkBg == nil || lightBg == nil {
		t.Errorf("expected valid background colors")
	}
	if darkBg == lightBg {
		t.Errorf("expected dark and light variant backgrounds to differ")
	}
}

func TestPreferenceFontScale(t *testing.T) {
	app := test.NewApp()
	pref := app.Preferences()

	pref.SetFloat(preferenceFontScale, 1.30)
	got := float32(pref.FloatWithFallback(preferenceFontScale, 1.0))
	if math.Abs(float64(got-1.30)) > 0.01 {
		t.Fatalf("expected 1.30, got %v", got)
	}
}

