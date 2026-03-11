# Release Procedure

## Overview

To push new changes to the release, start with a PR to the main branch. Make sure to follow the PR template.

### 🏷️ Versioning

When releasing machine-controller-manager-provider-stackit, we follow semantic versioning (see https://semver.org/).

In short:
- versions have the `vX.Y.Z` pattern, with `X` being the major version, `Y` being the minor version, and `Z` being the patch version
- ⚠️ a new major version (`vX.0.0`, `X` is bumped)
   - brings new features/refactorings/etc.
   - implies breaking changes to consumers of the package (i.e., incompatible with the last major)
- 🚀 a new minor version (`vX.Y.0`, `Y` is bumped)
   - brings new features/refactorings/etc.
   - does not imply breaking changes (i.e., compatible with the last minor)
- 🚑 a new patch version (`vX.Y.Z`, `Z` is bumped)
   - brings bug fixes without new features/refactorings/etc.
   - does not imply breaking changes (i.e., compatible with the last patch)

For major version changes, consumers typically need to adapt their usage of the package to the breaking changes before they can successfully upgrade. For minor and patch version changes, no adaptions are needed.

In case of machine-controller-manager-provider-stackit, we use the following types of releases:

- ⚠️ a new major version is released if
   - manual steps are required after rolling out the change TODO: ask if it is required
   - Gardener Upgrades
- 🚀 a new minor version is released if
   - it doesn't come with a known impact on customers
   - it can be promoted automatically
- 🚑 a new patch version is released if
   - a critical change needs to be rolled out without other major/minor changes
   - it can be promoted automatically

## General Information

- **Versioning:** Versioning follows official [SemVer 2.0](https://semver.org/)
- **CI/CD System:** All release and image builds are managed by our **Prow CI** infrastructure.

## Automated Release Process (Primary Method)

The primary release method is automated using a tool called `release-tool`. This process is designed to be straightforward and require minimal manual intervention.

1. **Draft Creation:** On every successful merge (post-submit) to the `main` branch, a Prow job automatically runs the `release-tool`. This tool creates a new draft release on GitHub or updates the existing one with a changelog generated from recent commits.
2. **Publishing the Release:** When the draft is ready, navigate to the repository's "Releases" page on GitHub. Locate the draft, review the changelog, replace the placeholder with your GitHub handle and publish it by clicking the "Publish release" button.

Publishing the release automatically creates the corresponding Git tag (e.g., `v1.3.1`), which triggers a separate Prow job to build the final container images and attach them to the GitHub release.

## Manual Release Process (Fallback Method)

If the `release-tool` or its associated Prow job fails, use the GitHub web UI to create and publish a release:

1. Go to the repository on GitHub and click **Releases** on the right side, then click **Draft new release**.

2. Open the **Select tag** dropdown and choose **Create new tag** at the bottom. Enter the new tag name (for example `v2.1.0`) and pick the target branch/commit, then confirm.

3. Click **Generate release notes** to let GitHub populate the changelog.

4. In the release description, add a line `Released by @<your github handle>` to indicate the publisher.

5. Click **Publish release** to create the release.

Publishing a new release triggers the same Prow release job that builds and publishes the final container images.
