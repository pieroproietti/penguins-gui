package builder

import "testing"

func TestFedoraRecipeVersionValidation(t *testing.T) {
	for _, tc := range []struct {
		base, revision string
		valid          bool
	}{
		{"26.9.23", "1", true},
		{"1.2.3.rc.1", "12", true},
		{"", "1", false},
		{"1.0%{lua:print(1)}", "1", false},
		{"1.0\nRequires: injected", "1", false},
		{"1.0", "1%{?dist}", false},
		{"1.0", "0", false},
		{"1.0", "-1", false},
	} {
		_, err := fedoraRecipe(RecipeData{BaseVersion: tc.base, Rel: tc.revision})
		if (err == nil) != tc.valid {
			t.Errorf("fedoraRecipe(%q, %q): %v; want valid=%v", tc.base, tc.revision, err, tc.valid)
		}
	}
}
