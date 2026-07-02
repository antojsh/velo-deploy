import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// In dev the user expects http://localhost:4321/ to work.
// In production (GitHub Pages) the site lives under /velo-deploy.
const base = process.env.NODE_ENV === 'production' ? '/velo-deploy' : '/';

export default defineConfig({
  site: 'https://antojsh.github.io',
  base,
  integrations: [
    starlight({
      title: 'Velo Deploy',
      description:
        'Bare Metal PaaS — Deploy Node.js applications and static sites to any VPS without Docker.',
      logo: {
        src: './src/assets/logo-mark.svg',
        replacesTitle: false,
      },
      favicon: '/favicon.svg',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/antojsh/velo-deploy',
        },
      ],
      editLink: {
        baseUrl: 'https://github.com/antojsh/velo-deploy/edit/master/docs/',
      },
      sidebar: [
        {
          label: 'Getting Started',
          items: [{ autogenerate: { directory: 'getting-started' } }],
        },
        {
          label: 'Guides',
          items: [{ autogenerate: { directory: 'guide' } }],
        },
        {
          label: 'Architecture',
          items: [{ autogenerate: { directory: 'architecture' } }],
        },
        {
          label: 'Reference',
          items: [{ autogenerate: { directory: 'reference' } }],
        },
        {
          label: 'Operations',
          items: [{ autogenerate: { directory: 'operations' } }],
        },
      ],
      defaultLocale: 'root',
      locales: {
        root: { label: 'English', lang: 'en-US' },
        es: { label: 'Español', lang: 'es-ES' },
      },
      lastUpdated: true,
      expressiveCode: true,
      customCss: ['./src/styles/brand.css'],
    }),
  ],
});
