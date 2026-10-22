# scaffold-go documentation site

This folder is the [Docusaurus](https://docusaurus.io/) wrapper for the Markdown
in [`../docs`](../docs). The Go CLI itself is unchanged; GitHub Pages builds
this site.

## Local preview

Requires Node.js 18 or newer.

```bash
cd website
npm install
npm start
```

The site is served at `http://localhost:3000/scaffold-go/` (`baseUrl` matches
GitHub Pages).

## Production build

```bash
cd website
npm install
npm run build
npm run serve
```

Output is written to `website/build/`. The GitHub Actions workflow in
`.github/workflows/docs.yml` runs the same build and publishes `website/build`
to the `gh-pages` branch.

## Content

Edit Markdown under `docs/` at the repository root, not under `website/`.
Sidebar order lives in [`sidebars.js`](sidebars.js). Site URL, `baseUrl`, and
theme settings live in [`docusaurus.config.js`](docusaurus.config.js).
