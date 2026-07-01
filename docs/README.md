# Velo Deploy — Documentation

This directory contains the source for the Velo Deploy documentation site, built with [Astro](https://astro.build/) 5 + [Starlight](https://starlight.astro.build/).

The site is deployed automatically to **https://antojsh.github.io/velo-deploy/** via GitHub Actions (see [`.github/workflows/docs.yml`](../.github/workflows/docs.yml)).

## Quick start

Requirements: Node.js 20+ and pnpm 9+.

```bash
pnpm install
pnpm dev
```

The dev server runs on **http://localhost:4321/velo-deploy/** with hot-reload.

## Scripts

| Script | Description |
| --- | --- |
| `pnpm dev` | Start the dev server with hot-reload. |
| `pnpm build` | Build the static site to `dist/`. |
| `pnpm preview` | Serve the production build locally. |
| `pnpm check` | Type-check and validate internal links. |
| `pnpm astro -- --help` | Run the underlying Astro CLI. |

## Project structure

```
docs/
├── astro.config.mjs            # Starlight + i18n config
├── package.json
├── tsconfig.json
├── src/
│   ├── assets/                 # Logo SVGs
│   ├── content.config.ts       # Content collection schema
│   ├── env.d.ts
│   └── content/
│       └── docs/
│           ├── en/             # English (default locale)
│           │   ├── index.md
│           │   ├── getting-started/
│           │   ├── guide/
│           │   ├── architecture/
│           │   ├── reference/
│           │   └── operations/
│           └── es/             # Español
│               └── … same structure …
└── public/                     # Static assets (favicon, og image)
```

## Authoring

- Every page is a Markdown file with frontmatter. See `src/content/docs/en/index.md` for the hero/landing page format.
- Each English page should have a matching Spanish page. The Spanish file lives at the same relative path under `es/`.
- Starlight autogenerates the sidebar from the directory structure, so you usually do not need to touch `astro.config.mjs` to add a new page.
- For components like `<Card>`, `<Tabs>`, `<Aside>`, etc., import from `@astrojs/starlight/components` at the top of the file.

## Adding a new page

1. Create `src/content/docs/en/<path>.md`.
2. Create `src/content/docs/es/<path>.md` with the translation.
3. The sidebar updates automatically.
4. Run `pnpm check` to validate links and frontmatter.

## Adding a new section

1. Create a new directory under `src/content/docs/en/`.
2. Mirror it under `src/content/docs/es/`.
3. Add the section to the sidebar in `astro.config.mjs` if you want a custom label.

## Deployment

The site is built and deployed to GitHub Pages on every push to `master` that touches files under `docs/`. See the workflow in `.github/workflows/docs.yml`.

If you need to deploy manually:

```bash
pnpm build
# Then upload dist/ to your static host of choice
```

The build output is fully static — no Node.js server is required to serve it.
