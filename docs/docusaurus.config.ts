import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'lazyoci',
  tagline: 'Terminal UI for exploring OCI container registries',
  favicon: 'img/favicon.ico',

  url: 'https://lazyoci.dev',
  baseUrl: '/',

  organizationName: 'mistergrinvalds',
  projectName: 'lazyoci',

  onBrokenLinks: 'throw',

  markdown: {
    hooks: {
      onBrokenMarkdownLinks: 'warn',
    },
  },

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  plugins: [
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'tutorials',
        path: 'content/tutorials',
        routeBasePath: 'tutorials',
        sidebarPath: './sidebars/tutorials.js',
        editUrl: 'https://github.com/mistergrinvalds/lazyoci/tree/main/docs/',
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'guides',
        path: 'content/guides',
        routeBasePath: 'guides',
        sidebarPath: './sidebars/guides.js',
        editUrl: 'https://github.com/mistergrinvalds/lazyoci/tree/main/docs/',
      },
    ],
    [
      '@docusaurus/plugin-content-docs',
      {
        id: 'explanation',
        path: 'content/explanation',
        routeBasePath: 'explanation',
        sidebarPath: './sidebars/explanation.js',
        editUrl: 'https://github.com/mistergrinvalds/lazyoci/tree/main/docs/',
      },
    ],
  ],

  presets: [
    [
      'classic',
      {
        docs: {
          path: 'content/reference',
          routeBasePath: 'reference',
          sidebarPath: './sidebars/reference.js',
          editUrl: 'https://github.com/mistergrinvalds/lazyoci/tree/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    navbar: {
      title: 'lazyoci',
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'tutorialsSidebar',
          docsPluginId: 'tutorials',
          position: 'left',
          label: 'Tutorials',
        },
        {
          type: 'docSidebar',
          sidebarId: 'guidesSidebar',
          docsPluginId: 'guides',
          position: 'left',
          label: 'How-to Guides',
        },
        {
          type: 'docSidebar',
          sidebarId: 'referenceSidebar',
          position: 'left',
          label: 'Reference',
        },
        {
          type: 'docSidebar',
          sidebarId: 'explanationSidebar',
          docsPluginId: 'explanation',
          position: 'left',
          label: 'Explanation',
        },
        {
          href: 'https://github.com/mistergrinvalds/lazyoci',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Learn',
          items: [
            {label: 'Getting Started', to: '/tutorials/getting-started'},
            {label: 'Authentication', to: '/guides/authentication'},
          ],
        },
        {
          title: 'Reference',
          items: [
            {label: 'CLI', to: '/reference/cli/'},
            {label: 'TUI Keybindings', to: '/reference/tui-keybindings'},
            {label: 'Configuration', to: '/reference/configuration'},
          ],
        },
        {
          title: 'Community',
          items: [
            {label: 'GitHub', href: 'https://github.com/mistergrinvalds/lazyoci'},
            {label: 'Issues', href: 'https://github.com/mistergrinvalds/lazyoci/issues'},
          ],
        },
      ],
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} lazyoci contributors.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'yaml', 'json', 'go'],
    },
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
