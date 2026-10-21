+++
Title = "Erste Schritte"
Description = "astwerk installieren, einen Baum schreiben, eine Seite generieren."
Order = 40
Group = "Build a site"
+++

astwerk generiert eine statische Seite aus einem Go-Programm. Du beschreibst
das Ausgabeverzeichnis als einen Baum aus `Node`s; `Build` durchläuft den Baum
und schreibt Dateien. Vorlagen sind
[templ](https://templ.guide)-Komponenten — echtes Go, zur Compilezeit
typprüfbar.

Neu hier? Lies zuerst [Wie astwerk denkt](../how-astwerk-thinks/) — die ganze
Architektur auf einer Seite.

## Voraussetzungen

- **Go 1.27 oder neuer** — die Go-Werkzeugkette ist das Build-System.
- **Die templ-CLI** — kompiliert `.templ`-Dateien zu Go. Das ist die gesamte
  Werkzeugkette: kein Node.js, kein Bundler, kein Plugin-System.
- **Ein aktueller Browser**, um dir die erzeugte Seite anzusehen. Der Kern von
  astwerk gibt reines HTML und CSS aus; Skripte sind nur da, wenn du sie
  hinzufügst.

## Installieren

```sh
go get github.com/LukasDerBaum42/astwerk
go install github.com/a-h/templ/cmd/templ@latest
```

`go get` fügt astwerk der `go.mod` deines Moduls hinzu. Die templ-CLI
kompiliert `.templ`-Dateien zu Go — das ist die ganze Werkzeugkette, kein
Node.js, kein Bundler, kein Plugin-System. Wenn `templ` nicht auf deinem
`PATH` ist, funktioniert die `go run`-Form ohne Installation von etwas:

```sh
go run github.com/a-h/templ/cmd/templ@latest generate
```

## Beispiel

Eine Vorlage, `views/page.templ`:

```templ
package views

templ Page(title string) {
	<main>
		<h1>{ title }</h1>
		<p>Hallo, astwerk!</p>
	</main>
}
```

Und das Programm, das daraus eine Seite macht — `main.go`:

```go
package main

import (
	"log"

	"example.com/site/views"
	"github.com/LukasDerBaum42/astwerk/ssg"
)

func main() {
	root := ssg.Node{
		Title: "Meine Seite",
		Page:  ssg.Templ(views.Page),
	}
	if err := ssg.Build(root, ssg.BuildOptions{}); err != nil {
		log.Fatal(err)
	}
}
```

```sh
templ generate
go run .
```

Das schreibt `build/index.html`. Mehr Seiten, eine Vorlage darum, Markdown
dazu und ein Entwicklungs-Server — siehe
[The Node tree](../the-node-tree/), [Content & collections](../content-and-collections/)
und [Dev server](../dev-server/).
