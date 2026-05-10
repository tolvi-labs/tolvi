#!/usr/bin/env node
// Usage:
//   validate-schema.js
//
// Validates the entire sample vault against the JSON Schemas in spec/schemas/.
// Walks examples/sample-vault/ and runs each file through the appropriate schema:
//   - .vault-meta.json → spec/schemas/vault-meta.json
//   - decisions/*.md  → spec/schemas/decision.json (frontmatter only)
//   - sessions/*.md   → spec/schemas/session.json  (frontmatter only)
//   - patterns/*.md   → spec/schemas/pattern.json  (frontmatter only)
//
// Exits 0 if every file validates; exits 1 with details if any fails.

const fs = require('fs');
const path = require('path');
const { execFileSync } = require('child_process');
const Ajv = require('ajv/dist/2020');

const repoRoot = path.resolve(__dirname, '..', '..');
const schemasDir = path.join(repoRoot, 'spec', 'schemas');
const sampleVault = path.join(repoRoot, 'examples', 'sample-vault');
const extractor = path.join(__dirname, 'extract-frontmatter.js');

function loadSchema(name) {
  return JSON.parse(fs.readFileSync(path.join(schemasDir, name), 'utf8'));
}

function extractFrontmatter(mdPath) {
  const out = execFileSync('node', [extractor, mdPath], { encoding: 'utf8' });
  return JSON.parse(out);
}

const ajv = new Ajv({ strict: true, allErrors: true });
const validators = {
  vaultMeta: ajv.compile(loadSchema('vault-meta.json')),
  decision: ajv.compile(loadSchema('decision.json')),
  session: ajv.compile(loadSchema('session.json')),
  pattern: ajv.compile(loadSchema('pattern.json')),
};

let failures = 0;

function check(label, validator, data, source) {
  if (validator(data)) {
    console.log(`OK: ${label} ${source}`);
    return;
  }
  failures += 1;
  console.error(`FAIL: ${label} ${source}`);
  for (const err of validator.errors) {
    console.error(`  ${err.instancePath || '/'} ${err.message} (${JSON.stringify(err.params)})`);
  }
}

// 1. .vault-meta.json
const metaPath = path.join(sampleVault, '.vault-meta.json');
check('vault-meta', validators.vaultMeta, JSON.parse(fs.readFileSync(metaPath, 'utf8')), metaPath);

// 2-4. decisions, sessions, patterns
const types = [
  { dir: 'decisions', validator: validators.decision, label: 'decision' },
  { dir: 'sessions',  validator: validators.session,  label: 'session'  },
  { dir: 'patterns',  validator: validators.pattern,  label: 'pattern'  },
];

for (const { dir, validator, label } of types) {
  const dirPath = path.join(sampleVault, dir);
  const files = fs.readdirSync(dirPath).filter((f) => f.endsWith('.md')).sort();
  for (const f of files) {
    const filePath = path.join(dirPath, f);
    const fm = extractFrontmatter(filePath);
    check(label, validator, fm, filePath);
  }
}

if (failures > 0) {
  console.error(`\n${failures} file(s) failed schema validation.`);
  process.exit(1);
}

console.log(`\nAll vault files conform to tolvi-format-v1.`);
