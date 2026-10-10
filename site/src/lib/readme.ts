// Snapshots contain sanitized standalone README HTML. Nest headings beneath the page section.
export function nestedReadmeHTML(html: string): string {
  return html.replace(/<(\/?)h([1-6])(\s|>)/gi, (_, end, level, boundary) => `<${end}h${Math.min(6, Number(level) + 2)}${boundary}`);
}
