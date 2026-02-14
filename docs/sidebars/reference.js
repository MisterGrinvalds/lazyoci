/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
module.exports = {
  referenceSidebar: [
    'index',
    {
      type: 'category',
      label: 'CLI',
      collapsed: false,
      link: {type: 'doc', id: 'cli/index'},
      items: [
        'cli/pull',
        'cli/build',
        'cli/browse',
        'cli/registry',
        'cli/config',
      ],
    },
    'lazy-config',
    'tui-keybindings',
    'configuration',
    'environment-variables',
    'artifact-types',
    'registry-compatibility',
    'themes',
  ],
};
