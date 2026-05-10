#!/usr/bin/env node
// Usage: extract-frontmatter.js <markdown-file>
//
// Reads a markdown file, extracts the YAML frontmatter between leading --- markers,
// converts it to JSON, and writes the JSON to stdout. Exits non-zero if no frontmatter
// is found.
//
// Handles the small YAML subset Tolvi vault frontmatter actually uses: scalar strings,
// integers, inline arrays of strings ([a, b]), and block arrays of strings (- a / - b).
// No nesting, no anchors, no multi-line strings — by design. We do not depend on
// js-yaml because adding a runtime YAML dep is more weight than this 60-line parser.

const fs = require('fs');

const file = process.argv[2];
if (!file) {
  console.error('Usage: extract-frontmatter.js <file.md>');
  process.exit(2);
}

const text = fs.readFileSync(file, 'utf8');
const match = text.match(/^---\n([\s\S]*?)\n---/);
if (!match) {
  console.error(`No frontmatter found in ${file}`);
  process.exit(1);
}

const obj = {};
const lines = match[1].split('\n');

let i = 0;
while (i < lines.length) {
  const line = lines[i];
  if (!line.trim() || line.trim().startsWith('#')) { i++; continue; }
  const kv = line.match(/^([a-zA-Z_][a-zA-Z0-9_-]*):\s*(.*)$/);
  if (!kv) { i++; continue; }
  const key = kv[1];
  const rawValue = kv[2].trim();

  if (rawValue === '') {
    // Block-style array on subsequent lines
    const arr = [];
    i++;
    while (i < lines.length && lines[i].match(/^\s*-\s+/)) {
      arr.push(lines[i].replace(/^\s*-\s+/, '').trim().replace(/^["']|["']$/g, ''));
      i++;
    }
    obj[key] = arr;
  } else if (rawValue.startsWith('[') && rawValue.endsWith(']')) {
    // Inline array
    const inner = rawValue.slice(1, -1).trim();
    obj[key] = inner === '' ? [] : inner.split(',').map((s) => s.trim().replace(/^["']|["']$/g, ''));
    i++;
  } else if (/^-?\d+$/.test(rawValue)) {
    obj[key] = parseInt(rawValue, 10);
    i++;
  } else {
    obj[key] = rawValue.replace(/^["']|["']$/g, '');
    i++;
  }
}

process.stdout.write(JSON.stringify(obj, null, 2) + '\n');
