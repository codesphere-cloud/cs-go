#!/bin/bash
# Copyright (c) Codesphere Inc.
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

make generate-license
make docs
if [[ $(git status --porcelain) ]]; then
  git add .
  git config --global user.name "${GITHUB_ACTOR}"
  git config --global user.email "${GITHUB_ACTOR}@users.noreply.github.com"
  git commit -m "docs(licenses,usage): Update docs & licenses"
  git push origin main
fi

