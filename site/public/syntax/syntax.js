(() => {
  'use strict';
  const highlighter = globalThis.hljs;
  if (!highlighter) return;
  highlighter.registerLanguage('cli', h => ({
    name: 'CLI commands',
    contains: [
      h.HASH_COMMENT_MODE, h.QUOTE_STRING_MODE, h.APOS_STRING_MODE,
      { scope: 'title.function', begin: /^[ \t]*(?:\$\s*)?[A-Za-z_][\w./-]*/m },
      { scope: 'attr', begin: /--?[\w][\w-]*(?==|\s|$)/ },
      { scope: 'string', begin: /(?:@[\w.-]+\/|https?:\/\/|github\.com\/)[^\s]+/ },
      { scope: 'variable', begin: /\$\{?[A-Za-z_][\w]*\}?/ },
      { scope: 'number', begin: /\b\d+(?:\.\d+)+\b/ }
    ]
  }));
  function surface(block) {
    if (typeof getComputedStyle !== 'function') return;
    for (let node = block; node; node = node.parentElement) {
      const colour = getComputedStyle(node).backgroundColor.match(/[\d.]+/g);
      if (!colour || colour.length < 3 || (colour.length > 3 && Number(colour[3]) < 0.5)) continue;
      const rgb = colour.slice(0, 3).map(value => Number(value) / 255)
        .map(value => value <= 0.04045 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4);
      const light = 0.2126 * rgb[0] + 0.7152 * rgb[1] + 0.0722 * rgb[2];
      block.dataset.syntaxSurface = light > 0.45 ? 'light' : 'dark';
      return;
    }
  }
  // Language labels from Markdown and product profiles are authoritative.
  // Reports and architecture diagrams are plain text, never guessed as code.
  const blocks = document.querySelectorAll('pre > code, .command > code, [data-syntax-language]');
  for (const block of blocks) {
    if (block.closest('.astro-code, .shiki') || block.dataset.syntaxHighlighted ||
        block.classList.contains('hljs') || block.classList.contains('nohighlight')) continue;
    const label = [...block.classList].find(value => /^(?:language|lang)-/.test(value));
    const language = block.dataset.syntaxLanguage || label?.replace(/^(?:language|lang)-/, '');
    if (!language || language === 'plaintext' || language === 'text' || !highlighter.getLanguage(language)) continue;
    const original = block.textContent;
    try {
      const result = highlighter.highlight(original, { language, ignoreIllegals: true });
      block.innerHTML = result.value;
      if (block.textContent !== original) {
        block.textContent = original;
        continue;
      }
      block.classList.add('hljs');
      surface(block);
      block.dataset.syntaxHighlighted = result.language;
    } catch {
      block.textContent = original;
    }
  }
})();
