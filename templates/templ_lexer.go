package templates

import (
	_ "embed"
	"testing/fstest"

	chroma "github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

//go:embed templ_lexer.xml
var templXML []byte

// TemplLexer highlights templ source: markup through the HTML lexer, with the
// Go bits — declarations, templ headers, { expressions }, @component calls and
// if/for blocks — through the Go lexer. Registering it makes ```templ fences
// in the markdown docs resolve to it too.
var TemplLexer = lexers.Register(chroma.DelegatingLexer(
	lexers.HTML,
	chroma.MustNewXMLLexer(fstest.MapFS{"templ.xml": &fstest.MapFile{Data: templXML}}, "templ.xml").SetConfig(&chroma.Config{
		Name:      "templ",
		Aliases:   []string{"templ"},
		Filenames: []string{"*.templ"},
	}),
))
