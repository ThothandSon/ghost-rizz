# Security policy — `ghost-rizz`

This document does two things:

1. Tells you **what `ghost-rizz` does and does not protect you against**,
   so you can decide whether it fits your threat model.
2. Tells you **how to report a vulnerability** in the tool itself.

It is written for adult users making informed decisions. If you are in
active physical danger and your safety depends on staying invisible to a
specific adversary, do not rely on any single tool — consult a digital
security professional (Access Now Helpline, EFF's SSD, or a local
organization equivalent).

---

## Threat model

### What `ghost-rizz` protects

- Metadata written by cameras, phones and editors into the **EXIF, XMP
  and IPTC blocks** of JPEG and PNG files (and via `exiftool` for HEIC).
- Standard fields that identify device, lens, timestamp, GPS, software,
  author and copyright.
- Bulk processing of thousands of files without loss of parallelism.

For the fields listed as "the seven that make 90% of the damage" in the
companion book — Make, Model, DateTimeOriginal, GPS coordinates,
Software, Artist/Copyright, Serial Number, Lens Model — both `clean` and
`fuzz` produce output that no downstream reader using a standard EXIF
parser can reconstruct back to the original values.

### What `ghost-rizz` DOES NOT protect

The following are **out of scope** and will not be fixed by any release of
this tool. If any of them matter to you, `ghost-rizz` alone is
insufficient.

- **Steganography in the pixels.** Information deliberately hidden
  inside the image pixels (LSB stego, watermark, adversarial patterns)
  survives every operation this tool performs. Only re-encoding through
  a lossy pipeline degrades it, and even that is not reliable.
- **Embedded JPEG thumbnails smaller than the main image.** The main
  image and its thumbnail live in the same file. Some editors, when
  cropping, update the pixels but leave the thumbnail intact — the
  cropped-out area continues to exist in the thumbnail. `ghost-rizz`
  removes the standard EXIF thumbnail block during `clean`, but does
  not guarantee removal of every vendor-specific miniature that a
  particular camera or app may embed elsewhere in the file.
- **PNG obscure chunks.** PNG allows arbitrary application-specific
  ancillary chunks. `ghost-rizz` handles `tEXt`, `iTXt` and `zTXt`
  (the standard text chunks used for metadata). Non-standard chunks
  written by exotic software may pass through untouched.
- **Custom / vendor MakerNote fields.** EXIF has a `MakerNote` block
  where camera makers store proprietary data whose structure is not
  documented. `ghost-rizz` clears the top-level MakerNote pointer, but
  does not attempt to decode the contents of every vendor's blob.
- **Traffic-side profiling.** Metadata is one signal among many that a
  platform, ad network or ISP uses to profile users. Cleaning your file
  does nothing about the timing of when you upload, the account you
  upload with, the fingerprint of your browser, or the IP you upload
  from. Combine with a properly configured VPN/Tor and account hygiene
  if that matters.
- **End-to-end encryption of the file in transit.** `ghost-rizz` writes
  files to your disk. Moving them to another party is out of scope —
  use Signal, Tresorit, Cryptomator or equivalent for that.
- **Court-grade proof of origin.** `clean` and `fuzz` are destructive
  to authenticity signals. If you need a file to serve as evidence
  (yours or a client's), do **not** run either — use `report` only, and
  preserve the original in an untouched chain of custody.
- **File-system side channels.** Filename, modification time, size,
  location on disk and backup history all leak information that
  `ghost-rizz` does not touch. Rename the file, set `touch -d`, or
  store it on an encrypted volume as needed.
- **RAW files, DNG, video, PDF, DOCX.** Out of scope. Use
  [`exiftool`](https://exiftool.org) for these formats.

### Assumptions the tool makes

- You run `ghost-rizz` on a machine you control, with a filesystem you
  trust. It does not defend against a malicious binary posing as
  `ghost-rizz` — verify checksums from the GitHub release page.
- You have a copy of the original file if you might ever need it. Both
  `clean` and `fuzz` write to a separate output folder, but any
  automation you build on top of them is your responsibility.
- The `exiftool` binary used for HEIC handling is the real one, from
  [exiftool.org](https://exiftool.org), and not a tampered copy.

---

## Reporting a vulnerability

If you believe you have found a security-relevant bug in `ghost-rizz`
itself (crash, memory disclosure, incorrect stripping that leaks
supposedly-cleaned data, path traversal, arbitrary code execution),
please **do not open a public issue**.

Instead, send an email to:

> **thothandson@icloud.com**

with subject `ghost-rizz security` and include:

- The version of `ghost-rizz` (`ghost-rizz --version` or the release
  tag).
- Operating system and architecture.
- A minimal reproducer (sample file + command line), when possible.
- Your assessment of impact.

Please give at least **30 days** before public disclosure. A patch
release will be published on the standard GitHub Releases page along
with a note describing the fix (without gratuitously enumerating the
attack).

Cosmetic issues, feature requests and non-security bugs belong in the
public [issues](https://github.com/ThothandSon/ghost-rizz/issues)
tab.

---

## Verifying a release

Every release includes SHA-256 checksums for all binaries. To verify:

```
shasum -a 256 -c ghost-rizz_<version>_checksums.txt
```

If a checksum does not match, do not run the binary — open an issue and
mention which file failed.

---

## Related projects

- **[exiftool](https://exiftool.org)** — the canonical reference for
  reading and writing metadata in dozens of formats.
