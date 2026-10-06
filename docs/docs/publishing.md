---
title: Publishing docs
description: Run Docusaurus locally and deploy this site to GitHub Pages.
---

# Publishing docs

User-facing Markdown lives in the repository **`docs/docs`** folder.
The root of the **`docs/`** folder is a Docusaurus 3 app that points
`docs.path` at `./docs` (subfolder) so the Go module stays free of
Node tooling at the root.

## Local development

Node.js 18+.

```bash
cd docs
npm install
npm start
```

Open `http://localhost:3000/scaffold-go/` — `baseUrl` is `/scaffold-go/` to
match GitHub Pages project sites
(`https://bancolombia.github.io/scaffold-go/`).

```bash
npm run build
npm run serve
```

`docs/build/` is the static output (gitignored).

## Configuration

| File | Role |
|------|------|
| [website/docusaurus.config.js](https://github.com/bancolombia/scaffold-go/blob/master/website/docusaurus.config.js) | `url`, `baseUrl`, Mermaid, navbar |
| [website/sidebars.js](https://github.com/bancolombia/scaffold-go/blob/master/website/sidebars.js) | Sidebar order |
| [docs/intro.md](https://github.com/bancolombia/scaffold-go/blob/master/docs/intro.md) | Home (`slug: /`) |
| [.github/workflows/docs.yml](https://github.com/bancolombia/scaffold-go/blob/master/.github/workflows/docs.yml) | CI deploy |

If the GitHub Pages URL is a **user/org** site instead of a project site,
set `baseUrl: '/'` and adjust `url`.

## GitHub Pages (one-time)

1. Merge the Docusaurus files to `master`.
2. Repo **Settings → Pages**: source **Deploy from a branch**, branch
   **`gh-pages`**, folder `/` (root).
3. Ensure **Actions** can write to the repo (`contents: write` on the
   workflow).
4. After the `docs` workflow succeeds, wait a minute and open
   `https://bancolombia.github.io/scaffold-go/`.

The workflow builds on pushes to `master` that touch `docs/**`,
or the workflow file, then publishes `docs/build` to `gh-pages`.

:::note Custom domain
Set `url` to the domain and `baseUrl` to `'/'`, then add a `CNAME` in
`website/static/`.
:::

## Editing content

- Add a page under `docs/` with Docusaurus frontmatter (`title`,
  `description`).
- Register it in `website/sidebars.js`.
- Use relative links (`./entities.md`). Mermaid is enabled
  (`@docusaurus/theme-mermaid`).
- Admonitions: `:::tip`, `:::warning`, `:::note`, `:::info`.

## Related pages

| Page | Relationship |
|------|-------------|
| [Introduction](./intro.md) | Site home |
| [Contributing](./contributing.md) | Docs PRs |
