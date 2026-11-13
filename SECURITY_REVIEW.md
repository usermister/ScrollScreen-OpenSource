# ScrollScreen 1.2 source review (2026-11-12)

Scope: repository source only, without executing the Windows GUI. This is not a security certification.

## Findings

1. HIGH privacy risk (addressed in this revision): unredacted screenshots could be written to a `capture_session_*` directory by an optional raw-frame debug path; diagnostic logs were created beside the EXE on every capture. Those code paths have been removed.
2. MEDIUM (open): ZIP clipboard export leaves a redacted ZIP file in the user TEMP folder. This is necessary for clipboard file-path pasting; delete after use.
3. MEDIUM (open): unredacted original image and frames reside in RAM while running. Operating-system memory capture or process dumping could expose them.
4. MEDIUM (open): both clipboard writers use an unsafe conversion from Windows GlobalLock pointer (previously flagged by go vet on lines 823 and 893 in the older version). Disabling the vet `unsafeptr` checker does NOT resolve this; review Win32 API lifetime and test on Windows.
5. MEDIUM (open): GUI cropping, image masking, DPI scaling, and clipboard formats need manual Windows regression testing.
6. LOW (addressed): About dialog was still reporting version 1.1; updated to 1.2.
7. Review publication ownership and rights; success of compilation is not permission for public release or a guarantee of code-signing acceptance.

No obvious embedded passwords, tokens, corporate hostnames, email addresses or static internal IPs were found in supplied Go source. This string search is not a complete secret scan.

## Validation

Run (Windows runner):

    $env:GO111MODULE='off'
    go vet -unsafeptr=false main.go
    go build -trimpath -ldflags='-H=windowsgui -s -w' -o ScrollScreen.exe main.go

Do not claim full `go vet` passes until the `unsafe.Pointer` findings have been reviewed or fixed. Do not make public until review and permissions are complete.
