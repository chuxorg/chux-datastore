#!/bin/bash

CURRENT_VERSION=$(git tag --sort=-version:refname | head -n 1)
MAJOR=$(echo $CURRENT_VERSION | cut -d. -f1)
MINOR=$(echo $CURRENT_VERSION | cut -d. -f2)
PATCH=$(echo $CURRENT_VERSION | cut -d. -f3)

NEW_MAJOR=${MAJOR_VALUE:-$MAJOR}
NEW_MINOR=${MINOR_VALUE:-$MINOR}
NEW_PATCH=0

if [ -z "$MAJOR_VALUE" ] && [ -z "$MINOR_VALUE" ]; then
  NEW_PATCH=$((PATCH + 1))
fi

NEW_VERSION="$NEW_MAJOR.$NEW_MINOR.$NEW_PATCH"

git add .
git commit -m "Release Version: $NEW_VERSION"
git tag $NEW_VERSION
git push && git push --tags