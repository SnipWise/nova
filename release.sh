#!/bin/bash
: <<'COMMENT'
# Relase script for Snip SDK
Usage:
1. Update the version in release.env
2. Run this script: ./release.sh
3. Create a GitHub release with the new tag and description
COMMENT

set -o allexport; source release.env; set +o allexport

echo "Generating release: ${TAG} ${ABOUT}"

find . -name '.DS_Store' -type f -delete

git add .
git commit -m "📦 ${ABOUT}"
git push origin main

git tag -a ${TAG} -m "${ABOUT}"
git push origin ${TAG}

