+++
Title = "Vergleich"
Description = "Wie sich astwerk von Hugo, Zola, Eleventy, Astro und Svelte unterscheidet — und wann eines davon die bessere Wahl ist."
Order = 30
Group = "Introduction"
+++

Dies ist ein Static-Site-Generator unter vielen. Der ehrliche Vergleich ist
der mit den Werkzeugen, nach denen man zuerst greift — auf den Achsen, die
einen Build wirklich verändern: wie Inhalte beschrieben werden, was die
Templating-Sprache ist und was clientseitige Interaktivität kostet.

## Die Kurzfassung

Wähle astwerk, wenn das Team bereits Go schreibt, die Vorlagen zur
Kompilierzeit typprüfbar sein sollen, die Struktur der Seite in Code lebt statt
in einer Verzeichnis-Konvention, und Interaktivität keine JavaScript-Toolchain
mitbringen soll. Wähle Hugo, Zola, Eleventy oder Astro, wenn du eine
Theme-Ökosystem, ein Komponenten-Framework für Islands oder Builds in einer
Größenordnung brauchst, bei der inkrementelles Kompilieren zählt. Wähle Svelte,
wenn du zur Kompilierzeit aufgelöste Reaktivität willst — die Idee, auf der
astwerks `x` aufbaut —, aber die Komponentenschicht in
JavaScript/TypeScript.

## Auf einen Blick

| | astwerk | Hugo | Zola | Eleventy | Astro | Svelte |
| --- | --- | --- | --- | --- | --- | --- |
| Sprache | Go | Go | Rust | JS/Node | JS/Node | JS/TS |
| Vorlagen | templ (kompiliertes Go) | Go `html/template` | Tera | Nunjucks/Liquid/JS | `.astro` (JSX-artig) | `.svelte` (kompiliert) |
| Seitenstruktur | ein `Node`-Baum, den du schreibst | Konvention: `content/` + `layouts/` | Konvention: `content/` + `templates/` | Konvention + Konfiguration | dateibasiertes Routing in `src/pages` | SvelteKit-Routen in `src/routes` |
| Inhaltsfindung | explizit, über `Generate` | Verzeichnis-Scan | Verzeichnis-Scan | Verzeichnis-Scan | Verzeichnis-Scan | SvelteKit-Load-Funktionen |
| Client-JS | Go → WASM, opt-in | keins eingebaut | keins eingebaut | keins eingebaut | Islands, beliebiges Framework | zur Buildzeit kompiliert, kein Runtime-Framework |
| Konfiguration | keine — Go-Struct-Literale | TOML/YAML/JSON | TOML | JS/JSON | JS/TS | `svelte.config.js` |
| Toolchain | `go` + templ-CLI | einzelnes Binary | einzelnes Binary | Node + npm | Node + npm | Node + npm/Vite |

## Wo astwerk gut passt

**Du schreibst bereits Go und willst den Typprüfer auf deinen Vorlagen.**
templ-Komponenten sind Go-Funktionen — ein Tippfehler in einem Variablennamen
ist ein Kompilierfehler, keine leere Stelle in Produktion. Hugos
`html/template` ist auch Go, aber seine Vorlagen sind Strings, die zur Laufzeit
interpretiert werden; ein falscher Feldzugriff scheitert dort bestenfalls zur
Buildzeit, schlimmstenfalls im Browser.

**Du willst, dass die Struktur der Seite Code ist, keine Verzeichnis-Konvention,
die der Generator erraten muss.** Hugo, Zola und Eleventy scannen `content/`
und ordnen es einer Vorlage per Namenskonvention zu. Das ist schnell
losgelegt, aber unbequem beim Abweichen — ein ungewöhnliches Routing-Bedürfnis
bedeutet, gegen die Konvention zu kämpfen. Ein `Node`-Baum hat keine
Konvention, gegen die man kämpfen müsste; er ist ein Go-Wert, also ist alles,
was in Go ausdrückbar ist, auch im Baum ausdrückbar — einschließlich eines
`Generate`, das Kinder zur Buildzeit aus einer API-Antwort oder einer
Datenbankabfrage berechnet.

**Du willst Interaktivität, ohne ein JS-Framework zu übernehmen.** Astros
Islands sind der nächste Vergleich — JS nur dort ausliefern, wo eine Komponente
es braucht —, aber die Island selbst ist immer noch React, Vue, Svelte oder
ähnliches, mit dem eigenen Build-Schritt dieses Frameworks. astwerks
interaktive Wege sind [`x`](../x/) — kompiliertes Markup, schlichtes JavaScript
zur Buildzeit, gar kein Framework — und `reactive`, Go kompiliert zu WASM:
eine Sprache, eine Toolchain, keine `npm install`-Spaltung zwischen der
statischen und der interaktiven Hälfte einer Seite.

### astwerk vs. Astro

Astro ist der nächste philosophische Verwandte — HTML zuerst, Islands,
Null-JS als Standard —, daher ist der Kontrast am schärfsten:

| | Astro | astwerk |
| --- | --- | --- |
| Standard-Ausgabe | HTML zuerst | HTML zuerst |
| Islands | JS-Islands (React, Vue, Svelte…) | kompilierte Go/JS-Islands (`x`) oder Go/WASM (`reactive`) |
| Toolchain | Node-Ökosystem | Go-Ökosystem |
| Vorlagen | `.astro` (JSX-artig) | templ (kompiliertes Go, typprüfbar) |
| Seitenstruktur | dateibasiertes Routing in `src/pages` | ein `Node`-Baum, den du schreibst |

Wähle Astro, wenn du Komponenten-Islands über mehrere UI-Frameworks in einem
Projekt willst. Wähle astwerk, wenn das Team bereits Go schreibt und die
interaktive Hälfte der Seite keinen JS-Framework-Build-Schritt mitbringen soll.

### astwerk vs. Svelte

Svelte ist der nächste Verwandte von astwerks [`x`](../x/) — und seine
Inspiration. Beide haben dieselbe Entscheidung getroffen: Reaktivität ist ein
Kompilierzeit-Problem, also liefere die Ausgabe des Compilers aus statt eines
Frameworks. Der Unterschied ist, wo der Compiler lebt und was er kompiliert.

| | Svelte | astwerk `x` |
| --- | --- | --- |
| Reaktivität | zur Buildzeit kompiliert | zur Buildzeit kompiliert |
| Laufzeit | kleine Laufzeit, kein virtuelles DOM | Fragment-Runtime pro Seite, kein virtuelles DOM |
| Komponenten | `.svelte`-Dateien, eine Komponenten-Syntax | gewöhnliche Go-Funktionen in templ |
| Sprache | JavaScript/TypeScript | Go |
| Toolchain | Node + Vite/SvelteKit | `go` + die templ-CLI |

Svelte kompiliert eine Komponentendatei in Vanilla-JavaScript; `x` kompiliert
die Marker, die eine Seite tatsächlich nutzt, in ein kleines Skript — pro Seite
aus Fragmenten zusammengesetzt, und nur, wenn eine Seite mit `x.Document`
opt-in geht. Beides sind Antworten auf „warum ein Framework ausliefern, wenn
der Compiler die Arbeit tun kann?“. Der Unterschied: astwerk hält den ganzen
Stack in einer Sprache — kein `.svelte`-Dateiformat, kein zweites Ökosystem,
keine Node-Toolchain im Build. Wähle Svelte, wenn die Komponentenschicht in
JavaScript/TypeScript mit eigenem Ökosystem leben soll; wähle astwerk, wenn
sie Go sein soll, in deinen bestehenden templ-Komponenten.

## Die bewussten Absenzen

astwerk hat diese Dinge nicht — und das ist eine Entscheidung, keine Lücke:

- **Keine Konfigurationsdatei** — Go-Struct-Literale sind die Konfiguration. Es
  gibt kein Format zu lernen und kein Schema nachzuschlagen.
- **Keine Konventionen** — nichts wird entdeckt. Du schreibst den Baum, es gibt
  also keine Framework-Annahme, gegen die man kämpfen müsste, wenn eine Seite
  abweicht.
- **Keine Plugins** — Erweiterbarkeit ist Go. Ein eigener Generator ist eine
  Funktion, kein Interface zum Implementieren und keine Registry zum
  Registrieren.
- **Keine Themes** — es gibt einen Ordner mit Dateien, die du kopierst und
  editierst. Kein Marktplatz, keine Vererbung, kein Update, das deine
  Anpassungen bricht.
- **Kein virtuelles DOM** — Bindings patchen jeweils eine Eigenschaft, es wird
  also nie etwas hinter deinem Rücken zerstört und neu aufgebaut.
- **Kein Node.js** — die Toolchain ist `go` plus die templ-CLI, und das ist die
  ganze Toolchain.

## Wo etwas anderes die bessere Wahl ist

**Du willst ein Plugin-Ökosystem und einen Theme-Marktplatz.** Hugo und
Eleventy haben große Ökosysteme aus fertigen Themes und Shortcodes. astwerks
`starter/` sind dreizehn Dateien, die du kopierst und editierst — es gibt
keinen Marktplatz, und es soll keinen geben. Wenn „Theme installieren, Logo
ändern“ die ganze Anforderung ist, ist Hugo schneller dort.

**Du willst Go auf dem Client und nichts anderes macht es so billig.** Das
braucht eine Korrektur statt einer Einschränkung: astwerk verlangt kein WASM.
Eine Seite ohne `CompileFrom`-Knoten liefert genau so viel Client-JS aus wie
eine Hugo- oder Zola-Seite — nämlich null —, und `CompileScripts` reicht
`.ts`-Dateien im selben Build-Schritt an `tsc` weiter, sodass ein Skript in
schlichtem JavaScript oder TypeScript nichts über das hinaus kostet, was es auf
jedem anderen Generator kosten würde. Das ~2-MB-WASM-Minimum ist real, aber es
ist der Preis einer bestimmten Wahl — die Logik dieses Skripts in Go zu
schreiben und `reactive` oder `wasmwrap` zu nutzen —, keine Steuer, die astwerk
auf jeder Seite erhebt. Greif danach, wenn die interaktive Logik substanziell
genug ist, dass das Teilen einer Sprache mit dem Rest der Seite die Laufzeit
einspielt; greif zu einem `<script>`-Tag oder einer `.ts`-Datei, wie überall
sonst auch, wenn nicht.

**Das Team ist bereits fließend in JS/TS und will Komponenten-Islands über
mehrere UI-Frameworks in einem Projekt.** Das ist Astros spezifische Stärke —
React-, Svelte- und Vue-Islands auf einer Seite zu mischen — und das versucht
astwerk nicht zu tun. `reactive` ist eine kleine, feingranulare
State-Bibliothek, kein Ort, an dem man ein bestehendes Komponenten-Framework
einsteckt.

**Du willst Reaktivität zur Kompilierzeit, aber die Komponentenschicht in
JavaScript.** Das ist exakt Sveltes Angebot — und die Idee, auf der astwerks
`x` aufbaut. Wenn das Team JS/TS-first ist und das Komponenten-Ökosystem um
Svelte (SvelteKit, Stores, Transitions) will, ist Svelte die bessere Wahl;
wenn die Komponenten Go-Funktionen in templ sein sollen, hält astwerk den
ganzen Stack in einer Sprache — siehe [astwerk vs. Svelte](#astwerk-vs-svelte)
oben.

**Du brauchst inkrementelle/parallele Builds für eine sehr große Seite
(10.000+ Seiten).** Besonders Hugo ist darauf hart getunt. astwerks Walker ist
eine geradlinige Tiefensuche ohne Inkremental-Cache — jeder Build ist ein
voller Build. Für die meisten Projekt- und Doku-Seiten ist das unmerklich; für
eine große Content-Seite ist es eine echte Lücke (als mögliche Verbesserung
notiert, heute nicht gelöst).
