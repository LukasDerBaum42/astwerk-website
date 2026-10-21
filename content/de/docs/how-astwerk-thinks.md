+++
Title = "Wie astwerk denkt"
Description = "Ein Go-Programm beschreibt das ganze Ausgabeverzeichnis als Baum; Build schreibt es. Interaktivität ist eine Wahl: winziges kompiliertes JavaScript oder volles Go im Browser."
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

## Und dann die Client-Seite

Statisches HTML ist der Standard und bleibt statisch, bis du dich aktiv
dafür entscheidest:

```
Statisches HTML
    │
    ├── x        → winziges generiertes JavaScript: ein bisschen Interaktivität
    │
    └── reactive → volles Go im Browser (über WASM)
```

- **`x`** kompiliert ein eingeschränktes reaktives Modell zur Buildzeit in ein
  kleines Skript. Ein Zähler, ein Umschalter, ein Formular, eine keyed list —
  die üblichen 90 % — kosten ein paar Kilobyte generiertes JavaScript und
  keinen WASM-Download.
- **`reactive`** ist dasselbe Modell als echtes Go. Wenn die interaktive Hälfte
  einer Seite *ein Programm* ist — Goroutinen, die Standardbibliothek, ein
  Router —, dann ist das der WASM-Weg, mit `wasmwrap` darunter als roher
  DOM-Layer.

Die beiden sind [zwei Implementierungen eines reaktiven Modells](../x/):
gleiche Konzepte, andere Laufzeit. [Welches Werkzeug brauche ich?](../which-tool/)
ist die Entscheidungstabelle.

## Die drei Mechanismen

Alles, was das Framework tut, passt unter drei Überschriften, jede in einer
Sitzung lesbar (ausführlich auf [How it works](../how-it-works/)):

| Mechanismus | Was er tut |
| --- | --- |
| Der Baum-Walker | `Build` durchläuft den Baum rekursiv und schreibt Dateien — depth-first, in sortierter Schlüsselreihenfolge, reproduzierbar |
| Locale-Ableitung | `BuildLocales` spiegelt den Baum pro Sprache; du überschreibst nur, was wirklich anders ist |
| Abhängigkeitsverfolgung | Signale merken sich, wer sie liest, sodass Bindings genau das aktualisieren, was ihnen gehört |

## Was das für dich bedeutet

Weil der Baum Daten in einem Go-Programm sind, steht zur Buildzeit alles zur
Verfügung, was die Sprache kann: Kinder aus einer Datenbank, einer API oder
einem Markdown-Verzeichnis berechnen; Seiten in einer Schleife generieren;
einen ganzen Teilbaum bedingt einfügen. Und weil Seiten templ-Komponenten
sind, läuft der Typprüfer über deine Vorlagen.

Die Kehrseite: Nichts passiert per Konvention. Wenn eine Seite etwas braucht,
das der Baum nicht ausdrücken kann, fällt es auf reines Go zurück —
[niemals eine Wand, an der es nicht vorbeikommt](../escape-hatches/).
