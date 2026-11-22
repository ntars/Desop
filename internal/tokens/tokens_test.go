package tokens

import "testing"

func TestCurrentIsComplete(t *testing.T) {
	tk := Current()
	if tk.RadiusSmall <= 0 || tk.RadiusSmall >= tk.RadiusMedium || tk.RadiusMedium >= tk.RadiusLarge {
		t.Fatalf("radius order invalid: %+v", tk)
	}
	if tk.SpacingXS <= 0 || tk.SpacingXS >= tk.SpacingSM || tk.SpacingSM >= tk.SpacingMD || tk.SpacingMD >= tk.SpacingLG {
		t.Fatalf("spacing order invalid: %+v", tk)
	}
	if tk.BlurRadius <= 0 || tk.WindowAlpha == 0 || tk.AnimationFast <= 0 || tk.AnimationFast >= tk.AnimationNorm || tk.AnimationNorm >= tk.AnimationSlow {
		t.Fatalf("visual or motion token invalid: %+v", tk)
	}
	if tk.FontFace == "" || tk.FontTitleSize <= tk.FontBodySize || tk.FontBodySize <= tk.FontSmallSize {
		t.Fatalf("font tokens invalid: %+v", tk)
	}
}

func TestCurrentIsStable(t *testing.T) {
	if Current() != Current() {
		t.Fatal("design tokens must be stable across calls")
	}
}
