+++
Title = "Häufige Fehler"
Description = "Die Dinge, die beißen: die templ-ctx-Kollision, absolute URLs, Map-Iterationsreihenfolge, Parallel-Races."
Order = 240
Group = "Reference"
+++

Eine kurze Liste der Fehler, die echte Zeit kosten, in der Reihenfolge, in der
sie dich wahrscheinlich treffen. Alle sind „funktioniert lokal, bricht dann auf
eine Weise, die man eine Weile nicht benennen kann“-Probleme.

## Nenne einen templ-Parameter `ctx`

templs generierter Code deklariert sein eigenes `ctx context.Context`; ein
Parameter oder eine Variable mit diesem Namen kollidiert und die generierte
Datei kompiliert nicht. Nenne den Parameter `c` und übergib `ssg.Ctx` als
erstes Argument:

```templ
templ Page(c ssg.Ctx) { … }
```

## Absolute URLs als Literale schreiben

Jeder In-Page-Link und jedes Asset muss durch `Ctx.Asset` und `Ctx.Link` gehen.
Ein Literal `/style.css` bricht in dem Moment, in dem die Seite von einem
Basispfad ausgeliefert wird — GitHub Pages zum Beispiel —, denn der Sinn
relativer URLs ist, dass ein Build überall funktioniert:

```templ
<link rel="stylesheet" href={ c.Asset("style/style.css") }/>
<a href={ c.Link("docs/") }>Docs</a>
```

Der Build dieser Seite scheitert an einem absoluten oder toten Link
(`site_test.go` geht jede Seite durch und löst jeden Link auf), der Fehler
fällt hier also schon zur Buildzeit auf — aber deine Seite sollte dieselbe
Gewohnheit übernehmen: Ausgabe bauen und die Links prüfen.

## Über eine `map[string]Page` ohne `Slugs` iterieren

```go
// falsch — die Reihenfolge ist zufällig, der Index mischt sich also bei jedem
// Build neu:
for slug, page := range pages { … }

// richtig — sortiert, deterministisch:
for _, slug := range content.Slugs(pages) { … }
```

Die erste mit astwerk gebaute Seite produzierte byte-identische Ausgabe auf
jeder Seite außer dem Collection-Index, der seine Kacheln bei jedem Lauf
mischte. `Slugs` existiert, damit dieser Fehlermodus einen expliziten Aufruf
entfernt ist.

## Geteilten Zustand in `Generate`-Closures unter `Parallel: true` mutieren

`Parallel: true` verschiebt deine `Generate`- und `Page`-Funktionen auf andere
Goroutinen. Eine templ-Komponente zu rendern ist sicher; eine
`Generate`-Closure, die an eine eingefangene Slice anhängt, nicht. Entscheide
dich bewusst für `Parallel` und behandle geteilten Zustand so, wie du es in
jedem nebenläufigen Programm tun würdest.

## `Generate` und `Children` denselben Schlüssel geben

Generierte Schlüssel gewinnen — still. Wenn ein `Generate` einen Schlüssel
zurückgibt, der auch in `Children` existiert, ersetzt der generierte Knoten den
handgeschriebenen. Benenne sie bewusst und teste die Ausgabe.

## Annehmen, dass eine Seite JavaScript ausliefert

Der Kern von astwerk gibt HTML und CSS aus. Eine Seite hat nur dann
JavaScript, wenn du ein `<script>`-Tag oder einen `CompileFrom`-Knoten
hinzufügst. Wenn eine generierte Seite ein Skript enthält, das du nicht
geschrieben hast, suche die Vorlage oder Komponente, die es injiziert hat.
