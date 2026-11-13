# Clean repository notes

This package contains only a snapshot of reviewed source files and configurations; it contains no `.git` directory or commit history. It was prepared from the previous security-review source and the Go tests provided in the conversation. It is **not** a certification of security or functional correctness.

Before public release, complete the rights and security review, test the Windows user interface (capture, crop, masks, export and clipboard), and address the `unsafe.Pointer` warning rather than permanently suppressing it. No signing has been performed. GitHub Actions builds unsigned executables only. Do not publish an untested build as a verified stable release.

The original private repository should remain private until reviewed; deleting it is not necessary. Switching that old repository to Public would expose its complete history.
