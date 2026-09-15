#!/usr/bin/env bash
# Fails (exit 1) if the per-agent plugin manifests disagree.
#
# This repo ships the same plugin to several ecosystems, each wanting its own
# manifest: .claude-plugin, .codex-plugin, .cursor-plugin, .devin-plugin and
# .hermes-plugin. That is five copies of a name and a version, and copies of
# one fact drift. This repo has already paid for that twice, with the schemas
# served at the $id URL and with the CLI and skill deriving routing separately.
#
# Every path a manifest declares is checked too: a manifest pointing at a file
# that does not exist is worse than one that stays quiet about the capability.
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../.."

node -e '
const fs = require("fs"), path = require("path");
const dirs = [".claude-plugin", ".codex-plugin", ".cursor-plugin", ".devin-plugin"];
let fail = 0;
const seen = {};

for (const d of dirs) {
  const f = path.join(d, "plugin.json");
  if (!fs.existsSync(f)) continue;
  const m = JSON.parse(fs.readFileSync(f, "utf8"));
  seen[d] = { name: m.name, version: m.version };

  for (const key of ["skills", "hooks", "logo"]) {
    const v = m[key] ?? (m.interface || {})[key];
    if (typeof v === "string" && v.startsWith("./") && !fs.existsSync(v.slice(2))) {
      console.log(`✗ ${f}: ${key} points at ${v}, which does not exist`);
      fail = 1;
    }
  }
}

const yaml = ".hermes-plugin/plugin.yaml";
if (fs.existsSync(yaml)) {
  const lines = fs.readFileSync(yaml, "utf8").split("\n");
  const g = (k) => {
    const hit = lines.find((l) => l.startsWith(k + ":"));
    return hit ? hit.slice(k.length + 1).trim() : undefined;
  };
  seen[".hermes-plugin"] = { name: g("name"), version: g("version") };
}

const gem = "gemini-extension.json";
if (fs.existsSync(gem)) {
  const m = JSON.parse(fs.readFileSync(gem, "utf8"));
  seen[gem] = { name: m.name, version: m.version };
  if (m.contextFileName && !fs.existsSync(m.contextFileName)) {
    console.log(`✗ ${gem}: contextFileName ${m.contextFileName} does not exist`);
    fail = 1;
  }
}

const entries = Object.entries(seen);
if (entries.length === 0) { console.log("no plugin manifests here; nothing to check"); process.exit(0); }
const [refKey, ref] = entries[0];
for (const [k, v] of entries.slice(1)) {
  if (v.name !== ref.name) { console.log(`✗ ${k}: name ${v.name} != ${refKey} name ${ref.name}`); fail = 1; }
  if (v.version !== ref.version) { console.log(`✗ ${k}: version ${v.version} != ${refKey} version ${ref.version}`); fail = 1; }
}
if (!fail) console.log(`✓ ${entries.length} plugin manifests agree on ${ref.name} ${ref.version}, and every declared path exists`);
process.exit(fail);
'
