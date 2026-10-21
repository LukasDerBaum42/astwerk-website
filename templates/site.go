package templates

import (
	"strings"

	x "github.com/LukasDerBaum42/astwerk-live/x"
	"github.com/LukasDerBaum42/astwerk/ssg"
)

// Translated lists the site paths that exist in German, keyed by Ctx.Rel ("" is
// the home page, "docs/getting-started/" a docs page). Any other page shows an
// "English only" banner under the de locale. Set once per build by main.
//
// The site is itself the i18n demo: English is the base tree, German is a
// BuildLocales subtree, and only these pages — plus the shell — are actually
// translated.
var Translated = map[string]bool{}

// Version is the astwerk version the site is built against, read from go.mod —
// the pinned tag, not a working tree a go.work might redirect to.
var Version = "dev"

// MinifiedX is true for production builds, when every page's x runtime ships
// compact rather than readable. Set once per build by main; pages pass
// XOptions() to their x.Document calls.
var MinifiedX = false

// XOptions returns the x.Document options for this build: x.Minified() in
// production, nothing in dev — the exact production recipe the docs describe.
func XOptions() []x.Option {
	if MinifiedX {
		return []x.Option{x.Minified()}
	}
	return nil
}

// label picks the German string under the de locale.
func label(c ssg.Ctx, en, de string) string {
	if c.Locale == "de" {
		return de
	}
	return en
}

// searchPrefix is the relative path from the current page up to the site root:
// "" on the home page, "../../" from a page two directories deep. Search
// results are stored as site-relative paths, so the box prefixes each hit with
// this to build a link that resolves from wherever the reader is.
func searchPrefix(c ssg.Ctx) string {
	if rel := strings.Trim(c.Path, "/"); rel != "" {
		return strings.Repeat("../", strings.Count(rel, "/")+1)
	}
	return ""
}
