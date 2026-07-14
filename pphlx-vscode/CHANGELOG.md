# Changelog

All notable changes to the PPHLX Visual Studio Code extension will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.4] - 2026-07-15

### Changed
- Bumped version to synchronize extension with PPHLX Compiler Suite v1.0.4.

---

## [1.0.1] - 2026-07-14

### Added
- Added custom light and dark mode file icon definitions (`./logo/file-icon.svg`) to enhance workspace UI.

### Fixed
- Fixed auto-closing pairs behavior for brace-pipe delimiters (`{|` and `{|=`), preventing unnecessary trailing spaces.
- Improved TextMate grammar scope resolutions for embedded CSS and JS blocks inside template islands.
- Resolved bracket matching conflicts when nested under complex nested conditions.

---

## [1.0.0] - 2026-02-18

### Added
- Initial Release of official PPHLX Extension support for VS Code.
- Native TextMate grammar parser (`pphlx.tmLanguage.json`) for custom `.pphx` template structures.
- Support for brace-pipe delimiters: logical scopes `{| ... |}` and echo scopes `{|= ... |}`.
- Intelligent brace completion configurations (`language-configuration.json`).
- Automatically maps `.pphx` file extensions with the custom PPHLX language identifier.
