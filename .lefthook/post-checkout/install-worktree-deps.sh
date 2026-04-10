#!/usr/bin/env bash
set -euo pipefail

# In a linked worktree, --git-dir differs from --git-common-dir
if [ "$(git rev-parse --git-dir)" = "$(git rev-parse --git-common-dir)" ]; then
  exit 0
fi

task install:ui
task install:e2e
