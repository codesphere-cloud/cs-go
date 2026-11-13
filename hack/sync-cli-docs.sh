#!/usr/bin/env bash
# Copyright (c) Codesphere Inc.
# SPDX-License-Identifier: Apache-2.0

# Sync the Docusaurus output of hack/gendocs into the docs repository.
set -euo pipefail
shopt -s nullglob

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 <docusaurus-docs> <docs-repo>" >&2
  exit 1
fi
source_dir=$1
reference=$2/docs/APIs_and_Automations
target=$reference/cs-cli

if [[ ! -f "$source_dir/cs.md" || ! -f "$source_dir/_category_.json" ]]; then
  echo "No Docusaurus CLI reference found; run make docs-docusaurus first" >&2
  exit 1
fi
if [[ ! -d "$reference" ]]; then
  echo "Missing APIs and Automations directory: $reference" >&2
  exit 1
fi

# Compare the whole reference without its generated version notes. Retain the
# published version unless command content, pages, or sidebar metadata changed.
reference_changed=false
for page in "$source_dir"/cs*.md "$source_dir/_category_.json"; do
  name=${page##*/}
  if [[ ! -f "$target/$name" ]] || ! cmp -s \
    <(sed '/^Generated from CS CLI version `[^`]*`\.$/d' "$page") \
    <(sed '/^Generated from CS CLI version `[^`]*`\.$/d' "$target/$name"); then
    reference_changed=true
    break
  fi
done
for stale in "$target"/cs*.md; do
  if [[ ! -f "$source_dir/${stale##*/}" ]]; then
    reference_changed=true
  fi
done

if "$reference_changed"; then
  mkdir -p "$target"
  # Remove obsolete commands only from the generator-owned section.
  for stale in "$target"/cs*.md; do
    if [[ ! -f "$source_dir/${stale##*/}" ]]; then
      rm -- "$stale"
    fi
  done
  cp -- "$source_dir"/cs*.md "$source_dir/_category_.json" "$target/"
else
  echo 'CLI reference unchanged apart from version; keeping the published version.'
fi

# Keep the hand-written workflow guide distinct from the generated reference.
guide=$reference/cli-command-reference.mdx
if [[ -f "$guide" ]]; then
  sed -i 's/^title: "CLI: Commands \& Workflows"$/title: "CLI Workflow Examples"/; s/^# CLI: Commands \& Workflows$/# CLI Workflow Examples/' "$guide"
  if ! grep -qF '[CS CLI Reference](./cs-cli/cs.md)' "$guide"; then
    sed -i '/^This guide breaks down the available commands/a\
\nFor full command and flag documentation, see [CS CLI Reference](./cs-cli/cs.md).' "$guide"
  fi
fi
