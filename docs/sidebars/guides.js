/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
module.exports = {
  guidesSidebar: [
    'index',
    {
      type: 'category',
      label: 'Authentication',
      collapsed: false,
      items: [
        'authentication',
        'cloud-registries',
        'insecure-registries',
      ],
    },
    {
      type: 'category',
      label: 'Working with Artifacts',
      collapsed: false,
      items: [
        'pulling-to-docker',
        'multi-platform-images',
        'custom-storage',
      ],
    },
    {
      type: 'category',
      label: 'Building & Pushing',
      collapsed: false,
      items: [
        'building-artifacts',
        'pushing-to-registries',
        'ci-cd-github-actions',
      ],
    },
    {
      type: 'category',
      label: 'Mirroring',
      collapsed: false,
      items: [
        'mirroring-charts',
      ],
    },
    {
      type: 'category',
      label: 'Troubleshooting',
      collapsed: false,
      items: [
        'troubleshooting-auth',
        'troubleshooting-connectivity',
        'troubleshooting-docker',
      ],
    },
  ],
};
