const previewByPath = new Map<string, string>();

export function rememberPreview(path: string, previewUrl: string) {
  if (!path || !previewUrl) return;
  previewByPath.set(path.replace(/\\/g, '/'), previewUrl);
}

export function takePreview(path: string): string | undefined {
  const key = path.replace(/\\/g, '/');
  const url = previewByPath.get(key);
  if (url) previewByPath.delete(key);
  return url;
}

export function peekPreview(path: string): string | undefined {
  return previewByPath.get(path.replace(/\\/g, '/'));
}

export function isImageName(name: string): boolean {
  return /\.(png|jpe?g|gif|webp|bmp|svg)$/i.test(name);
}

/** Build a preview URL for a workspace-relative upload/path. */
export function workspacePreviewUrl(baseUrl: string, path: string): string {
  const rel = path.replace(/\\/g, '/').replace(/^\.\//, '');
  return `${baseUrl}/api/raw?path=${encodeURIComponent(rel)}`;
}
