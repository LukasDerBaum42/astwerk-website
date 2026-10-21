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
// Framework-Abstraktion
@count

// Notausgang
<script> … </script>

// ultimativer Notausgang
syscall/js
```

Jede Sprosse ist ein Weg nach unten, keine Klippe. `x` kompiliert die üblichen
Fälle; wenn ein Ausdruck aufhört, sich wie das JavaScript zu lesen, das er
erzeugt, schreibst du das JavaScript. Wenn JavaScript nicht reicht, kapselt
`wasmwrap` `syscall/js`, damit Go im Browser läuft. Niemand sitzt auf der
obersten Sprosse fest.

## Dieselbe Form, überall

Die Leiter ist nicht auf Interaktivität beschränkt — sie ist das Muster des
ganzen Frameworks:

| Bequemlichkeit | Notausgang |
| --- | --- |
| `x.Text(signal)`, `x.When`, `x.List` | ein `<script>`, das die benannten Signale liest |
| `content.LoadCollection[T]` | `LoadDir` + `Decode` + `Slugs` + `Generate` von Hand |
| `Node.Generate` | ein handgeschriebener `Generate`-Closure, der alles kann, was Go kann |
| `starter/`-Vorlagen | die Dateien kopieren und bearbeiten — sie sind dein Code |
| `ssg.Templ`-Adapter | ein `Page`-Closure, das `ssg.Ctx` direkt entgegennimmt |

## Warum das wichtig ist

Ein Framework ohne Notausgang muss für jedes Nutzerbedürfnis eine Funktion
dazubauen — so werden Frameworks zu Sprachen. astwerks Antwort ist das
Gegenteil: Der Markersatz bleibt bewusst klein, und alles darüber hinaus ist
normales Go, normales HTML, normales JavaScript. Die Grenze ist dokumentiert,
nicht bewacht.

> **Eine Abstraktion soll Boilerplate entfernen, nicht Fähigkeiten.**

Wenn du dich jemals dabei ertappst, gegen das Framework zu kämpfen — ein
Widget, das nicht kompiliert, eine Seite, die der Baum nicht ausdrücken kann —
dann ist das nicht das Ende des Weges, sondern der Punkt, an dem du eine Ebene
tiefer gehen sollst. Das ist Absicht.
