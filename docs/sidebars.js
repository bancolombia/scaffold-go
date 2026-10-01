/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  docs: [
    'intro',
    {
      type: 'category',
      label: 'Getting started',
      collapsed: false,
      items: [
        'getting-started/installation',
        'getting-started/first-project',
        'getting-started/naming',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      collapsed: false,
      items: [
        'guides/workflow',
        'guides/init',
        'guides/entities',
        'guides/use-cases',
        'guides/http-entry-points',
        'guides/consumers',
        'guides/database-adapters',
        'guides/api-clients',
        'guides/publishers',
        'guides/cache',
        'guides/configuration',
        'guides/dependency-injection',
      ],
    },
    {
      type: 'category',
      label: 'Architecture',
      items: [
        'architecture/overview',
        'architecture/generated-project',
        'architecture/cli-internals',
      ],
    },
    {
      type: 'category',
      label: 'Reference',
      items: [
        'reference/cli',
        'reference/templates',
        'reference/troubleshooting',
      ],
    },
    'contributing',
    'publishing',
  ],
};

export default sidebars;
