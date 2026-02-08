/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
module.exports = {
  tutorialsSidebar: [
    'index',
    {
      type: 'category',
      label: 'Getting Started',
      collapsed: false,
      link: {type: 'doc', id: 'getting-started/index'},
      items: [
        'getting-started/installation',
        'getting-started/first-browse',
        'getting-started/first-pull',
      ],
    },
    {
      type: 'category',
      label: 'Private Registries',
      link: {type: 'doc', id: 'private-registries/index'},
      items: [
        'private-registries/adding-a-registry',
        'private-registries/docker-credentials',
        'private-registries/browsing-private-artifacts',
      ],
    },
    {
      type: 'category',
      label: 'Local Development',
      link: {type: 'doc', id: 'local-dev/index'},
      items: [
        'local-dev/local-registry',
        'local-dev/pushing-artifacts',
        'local-dev/artifact-types',
      ],
    },
  ],
};
