import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: 'doc',
      id: 'intro',
      label: 'Gira',
    },
    {
      type: 'category',
      label: 'Concepts',
      items: [
        'concepts/target-system',
        'concepts/component',
        'concepts/service',
        'concepts/library',
        'concepts/infrastructure',
        'concepts/contract',
        'concepts/evidence-record',
        'concepts/macro-map',
        'concepts/extractor',
        'concepts/reference-corpus',
        'concepts/perturbed-variant',
        'concepts/change-maker',
        'concepts/impact-briefing',
        'concepts/impact-set',
      ],
    },
    {
      type: 'category',
      label: 'CLI Reference',
      items: [
        'cli/index',
        'cli/impact',
        'cli/slice',
        'cli/serve',
        'cli/export',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/quickstart',
        'guides/change-maker-workflow',
        'guides/evidence-inspection',
        'guides/export',
      ],
    },
  ],
};

export default sidebars;
