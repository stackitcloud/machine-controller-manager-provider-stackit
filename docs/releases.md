# Release Procedure

## Overview

This document outlines the standard procedure for creating new releases of the STACKIT machine-controller-manager.

### 🏷️ Versioning

When releasing machine-controller-manager-provider-stackit, we follow semantic versioning (see https://semver.org/).

In short:
- ⚠️ a new major version (`vX.0.0`, `X` is bumped)
   - brings new features/refactorings/etc.
   - implies breaking changes to consumers of the package (i.e., incompatible with the last major)
- 🚀 a new minor version (`vX.Y.0`, `Y` is bumped)
   - brings new features/refactorings/etc.
   - does not imply breaking changes (i.e., compatible with the last minor)
- 🚑 a new patch version (`vX.Y.Z`, `Z` is bumped)
   - brings bug fixes without new features/refactorings/etc.
   - does not imply breaking changes (i.e., compatible with the last patch)

For major version changes, the configuration typically needs to be adapted to accommodate breaking changes before successfully upgrading. For minor and patch updates, no configuration adjustments are required.

Both major and minor releases are created from the main branch. Patch releases are created from a release branch that is based on a minor version release.

To make sure we release with the correct version bump, every breaking PR needs to be labeled with the breaking label (e.g., via /label breaking) so that it is automatically categorized correctly when generating release notes.

## 🔖 Publishing a Release

When changes are merged into `main` or a `release-v*` branch, the `release-tool` creates a draft release to preview the upcoming updates.
The tool automatically determines the appropriate version tag based on the target branch and the labels of the merged Pull Requests:

To publish a release, follow these steps:

1. Open the repository's releases page.
2. Navigate to the corresponding draft release (minor/major for `main`, patch for `release-v*`).
3. Review to-be-released changes by checking the release notes.
4. Edit the release by pressing the pen icon.
5. Change `REPLACE_ME` with your github username.
6. Press the "Publish release" button.

## Manual Release Process (Fallback Method)

If the `release-tool` or its associated Prow job fails, use the GitHub web UI to create and publish a release:

1. Go to the repository on GitHub and click **Releases** on the right side, then click **Draft new release**.

2. Open the **Select tag** dropdown and choose **Create new tag** at the bottom. Enter the new tag name (for example `v2.1.0`) and pick the target branch/commit, then confirm.

3. Click **Generate release notes** to let GitHub populate the changelog.

4. In the release description, add a line `Released by @<your github handle>` to indicate the publisher.

5. Click **Publish release** to create the release.

Publishing a new release triggers the same Prow release job that builds and publishes the final container images.
