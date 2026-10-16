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
        'cli/map',
        'cli/impact',
        'cli/slice',
        'cli/index',
        'cli/serve',
        'cli/explain',
        'cli/tour',
        'cli/diff',
        'cli/flows',
        'cli/locate',
        'cli/skills',
      ],
    },
    {
      type: 'category',
      label: 'Guides',
      items: [
        'guides/quickstart',
        'guides/change-maker-workflow',
        'guides/evidence-inspection',
      ],
    },
    {
      type: 'doc',
      id: 'cli/export',
      label: 'Export',
    },
  ],
};

export default sidebars;
