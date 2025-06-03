#!/bin/bash

make generate-licenses
make docs
if [[ `git status --porcelain` ]]; then
  git add .
  git commit -m "docs(licenses,usage): Update docs & licenses"
  git push origin main
fi

