# PPHLX Compiler v1.1.9 Release Notes

## 🚀 What's New in v1.1.9

### ⚡ Native PHP CGI Dev Server Execution (`pphlx dev`)
- **Full Dynamic PHP Serving**: Dynamic execution of `.php` backend scripts and API endpoints via `php-cgi` during local development, supporting HTTP redirects (`Location:` headers), sessions (`PHPSESSID`), cookie persistence, and JSON responses.
- **Cross-Platform `php-cgi` Discovery**: Automatic discovery of `php-cgi` across system `PATH` and standard platform installation directories.
- **Extensionless Routing**: Clean extensionless routing for PHP endpoints (e.g. `/api/users` dynamically resolves and executes `src/api/users.php`).
- **Environment Variable Inheritance**: Child CGI processes automatically inherit environment variables from project `.env` files.
- **Source Code Protection**: Built-in protection preventing raw `.php` files from being served as downloadable static assets.

---

## 🔒 SHA256 Checksums

| Archive File | SHA256 Checksum |
| :--- | :--- |
| `pphlx-darwin-arm64.tar.gz` | `d5d4d2530919746a5382d7b1054efc5012ebd9f1cf9b085f7e255012c7e463a6` |
| `pphlx-darwin-amd64.tar.gz` | `b08b87ff7fddedcd84dd1387a74217bfde8d5fd7a9bf83c858e79f2db5ae9494` |
| `pphlx-linux-arm64.tar.gz`  | `cf38dda7e0c7981b5f8bca6337332a0f1d883636878be99287d92ab742eca9e4` |
| `pphlx-linux-amd64.tar.gz`  | `38b05275770cf51d1c66c7bf3ca50ba0fdacd0e5809efc49f3cab4c4d6db4e71` |
| `pphlx-windows-amd64.zip`  | `9236c04e187e2781f10da6641c354fb2294ff53aca8d0e3213f2045a6e72a6ba` |
| `pphlx.msi`                | `9261e469cc808ca0d5228937e3a183621382638f5a07fea442eaac6bb613bb9a` |
