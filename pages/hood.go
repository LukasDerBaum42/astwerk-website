package pages

import (
	"bytes"
	"context"
	"strings"

	x "github.com/LukasDerBaum42/astwerk/x"
)

// hoodOutput renders hoodDemo through x.Document and splits what the build
// emitted into the stamped HTML and the generated script (runtime elided), so
// the "under the hood" page can show its own output. It renders the widget
// again each call — that is the point: the output is deterministic, and the
// build can always show exactly what it generated.
func hoodOutput() (body, script string) {
	var buf bytes.Buffer
	if err := x.Document(hoodDemo()).Render(context.Background(), &buf); err != nil {
		return "", ""
	}
	s := buf.String()

	open := strings.Index(s, "<script>")
	firstClose := strings.Index(s, "</script>")
	lastOpen := strings.LastIndex(s, "<script>")
	lastClose := strings.LastIndex(s, "</script>")
	if open < 0 || firstClose < 0 || lastOpen < firstClose || lastClose < 0 {
		return "", ""
	}

	head := s[open+len("<script>") : firstClose]
	foot := s[lastOpen+len("<script>") : lastClose]
	body = strings.TrimSpace(s[firstClose+len("</script>") : lastOpen])

	decls := head
	if i := strings.Index(decls, "})();"); i >= 0 {
		decls = "// …the runtime — the fragments this page uses, assembled.\n" + strings.TrimSpace(decls[i+len("})();"):])
	}
	script = strings.TrimSpace(decls + "\n\n// …after the markup:\n" + strings.TrimSpace(foot))
	return body, script
}

func hoodHTML() string {
	body, _ := hoodOutput()
	return body
}

func hoodJS() string {
	_, script := hoodOutput()
	return script
}
