# ScrollScreen 1.2

A portable screenshot utility for Windows x64. The capture region is selected manually; the user scrolls the target window and ScrollScreen attempts to assemble a vertically scrolling panorama. An internal editor offers cropping, redaction with solid rectangles, zoom, and redacted PNG/clipboard export.

**Status:** Pre-release source review. Windows GUI features still require manual testing. No binary is signed or approved by SignPath Foundation.

## Build locally

Requires Go and a Windows target. From PowerShell in this directory:

```powershell
$env:GO111MODULE = 'off'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
go build -trimpath -ldflags='-H=windowsgui -s -w' -o ScrollScreen.exe main.go
```

The GitHub Actions workflow performs the same build on a Windows runner. It does **not** sign the executable.

## Usage and limitations

1. Select a region, start capture and scroll down manually.
2. Stop and assemble, then open the editor.
3. Crop by dragging a rectangle and confirming with Enter; Esc cancels it.
4. Redact sensitive information with opaque masks before copying or saving the PNG.

Horizontal and upward stitching is not yet supported reliably. GUI behavior has not been validated end-to-end on a Windows desktop in this review package. The interface uses labeled glyphs, not a finalized native icon set.

## Privacy and security

- The application captures the selected display area, potentially including confidential content. Do not share screenshots before reviewing redactions.
- Image masks are applied to a fresh pixel buffer for PNG and clipboard export, but the original panorama remains in memory while the program runs.
- The application does not write capture diagnostics or raw frames to disk. ZIP exports persist in the OS temporary directory until manually removed.
- Do not include screenshots, diagnostic captures, configuration files or internal system details in Git commits or support issues.
- The Windows build is **unsigned**. Never disable enterprise security controls to run it without authorization.

## Before going public

- Confirm that you own the code and every resource and have your organization's approval to publish it.
- The repository contains an MIT LICENSE. Confirm copyright ownership and organizational authorization before publication.
- Confirm rights and attribution for artwork/icons. This review package intentionally omits icon binaries and generated resources.
- Complete code/security review, GUI testing, public release, MFA setup and repository ownership controls.
- Read [Code signing policy](CODE_SIGNING_POLICY.md) before any SignPath application.

## Development

**Project author and maintainer:** usermister

**Developed with AI assistance (OpenAI ChatGPT).**

ScrollScreen is an open-source Windows application developed by usermister with the assistance of artificial intelligence for code generation, debugging, user interface design, and documentation.

## License

Distributed under the MIT License. See [LICENSE](LICENSE) for details.

## Security review notes (2026-11-12)

- Raw-frame diagnostic capture and automatic diagnostic text-file creation have been removed from this revision.
- Two Windows clipboard buffer conversions use `unsafe.Pointer` for memory obtained via `GlobalLock`. The `go vet` `unsafeptr` analyzer is excluded temporarily; these two blocks still require Windows API lifetime review and desktop testing.
- Unsanitized image pixels remain in process memory until released and may be exposed through operating-system memory dumps.
- ZIP exports are left in `%TEMP%\ScrollScreen_1_1` for compatibility with clipboard file pasting. Remove them manually after sharing.
- The application is not signed and no result here guarantees Smart App Control will allow it.

## Development

**Project author and maintainer:** usermister

Developed with AI assistance (OpenAI ChatGPT). The maintainer is responsible for reviewing and maintaining the code.

## License

Distributed under the MIT License. See [LICENSE](LICENSE).


## Code Signing

ScrollScreen is currently distributed as an unsigned
open-source application.

The project intends to use SignPath Foundation
for free open-source code signing.

Code signing has not yet been approved or enabled.

For details, see our
[Code Signing Policy](CODE_SIGNING_POLICY.md).

