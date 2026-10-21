+++
Title = "Wie astwerk denkt"
Description = "Ein Go-Programm beschreibt das ganze Ausgabeverzeichnis als Baum; Build schreibt es. Statisches HTML, keine Client-Laufzeit im Kern."
Order = 10
Group = "Introduction"
+++

astwerk generiert eine statische Seite aus einem Go-Programm. Es gibt keine
Konfigurationsdatei, keinen Konventions-Scanner, keine Plugin-Pipeline — ein
Programm, und alles andere folgt daraus, dass dieses Programm normales Go ist.

## Die ganze Pipeline

```
Dein Go-Programm
      │
      ▼
   Node-Baum        ← du schreibst ihn: das Ausgabeverzeichnis als Daten
      │
      ▼
   Build()
      ├── HTML          Seiten und Layouts (templ-Komponenten)
      ├── assets        kopierte Verzeichnisse
      ├── scripts       kompiliert zu WebAssembly oder TypeScript
      └── generated     zur Buildzeit berechnete Seiten (Content-Collections)
```

Du beschreibst das Ausgabeverzeichnis als einen Baum aus `Node`s und rufst
`Build` auf. Der Baum ist die Seite — nichts wird entdeckt, nichts wird
abgeleitet. Was du schreibst, ist genau das, was geschrieben wird.

## Die Client-Seite ist deine Wahl

Der Kern von astwerk gibt statisches HTML aus. Er liefert keine
Client-Laufzeit, kein Framework und kein eigenes JavaScript. Wenn eine Seite
Verhalten braucht, ist das ein normales `<script>`-Tag, eine `.ts`-Datei, die
`CompileScripts` an `tsc` gibt, oder eine Go-Datei, die
[`CompileFrom`](../scripts-and-wasm/) zu WASM kompiliert — alles gewöhnliche
Dateien im Baum.

Die optionalen Client-Pakete leben im separaten
[astwerk-live](https://github.com/LukasDerBaum42/astwerk-live)-Projekt:
kompiliertes Markup (`x`), Signale und Bindings (`reactive`), direkter
DOM-Zugriff (`wasmwrap`) und ein Browser-DOM für Tests (`jsdom`). Diese Seite
wird gerade stark umgebaut und wird sich sehr wahrscheinlich ändern oder
verschwinden. Die Seite [Client-Werkzeuge](../which-tool/) ist der ehrliche
Einstieg.

## Die drei Mechanismen

Alles, was das Framework tut, passt unter drei Überschriften, jede in einer
Sitzung lesbar (ausführlich auf [How it works](../how-it-works/)):

| Mechanismus | Was er tut |
| --- | --- |
| Der Baum-Walker | `Build` durchläuft den Baum rekursiv und schreibt Dateien — depth-first, in sortierter Schlüsselreihenfolge, reproduzierbar |
| Locale-Ableitung | `BuildLocales` spiegelt den Baum pro Sprache; du überschreibst nur, was wirklich anders ist |
| Content-Laden | `content` macht aus Markdown und Front Matter typisierte Items für `Generate` |

## Was das für dich bedeutet

Weil der Baum Daten in einem Go-Programm sind, steht zur Buildzeit alles zur
Verfügung, was die Sprache kann: Kinder aus einer Datenbank, einer API oder
einem Markdown-Verzeichnis berechnen; Seiten in einer Schleife generieren;
einen ganzen Teilbaum bedingt einfügen. Und weil Seiten templ-Komponenten
sind, läuft der Typprüfer über deine Vorlagen.

Die Kehrseite: Nichts passiert per Konvention. Wenn eine Seite etwas braucht,
das der Baum nicht ausdrücken kann, fällt es auf reines Go zurück —
[niemals eine Wand, an der es nicht vorbeikommt](../escape-hatches/).
