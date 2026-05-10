package cards

import (
	"strings"
	"testing"
)

func TestGenerate_ContainsExpectedFields(t *testing.T) {
	d := Data{
		Author: "octocat",
		Title:  "feat: add merge card generation",
		Number: 42,
		Repo:   "octocat/hello-world",
	}

	svg, err := Generate(d)
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}

	checks := map[string]string{
		"author":  "@octocat",
		"title":   "feat: add merge card generation",
		"number":  "#42",
		"repo":    "octocat/hello-world",
		"svg tag": "<svg ",
	}
	for name, want := range checks {
		if !strings.Contains(svg, want) {
			t.Errorf("generated SVG missing %s: expected to contain %q", name, want)
		}
	}
}

func TestGenerate_ValidSVGRoot(t *testing.T) {
	svg, err := Generate(Data{Author: "a", Title: "t", Number: 1, Repo: "o/r"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(svg, "<svg ") {
		t.Errorf("expected SVG to start with <svg, got: %.40s", svg)
	}
	if !strings.HasSuffix(svg, "</svg>") {
		t.Errorf("expected SVG to end with </svg>, got suffix: %.40s", svg[max(0, len(svg)-40):])
	}
}

func TestGenerate_HTMLEscapesSpecialChars(t *testing.T) {
	// A PR title with characters that must be HTML-escaped in an SVG context.
	d := Data{
		Author: "attacker",
		Title:  `<script>alert("xss")</script>`,
		Number: 1,
		Repo:   "a/b",
	}

	svg, err := Generate(d)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The raw script tag must NOT appear in the output.
	if strings.Contains(svg, "<script>") {
		t.Error("raw <script> tag found in SVG output — HTML escaping is broken")
	}
	// The escaped form should be present.
	if !strings.Contains(svg, "&lt;script&gt;") {
		t.Error("expected HTML-escaped title in SVG output")
	}
}

func TestGenerate_EmptyFields(t *testing.T) {
	// Generating with zero-value Data should not panic or error.
	_, err := Generate(Data{})
	if err != nil {
		t.Errorf("unexpected error for zero-value Data: %v", err)
	}
}

func TestGenerate_PRNumberZero(t *testing.T) {
	svg, err := Generate(Data{Author: "dev", Title: "some title", Number: 0, Repo: "x/y"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(svg, "#0") {
		t.Error("expected #0 in SVG for zero PR number")
	}
}

// max is a small helper to avoid importing slices just for bounds clamping.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
