+++
Title = "Deployment"
Description = "GitHub Pages, Netlify, Cloudflare Pages — die BaseURL-/RelativeURLs-Kombination für jeden Host."
Order = 250
Group = "Deploy"
+++

astwerk baut ein Verzeichnis statischer Dateien, deployt also überall hin, wo
statische Dateien deployen. Das Einzige, das man richtig hinbekommen muss, ist
die Kombination aus `BaseURL` und `RelativeURLs`: `BaseURL` speist Canonical-Tags
und Open Graph (die absolut sein müssen), während `RelativeURLs: true` jeden
In-Page-Link und jedes Asset relativ zu seiner Seite macht — **ein Build
funktioniert also von der Festplatte geöffnet, an einer Localhost-Wurzel und
unter jedem Basispfad**.

```go
root := ssg.Node{ /* … */ }

if err := ssg.Build(root, ssg.BuildOptions{
	OutDir:       "build",
	BaseURL:      base,          // variiert pro Host — siehe unten
	RelativeURLs: true,          // gleich für jeden Host
}); err != nil {
	log.Fatal(err)
}
```

Mit `RelativeURLs: true` kann die Seite nicht für den falschen Host gebaut
werden — es gibt keine pro-Host-URL, die man falsch machen könnte. Nur die
Canonical-Basis ändert sich.

## GitHub Pages

Project Pages werden von `https://<user>.github.io/<repo>/` ausgeliefert, die
Basis ist also `/<repo>` — für diese Seite `/astwerk-website`. Der
Deployment-Workflow hat dieselbe Form wie das eigene
[pages.yml](https://github.com/LukasDerBaum42/astwerk-website/blob/main/.github/workflows/pages.yml)
dieser Seite:

```yaml
- uses: actions/configure-pages@v5
- name: Generate templates
  run: go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate
- name: Build
  run: go run . -base "/${{ github.event.repository.name }}"
- uses: actions/upload-pages-artifact@v4
  with: { path: build }
```

Drei Details, die beißen:

- `-base` muss zu dem Pfad passen, unter dem Pages das Repo ausliefert. Bei
  einer eigenen Domain weglassen (`-base ""`) — alles funktioniert trotzdem,
  weil die Links relativ sind.
- GitHub Pages führt Jekyll aus, wenn man es nicht abbestellt, und Jekyll
  verwirft Dateien und Verzeichnisse, die mit einem Unterstrich beginnen.
  Schreibe eine `.nojekyll`-Datei in die Ausgabe
  (`Files: map[string]ssg.PageFunc{".nojekyll": …}`).
- `templ generate` ist ein Build-Schritt, kein eingechecktes Artefakt — in CI
  ausführen.

## Netlify

Netlify liefert von einer Domain-Wurzel aus, also ist `BaseURL` `""` (oder
deine eigene Domain) und `RelativeURLs: true` bleibt:

```toml
# netlify.toml
[build]
  command = "templ generate && go run . -base \"\""
  publish = "build"
```

## Cloudflare Pages

Gleiche Form — eine Domain-Wurzel, `-base ""`, das `build`-Verzeichnis
publizieren. Beide Dienste liefern die Ausgabe unverändert aus, mehr ist also
nicht zu konfigurieren.

## Ein Build, überall

Weil die Ausgabe relativ ist, kann dasselbe `build/`-Verzeichnis direkt von der
Festplatte geöffnet, unter `localhost:8000` ausgeliefert und auf jeden der
Hosts oben gepusht werden. Halte `BaseURL` als Build-Flag statt als Konstante,
und Deployment bleibt eine Einzeiler-Entscheidung pro Host.
