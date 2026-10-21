+++
Title = "Vergleich"
Description = "Wie sich astwerk von Hugo, Zola, Eleventy und Astro unterscheidet — und wann eines davon die bessere Wahl ist."
Order = 30
Group = "Introduction"
+++

Dies ist ein Static-Site-Generator unter vielen. Der ehrliche Vergleich ist
der mit denen, zu denen man zuerst greift — auf den Achsen, die einen Build
wirklich verändern: wie Inhalte beschrieben werden, welche Template-Sprache
verwendet wird und wie viel das Werkzeug vom Projekt übernehmen will.

## Die Kurzfassung

Wähle astwerk, wenn das Team bereits Go schreibt, die Templates zur
Compilezeit typprüfbar sein sollen und die Struktur der Seite im Code leben
soll statt in einer Verzeichnis-Konvention. Wähle Hugo, Zola, Eleventy oder
Astro, wenn du ein Theme-Ökosystem, ein Komponenten-Framework für Islands oder
Builds in einer Größenordnung brauchst, in der inkrementelle Kompilierung
zählt.

## Auf einen Blick

| | astwerk | Hugo | Zola | Eleventy | Astro |
| --- | --- | --- | --- | --- | --- |
| Sprache | Go | Go | Rust | JS/Node | JS/Node |
| Templates | templ (kompiliertes Go) | Go `html/template` | Tera | Nunjucks/Liquid/JS | `.astro` (JSX-artig) |
| Seitenstruktur | ein `Node`-Baum, den du schreibst | Konvention: `content/` + `layouts/` | Konvention: `content/` + `templates/` | Konvention + Konfiguration | Datei-Routing in `src/pages` |
| Content-Erkennung | explizit, über `Generate` | Verzeichnis-Scan | Verzeichnis-Scan | Verzeichnis-Scan | Verzeichnis-Scan |
| Client-JS | keins eingebaut; Skripte sind opt-in | keins eingebaut | keins eingebaut | keins eingebaut | Islands, beliebiges Framework |
| Konfigurationsformat | keins — Go-Struct-Literale | TOML/YAML/JSON | TOML | JS/JSON | JS/TS |
| Werkzeugkette | `go` + templ-CLI | eine Binary | eine Binary | Node + npm | Node + npm |

## Wo astwerk gut passt

**Du schreibst bereits Go und willst den Typprüfer auf deinen Templates.**
templ-Komponenten sind Go-Funktionen — ein Tippfehler in einem Variablennamen
ist ein Compile-Fehler, kein leeres Feld in der Produktion. Hugos
`html/template` ist auch Go, aber die Templates sind Strings, die zur Laufzeit
interpretiert werden; ein falscher Feldzugriff scheitert dort im besten Fall
zur Buildzeit, im schlechtesten im Browser.

**Du willst die Struktur der Seite als Code, nicht als Verzeichnis-Konvention,
die der Generator errät.** Hugo, Zola und Eleventy arbeiten alle, indem sie
`content/` scannen und es per Namenskonvention einem Layout zuordnen. Das ist
schnell zum Starten und awkward beim Abweichen — ein ungewöhnliches Routing
bedeutet, gegen die Konvention zu kämpfen. Ein `Node`-Baum hat keine Konvention
zum Bekämpfen; er ist ein Go-Wert, also ist alles ausdrückbar, was in Go
ausdrückbar ist — inklusive eines `Generate`, das Kinder aus einer API-Antwort
oder einer Datenbankabfrage zur Buildzeit berechnet.

**Du willst statische Ausgabe ohne angehängte Client-Laufzeit.** Der Kern von
astwerk liefert kein JavaScript. Eine Seite ohne Skript ist HTML und CSS,
genau wie eine handgeschriebene Seite. Wenn eine Seite Verhalten braucht, füge
ein normales `<script>`-Tag hinzu, eine TypeScript-Datei, die
`CompileScripts` an `tsc` gibt, oder ein wasm-Modul, das `CompileFrom` baut.
Nichts wird in Seiten injiziert, die nicht danach gefragt haben.

### astwerk vs. Astro

Astro ist der nächste philosophische Verwandte — HTML-first, standardmäßig
null JS —, also ist der Kontrast am schärfsten:

| | Astro | astwerk |
| --- | --- | --- |
| Standard-Ausgabe | HTML-first | HTML-first |
| Islands | JS-Islands (React, Vue, Svelte…) | keine im Kern; Skripte und wasm sind opt-in Build-Inputs |
| Werkzeugkette | Node-Ökosystem | Go-Ökosystem |
| Templates | `.astro` (JSX-artig) | templ (kompiliertes Go, typprüfbar) |
| Seitenstruktur | Datei-Routing in `src/pages` | ein `Node`-Baum, den du schreibst |

Wähle Astro, wenn du Komponenten-Islands über mehrere UI-Frameworks in einem
Projekt willst. Wähle astwerk, wenn das Team bereits Go schreibt und der Build
in einer Sprache und einer Werkzeugkette bleiben soll.

Astros Island-Modell ist wirklich besser, wenn eine Seite überwiegend
interaktiv ist und die interaktiven Teile React-/Vue-/Svelte-Komponenten sind,
die du schon hast. Der Kern von astwerk versucht das nicht zu sein. Seine
optionalen Client-Erweiterungen
([astwerk-live](https://github.com/LukasDerBaum42/astwerk-live)) sind separate
Pakete, noch stark im Umbau und nicht Teil des Kernversprechens.

## Die bewussten Lücken

astwerk hat diese Dinge nicht, und das ist eine Entscheidung, keine Lücke:

- **Keine Konfigurationsdatei** — Go-Struct-Literale sind die Konfiguration.
  Es gibt kein Format zu lernen und kein Schema nachzuschlagen.
- **Keine Konventionen** — nichts wird entdeckt. Du schreibst den Baum, also
  gibt es keine Framework-Annahme, gegen die man kämpfen muss, wenn eine Seite
  abweicht.
- **Keine Plugins** — Erweiterbarkeit ist Go. Ein eigener Generator ist eine
  Funktion, kein Interface mit Registry.
- **Keine Themes** — es gibt einen Ordner mit Dateien zum Kopieren und
  Bearbeiten. Kein Marktplatz, keine Vererbung, kein Update, das deine
  Anpassungen bricht.
- **Keine Client-Laufzeit** — der Kern gibt statisches HTML aus. Interaktivität
  ist ein Skript oder ein wasm-Modul, das du bewusst hinzufügst, kein
  Framework, das der Generator injiziert.
- **Kein Node.js** — die Werkzeugkette ist `go` plus die templ-CLI, und das ist
  die ganze Werkzeugkette.

## Wo etwas anderes die bessere Wahl ist

**Du willst ein Plugin-Ökosystem und einen Theme-Marktplatz.** Hugo und
Eleventy haben beide große Ökosysteme mit Drop-in-Themes und Shortcodes.
astwerks `starter/` sind dreizehn Dateien, die du kopierst und bearbeitest —
es gibt keinen Marktplatz, und es soll keinen geben. Wenn „Theme installieren,
Logo ändern“ die ganze Anforderung ist, kommt Hugo schneller ans Ziel.

**Du willst Komponenten-Islands über mehrere UI-Frameworks in einem Projekt.**
Das ist Astros spezifische Stärke — React-, Svelte- und Vue-Islands auf einer
Seite mischen — und nichts, was der Kern von astwerk versucht. Die optionalen
astwerk-live-Pakete sind eine kleine, separate Client-Geschichte, kein Ort, um
ein bestehendes Komponenten-Framework einzustecken — und sie werden gerade
umgebaut.

**Du brauchst inkrementelle/parallele Builds für eine sehr große Seite
(10k+ Seiten).** Hugo ist besonders darauf optimiert. astwerks Walker ist eine
geradlinige Depth-First-Rekursion ohne inkrementellen Cache — jeder Build ist
ein vollständiger Build. Für die meisten Projekt- und Doku-Seiten fällt das
nicht auf; für eine große Content-Seite ist es eine echte Lücke (als mögliche
Verbesserung notiert, heute nicht gelöst).
