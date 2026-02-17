# Release Procedure

## Overview

This document outlines the standard procedure for creating new releases of the STACKIT machine-controller-manager.

## General Information

- **Versioning:** Versioning follows official [SemVer 2.0](https://semver.org/)
- **CI/CD System:** All release and image builds are managed by our **Prow CI** infrastructure.

---

## Automated Release Process (Primary Method)

The primary release method is automated using a tool called `release-tool`. This process is designed to be straightforward and require minimal manual intervention.

1. **Draft Creation:** On every successful merge (post-submit) to the `main` branch, a Prow job automatically runs the `release-tool`. This tool creates a new draft release on GitHub or updates the existing one with a changelog generated from recent commits.
2. **Publishing the Release:** When the draft is ready, navigate to the repository's "Releases" page on GitHub. Locate the draft, review the changelog, replace the placeholder with your GitHub handle and publish it by clicking the "Publish release" button.

Publishing the release automatically creates the corresponding Git tag (e.g., `v1.3.1`), which triggers a separate Prow job to build the final container images and attach them to the GitHub release.

---

## Manual Release Process (Fallback Method)

If the `release-tool` or its associated Prow job fails, you can manually trigger a release by creating and pushing a Git tag from the appropriate release branch.

1. **Check out the release branch:** Ensure you have the latest changes from the correct release branch.

   ```bash
   git checkout main
   git pull origin main
   ```

2. **Create the Git tag:** Create a new, annotated tag for the release, following semantic versioning.

   ```bash
   git tag v2.1.0
   ```

3. **Push the tag to the remote repository:**

   ```bash
   git push origin v2.1.0
   ```

Pushing a tag that starts with `v` (e.g., `v2.1.0`) automatically triggers the same Prow release job that builds and publishes the final container images. You may need to manually update the release notes on GitHub afterward.
