+++
Title = "Keine Abstraktion ist ein Gefängnis"
Description = "Das Design-Gesetz in der Praxis: jede Bequemlichkeit hat einen Weg zurück zu normalem Code."
Order = 20
Group = "Introduction"
+++

Das Design-Gesetz lautet **abstrahiere nur den sich wiederholenden
Boilerplate**. Seine andere Hälfte ist das Prinzip, das ihm Zähne gibt: Eine
Abstraktion soll Boilerplate entfernen, *nicht Fähigkeiten*. Alles, was ein
Marker oder ein Helfer nicht ausdrücken kann, sagst du in normalem Code — das
Framework wird niemals eine Wand.

## Die Leiter

```go
// Bequemlichkeit: der übliche Fall, eine Zeile
"about": {Title: "About", Page: ssg.Templ(views.About)},

// Notausgang: dieselbe Seite, ausgeschrieben
"about": {Title: "About", Page: func(c ssg.Ctx) templ.Component {
	return views.About(c.Title, c.Path, c.Prefix)
}},

// ultimativer Notausgang: die Seite ist einfach Go
"about": {Title: "About", Page: func(c ssg.Ctx) templ.Component {
	return templ.Raw(renderAboutWithAnything(c))
}},
```

Jede Sprosse ist ein Weg nach unten, keine Klippe. `ssg.Templ` passt auf die
übliche Signatur und hält sich raus; ein `Page`-Closure bekommt genau den
Kontext, den du brauchst; und wenn die Seite gar kein HTML ist, ist sie eine
`templ.Component`, die du selbst gebaut hast.

## Dieselbe Form, überall

Die Leiter ist nicht auf Seiten beschränkt — sie ist das Muster des ganzen
Frameworks:

| Bequemlichkeit | Notausgang |
| --- | --- |
| `ssg.Templ(views.Page)` | ein `Page`-Closure, das `ssg.Ctx` direkt entgegennimmt |
| `content.LoadCollection[T]` | `LoadDir` + `Decode` + `Slugs` + `Generate` von Hand |
| `Node.Generate` | ein handgeschriebener `Generate`-Closure, der alles kann, was Go kann |
| `CopyFrom` | die Dateien in eigenem Go-Code kopieren oder generieren |
| Markdown-Seiten | rohes HTML im Markdown oder eine templ-Komponente |
| `starter/`-Vorlagen | die Dateien kopieren und bearbeiten — sie sind dein Code |

## Warum das wichtig ist

Ein Framework ohne Notausgang muss für jedes Nutzerbedürfnis eine Funktion
dazubauen — so werden Frameworks zu Sprachen. astwerks Antwort ist das
Gegenteil: Der Markersatz bleibt bewusst klein, und alles darüber hinaus ist
normales Go, normales HTML, normales templ. Die Grenze ist dokumentiert,
nicht bewacht.

> **Eine Abstraktion soll Boilerplate entfernen, nicht Fähigkeiten.**

Wenn du dich jemals dabei ertappst, gegen das Framework zu kämpfen — eine
Seite, die der Baum nicht ausdrücken kann, ein Helfer, der nicht passt —
dann ist das nicht das Ende des Weges, sondern der Punkt, an dem du eine Ebene
tiefer gehen sollst. Das ist Absicht.
