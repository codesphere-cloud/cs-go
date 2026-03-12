#!/bin/bash
set -euo pipefail

# This script splits the current PR branch (kubernetes_integration_tests) into
# two clean branches so the reviewer gets smaller, easier-to-review PRs:
#
#   PR 1  refactor/split-integration-tests   (based on main)
#         - Generate command path fixes + OpenAPI client cleanup
#         - Test file split: delete integration_test.go + create labeled files (same commit)
#         - Workflow matrix strategy + Makefile targets
#
#   PR 2  feat/kubernetes-export-tests       (based on PR 1 branch)
#         - New export_kubernetes_test.go + linters utility
#
# The original branch (kubernetes_integration_tests) is NOT modified or deleted.

ORIGINAL_BRANCH="kubernetes_integration_tests"
PR1_BRANCH="refactor/split-integration-tests"
PR2_BRANCH="feat/kubernetes-export-tests"

ROOT=$(git rev-parse --show-toplevel)
cd "$ROOT"

# --- Pre-flight checks -------------------------------------------------------

for branch in "$PR1_BRANCH" "$PR2_BRANCH"; do
    if git rev-parse --verify "$branch" >/dev/null 2>&1; then
        echo "ERROR: Branch '$branch' already exists. Delete it first or pick a different name."
        exit 1
    fi
done

STASHED=0
if ! git diff --quiet || ! git diff --cached --quiet; then
    echo "Stashing uncommitted changes on $ORIGINAL_BRANCH..."
    git stash push --include-untracked -m "pr-split-auto-stash"
    STASHED=1
fi

cleanup() {
    echo ""
    echo "Returning to original branch..."
    git checkout "$ORIGINAL_BRANCH" 2>/dev/null || true
    if [ "$STASHED" = "1" ]; then
        echo "Restoring stashed changes..."
        git stash pop 2>/dev/null || true
    fi
}
trap cleanup EXIT

###############################################################################
# PR 1: refactor/split-integration-tests
###############################################################################
echo ""
echo "============================================="
echo "  Creating PR 1: $PR1_BRANCH"
echo "============================================="

git checkout main
git checkout -b "$PR1_BRANCH"

# --- Commit 1: Generate command fixes + OpenAPI client cleanup ---------------
echo ""
echo "[PR1 · Commit 1/3] Generate command and OpenAPI client changes..."

git checkout "$ORIGINAL_BRANCH" -- \
    api/openapi_client/client.go \
    cli/cmd/generate_docker.go \
    cli/cmd/generate_docker_test.go \
    cli/cmd/generate_images.go \
    cli/cmd/generate_images_test.go \
    cli/cmd/generate_kubernetes.go \
    cli/cmd/generate_kubernetes_test.go \
    pkg/exporter/exporter.go \
    pkg/exporter/exporter_test.go

go mod tidy
git add -A
git commit -m "refactor: streamline file path handling in generate commands

- Use RepoRoot-relative paths instead of joining with current directory
- Introduce clientFactory pattern for lazy client creation in generate docker
- Format OpenAPI client code and fix float formatting"

echo "  ✓ Verifying compilation..."
go build ./...

# --- Commit 2: Split integration tests (delete + create in SAME commit) ------
echo ""
echo "[PR1 · Commit 2/3] Split integration tests into labeled files..."

# Delete the old monolithic test file
git rm int/integration_test.go

# Check out every split file + util changes from the original branch
git checkout "$ORIGINAL_BRANCH" -- \
    int/curl_test.go \
    int/error_handling_test.go \
    int/git_test.go \
    int/int_suite_test.go \
    int/list_test.go \
    int/log_test.go \
    int/monitor_test.go \
    int/pipeline_test.go \
    int/version_help_test.go \
    int/wakeup_test.go \
    int/workspace_test.go \
    int/util/constants.go \
    int/util/test_helpers.go \
    int/util/workspace.go

go mod tidy
git add -A
git commit -m "refactor: split integration tests into labeled files

Split the monolithic integration_test.go into individual files, each
with a Ginkgo label for selective execution via matrix CI:

  curl_test.go (curl), error_handling_test.go (error-handling),
  git_test.go (git), list_test.go (list), log_test.go (log),
  monitor_test.go (local), pipeline_test.go (pipeline),
  version_help_test.go (local), wakeup_test.go (wakeup),
  workspace_test.go (workspace)

Extract shared constants and helpers into int/util/."

echo "  ✓ Verifying compilation..."
go build ./...

# --- Commit 3: Workflow + Makefile -------------------------------------------
echo ""
echo "[PR1 · Commit 3/3] Workflow matrix strategy and Makefile targets..."

git checkout "$ORIGINAL_BRANCH" -- \
    .github/workflows/integration-test.yml \
    Makefile

git add -A
git commit -m "refactor: add matrix strategy to integration test workflow

- Split CI into build, test (parallel per-label matrix), and cleanup jobs
- Define label list once; derive matrix and unlabeled filter dynamically
- Add unique run-name with github.run_id to avoid parallel conflicts
- Fetch latest hadolint and kubeconform versions at install time
- Add per-label Makefile targets (test-int-workspace, test-int-list, etc.)"

###############################################################################
# PR 2: feat/kubernetes-export-tests
###############################################################################
echo ""
echo "============================================="
echo "  Creating PR 2: $PR2_BRANCH"
echo "============================================="

git checkout -b "$PR2_BRANCH"

echo ""
echo "[PR2 · Commit 1/1] Kubernetes export integration tests..."

git checkout "$ORIGINAL_BRANCH" -- \
    int/export_kubernetes_test.go \
    int/util/linters.go

go mod tidy
git add -A
git commit -m "feat: add integration tests for kubernetes export

Add comprehensive tests for generate docker, kubernetes, and images
commands. Tests validate generated Dockerfiles, shell scripts, and K8s
manifests (Deployments, Services, Ingress) by unmarshalling into typed
structs and running linters (hadolint, shellcheck, kubeconform) when
available.

Add int/util/linters.go with a single runLinter function that handles
tool discovery, skip-on-missing behavior, and failure reporting."

echo "  ✓ Verifying compilation..."
go build ./...

###############################################################################
# Summary
###############################################################################
echo ""
echo "============================================="
echo "  Done!"
echo "============================================="
echo ""
echo "Branches created:"
echo "  PR 1: $PR1_BRANCH  (3 commits on top of main)"
echo "  PR 2: $PR2_BRANCH  (1 commit on top of PR 1)"
echo ""
echo "Original branch '$ORIGINAL_BRANCH' is untouched."
echo ""
echo "Verify:"
echo "  git log --oneline main..$PR1_BRANCH"
echo "  git log --oneline $PR1_BRANCH..$PR2_BRANCH"
echo ""
echo "When ready:"
echo "  git push -u origin $PR1_BRANCH"
echo "  git push -u origin $PR2_BRANCH"
