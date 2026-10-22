// @ts-check
// Docusaurus site for scaffold-go. Markdown lives in ../docs so the
// Go repository keeps a conventional docs/ tree while this folder is the
// Node.js wrapper that GitHub Pages will build.

import {themes as prismThemes} from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'scaffold-go',
  tagline: 'Clean-architecture scaffolding for Go microservices',
  favicon: 'img/favicon.svg',

  url: 'https://bancolombia.github.io',
  baseUrl: '/scaffold-go/',

  organizationName: 'bancolombia',
  projectName: 'scaffold-go',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  markdown: {
    mermaid: true,
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },
  themes: ['@docusaurus/theme-mermaid'],

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          path: './docs',
          routeBasePath: '/',
          sidebarPath: './sidebars.js',
          editUrl: ({docPath}) =>
            `https://github.com/bancolombia/scaffold-go/edit/master/docs/${docPath}`,
          showLastUpdateTime: false,
        },
        blog: false,
        pages: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      image: 'img/social-card.svg',
      colorMode: {
        defaultMode: 'dark',
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'scaffold-go',
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'docs',
            position: 'left',
            label: 'Docs',
          },
          {
            href: 'https://github.com/bancolombia/scaffold-go',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {label: 'Introduction', to: '/'},
              {label: 'First project', to: '/getting-started/first-project'},
              {label: 'CLI reference', to: '/reference/cli'},
            ],
          },
          {
            title: 'Guides',
            items: [
              {label: 'Entities', to: '/guides/entities'},
              {label: 'Use cases', to: '/guides/use-cases'},
              {label: 'HTTP entry-points', to: '/guides/http-entry-points'},
            ],
          },
          {
            title: 'Project',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/bancolombia/scaffold-go',
              },
              {label: 'Contributing', to: '/contributing'},
              {label: 'Publishing docs', to: '/publishing'},
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Bancolombia. Built with Docusaurus.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
        additionalLanguages: ['bash', 'go', 'yaml', 'json'],
      },
      mermaid: {
        theme: {light: 'neutral', dark: 'dark'},
      },
    }),
};

export default config;
