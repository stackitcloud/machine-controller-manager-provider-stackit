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

Both major and minor releases are created from the main branch. Patch releases are created from a release branch that is based on a minor version release.

To make sure we release with the correct version bump, every breaking PR needs to be labeled with the breaking label (e.g., via /label breaking) so that it is automatically categorized correctly when generating release notes.

### 🚀 Basic Promotion (Minor Releases)

> [!NOTE]
> We are now reconciling in maintenance only.
> There is no more reconciliation every hour!

The basic promotion workflow is quite simple:

1. Create a PR to merge your changes on [:octocat: machine-controller-manager-provider-stackit](https://github.com/stackitcloud/machine-controller-manager-provider-stackit) into the `main` branch.
2. Once you have verified the changes on dev [TODO: if the dev is okay to write here], you can promote the changes by creating a new minor release of ske-base.
   For this, publish the draft release on the `main` branch for the next minor version (`vx.y.0`) (see [Publishing a Release](#-publishing-a-release)).
   Review the release notes and make sure, the changes since the last release don't contain any breaking changes.

### ⚠️ Promotion with Adaptions (Major Releases)

When promoting major releases of machine-controller-manager-provider-stackit, the promotion workflow is the following:

1. Create a PR to merge your changes on [:octocat: machine-controller-manager-provider-stackit](https://github.com/stackitcloud/machine-controller-manager-provider-stackit) into the `main` branch. Add `/label breaking` to your PR description.
2. Once you have verified the changes, you can promote the changes by creating a new major release of ske-base.
   For this, publish the draft release on the `main` branch for the next major version (`vx.0.0`) (see [Publishing a Release](#-publishing-a-release)).

TODO: should we add hofixes part?


## 🔖 Publishing a Release

When changes are merged into `main` or a `release-v*` branch, the [release-tool](https://github.com/stackitcloud/ske-ci-infra/blob/main/docs/release-tool.md) generates a draft release for the next release on the branch as a preview of the to-be-released changes.
It automatically chooses the correct tag. I.e., it bumps the major version if there are unreleased breaking changes on `main`, or otherwise bumps the minor version for `main` or bumps the patch version for `release-v*` branches.

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
