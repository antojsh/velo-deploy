import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

export default defineConfig({
  site: 'https://antojsh.github.io',
  base: '/velo-deploy',
  integrations: [
    starlight({
      title: 'Velo Deploy',
      description:
        'Bare Metal PaaS — Deploy Node.js applications and static sites to any VPS without Docker.',
      logo: {
        src: './src/assets/logo.svg',
        replacesTitle: false,
      },
      favicon: '/favicon.svg',
      social: [
        { icon: 'github', label: 'GitHub', href: 'https://github.com/antojsh/velo-deploy' },
      ],
      editLink: {
        baseUrl: 'https://github.com/antojsh/velo-deploy/edit/master/docs/',
      },
      sidebar: [
        {
          label: 'Getting Started',
          autogenerate: { directory: 'getting-started' },
        },
        {
          label: 'Guides',
          autogenerate: { directory: 'guide' },
        },
        {
          label: 'Architecture',
          autogenerate: { directory: 'architecture' },
        },
        {
          label: 'Reference',
          autogenerate: { directory: 'reference' },
        },
        {
          label: 'Operations',
          autogenerate: { directory: 'operations' },
        },
      ],
      defaultLocale: 'root',
      locales: {
        root: { label: 'English', lang: 'en-US' },
        es: { label: 'Español', lang: 'es-ES' },
      },
      lastUpdated: true,
      expressiveCode: true,
    }),
  ],
});
