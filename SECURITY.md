# Security Policy

The PPHLX core team takes the security of the PPHLX compiler and generated code seriously. We appreciate your efforts to responsibly disclose vulnerabilities.

---

## Supported Versions

We provide security updates for the following versions of PPHLX:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

---

## Reporting a Vulnerability

If you discover a security vulnerability in PPHLX Core or the generated PHP templates:

1. **Do NOT open a public GitHub issue** for security vulnerabilities.
2. Email your report directly to **`security@pphlx.org`**.
3. Include detailed steps to reproduce the issue, a minimal `.pphx` template snippet, and any impact details.

---

## Response Timeline

* **Initial Acknowledgment:** Within 48 hours of receipt.
* **Vulnerability Assessment & Fix:** We aim to release a patch release within 7 business days for critical vulnerabilities.
* **Public Disclosure:** A public CVE/release advisory will be published after a fix is available.

---

## Security Practices in PPHLX Compiler

* **Delimiter Tokenization Safety:** The PPHLX compiler tokenizes attributes and PHP blocks (`<\?php.*?\?>` and `{|= ... |}`) to prevent attribute injection or premature parser closure.
* **Zero Runtime Dependencies:** Production output `.php` files contain no dynamic runtime evaluations (`eval()`), keeping production overhead light and secure.
