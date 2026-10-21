+++
Title = "Deployment"
Description = "GitHub Pages, Netlify, Cloudflare Pages — the BaseURL and RelativeURLs combination for each."
Order = 250
Group = "Deploy"
+++

astwerk builds a directory of static files, so it deploys anywhere static
files deploy. The one thing to get right is the `BaseURL` + `RelativeURLs`
combination: `BaseURL` feeds canonical tags and Open Graph (which must be
absolute), while `RelativeURLs: true` makes every in-page link and asset
relative to its page — so **one build works opened from disk, at a localhost
root, and under any base path**.

```go
root := ssg.Node{ /* … */ }

if err := ssg.Build(root, ssg.BuildOptions{
	OutDir:       "build",
	BaseURL:      base,          // varies per host — see below
	RelativeURLs: true,          // the same for every host
}); err != nil {
	log.Fatal(err)
}
```

With `RelativeURLs: true`, the site can't be built for the wrong host — there
is no per-host URL to get wrong. Only the canonical base changes.

## GitHub Pages

Project pages serve from `https://<user>.github.io/<repo>/`, so the base is
`/<repo>` — for this site, `/astwerk-website`. The deployment workflow is the
same shape as this site's own [pages.yml](https://github.com/LukasDerBaum42/astwerk-website/blob/main/.github/workflows/pages.yml):

```yaml
- uses: actions/configure-pages@v5
- name: Generate templates
  run: go run github.com/a-h/templ/cmd/templ@v0.3.1020 generate
- name: Build
  run: go run . -base "/${{ github.event.repository.name }}"
- uses: actions/upload-pages-artifact@v4
  with: { path: build }
```

Three details that bite:

- `-base` must match where Pages serves the repo from. For a custom domain,
  drop it (`-base ""`) — everything still works, because the links are
  relative.
- GitHub Pages runs Jekyll unless told not to, and Jekyll drops files and
  directories beginning with an underscore. Write a `.nojekyll` file into the
  output (`Files: map[string]ssg.PageFunc{".nojekyll": …}`).
- `templ generate` is a build step, not a checked-in artifact — run it in CI.

## Netlify

Netlify serves from a domain root, so `BaseURL` is `""` (or your custom
domain) and `RelativeURLs: true` stays:

```toml
# netlify.toml
[build]
  command = "templ generate && go run . -base \"\""
  publish = "build"
```

## Cloudflare Pages

Same shape — a domain root, `-base ""`, publish the `build` directory. Both
services serve the output as-is, so nothing else needs configuring.

## One build, everywhere

Because the output is relative, the same `build/` directory can be opened
straight off disk, served at `localhost:8000`, and pushed to any of the hosts
above. Keep `BaseURL` as a build flag rather than a constant, and deployment
stays a one-line decision per host.
