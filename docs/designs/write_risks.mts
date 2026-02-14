import { writeFile } from 'node:fs/promises';
import { writeSectionStickie } from './common.mts';
import { risks } from './risks.mts';

const RISKS_PATH = 'docs/RISKS.md';

export const generateRisksReport = async () => {
  const entries = Object.values(risks).sort((a, b) =>
    a.name.localeCompare(b.name),
  );

  const lines: string[] = [];
  lines.push('# Risks Overview (Generated)');
  lines.push('');
  lines.push('This document summarizes key risks and mitigations.');
  lines.push('');

  // Summary table
  lines.push('## Summary');
  for (const r of entries) {
    lines.push(`- ${r.title} [${r.name}]`);
  }
  lines.push('');

  // Detailed sections
  for (const r of entries) {
    lines.push(`## ${r.title} [${r.name}]`);
    lines.push('');
    lines.push(`- Description: ${r.description}`);
    lines.push(`- Mitigation: ${r.mitigation}`);
    lines.push('');
  }

  await writeFile(RISKS_PATH, lines.join('\n'), 'utf8');

  // Also generate one stickie per risk with markdown-like sections in the note.
  for (const r of entries) {
    const title = `Risk: ${r.title}`;
    const note = [
      `Description:\n- ${r.description}`,
      `Mitigation:\n- ${r.mitigation}`,
    ].join('\n\n');
    // Place under notes/ root folder (no subfolder)
    await writeSectionStickie(title, note, 'notes');
  }
};
