import type { ComponentCall } from './common.mts';

// Helpers to suggest Go package, function, and file names based on call names
export const toTokens = (s: string) => s.split(/[^a-zA-Z0-9]+/).filter(Boolean);
export const toGoExported = (tokens: string[]) =>
  tokens.map((t) => t.charAt(0).toUpperCase() + t.slice(1)).join('');
export const toSnake = (tokens: string[]) =>
  tokens.map((t) => t.toLowerCase()).join('_');
export const guessPkg = (call: ComponentCall) => {
  if (call.directory) return call.directory;
  const n = call.name;
  if (n.startsWith('cli.')) return 'cmd/thoth';
  if (n.startsWith('fs.')) return 'internal/fs';
  if (n.startsWith('diagnose.config')) return 'internal/config';
  if (n.startsWith('meta.parse')) return 'internal/meta';
  if (n.startsWith('meta.load')) return 'internal/meta';
  if (n.startsWith('meta.save')) return 'internal/save';
  if (n.startsWith('meta.update')) return 'internal/save';
  if (n.startsWith('meta.diff')) return 'internal/diff';
  if (n.startsWith('output.')) return 'internal/output';
  if (n.startsWith('shell.')) return 'internal/shell';
  if (n.startsWith('files.')) return 'internal/pipeline';
  if (n.startsWith('flow.')) return 'internal/pipeline';
  if (n.startsWith('action.')) return 'internal/config';
  if (n.startsWith('meta.')) return 'internal/pipeline';
  return 'internal';
};
export const suggestFor = (call: ComponentCall) => {
  const pkg = guessPkg(call);
  const tokens = toTokens(call.name);
  const func = toGoExported(tokens);
  const basename = (() => {
    switch (call.name) {
      case 'meta.filter.step':
        return 'meta_filter_step';
      case 'files.filter.step':
        return 'files_filter_step';
      case 'meta.map.step':
        return 'meta_map_step';
      case 'files.map.step':
        return 'files_map_step';
      case 'meta.reduce.step':
        return 'meta_reduce_step';
      case 'meta.map.post-shell':
        return 'meta_post_shell';
      case 'files.map.post':
        return 'files_map_post';
      case 'files.map.post.update':
        return 'files_post_update';
      default:
        return toSnake(tokens.slice(-2)) || toSnake(tokens);
    }
  })();
  const file = `${pkg}/${basename}.go`;
  return { pkg, func, file };
};
