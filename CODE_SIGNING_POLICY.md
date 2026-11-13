# Code signing policy (draft, not active)

**Current status:** No trusted code signing has been approved or enabled. No files are signed by SignPath Foundation or any other code-signing service.

## Planned scope

Only `ScrollScreen.exe` for Windows x64, built from this repository in a reviewed, versioned GitHub Actions workflow, may be proposed for signing. Build input must match the reviewed release tag; no locally supplied or manually modified executable may be substituted.

## Proposed release process

1. Maintainer reviews changes and security implications.
2. A release tag triggers a trusted workflow that builds the EXE reproducibly where feasible.
3. Designated release approver reviews the release artifact and signing request.
4. After SignPath Foundation accepts the project and origin verification is configured, the automated build may request a signature.
5. Publish signed artifact, version/tag, and checksums; never relabel an unsigned artifact as signed.
6. If an artifact or signer is compromised, stop releases, remove the artifact and contact the signing provider.


## Roles and identity

Project owner and maintainer: [usermister](https://github.com/usermister)

- Author / Committer: usermister
- Reviewer: usermister (for external pull requests)
- Release and signing approver: usermister

The project is currently maintained by one person.

The maintainer uses multi-factor authentication
for GitHub access.

Every signing request must be explicitly approved
by the maintainer before release.


## Third-party signing

The project intends to apply for free OSS signing via SignPath Foundation, but the project has **not been accepted**. See https://signpath.org/terms for requirements. Approval and Smart App Control acceptance are not guaranteed.
