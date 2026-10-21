package tutorial

import "github.com/LukasDerBaum42/astwerk/content"

// samplePostHTML renders the step-4 sample post with the site's own content
// pipeline at build time — front matter stripped, markdown rendered — so what
// the reader writes is literally what the framework renders.
var samplePostHTML = renderSamplePost()

func renderSamplePost() string {
	page, err := content.Parse(samplePostSource)
	if err != nil {
		return "<p>failed to render sample post</p>"
	}
	return page.HTML
}
