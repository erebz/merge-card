// Package cards generates SVG merge-reward cards from an embedded template.
//
// The card template lives in ../../assets/card.svg.tmpl and is compiled into
// the binary via go:embed, so no external file access is required at runtime.
package cards

import (
	"bytes"
	_ "embed"
	"html/template"
	"strings"
)

//go:embed assets/card.svg.tmpl
var cardTemplate string

// Data holds the values interpolated into the SVG card template.
type Data struct {
	// Author is the GitHub login of the pull request author.
	Author string
	// Title is the pull request title. Long titles are not truncated by this
	// package; the SVG clip-path handles visual overflow.
	Title string
	// Number is the pull request number (e.g. 42).
	Number int
	// Repo is the "owner/repo" string shown at the bottom of the card.
	Repo string
}

// Generate renders the SVG card template with the provided data and returns
// the resulting SVG as a string. An error is returned only if the template
// fails to parse or execute (which should not happen in production since the
// template is embedded and validated by tests).
func Generate(d Data) (string, error) {
	// html/template is used instead of text/template so that values are
	// HTML-escaped, preventing SVG injection via PR titles or author names.
	tmpl, err := template.New("card").Parse(cardTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, d); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
