---
seo:
  title: Telemetry
  description: PPHLX collects anonymous telemetry data about general usage to help inform our roadmap. Participation is optional and you may opt-out at any time.
updated_date: 2026-05-28
---

The `pphlx` CLI collects **anonymous telemetry data** about general usage. Participation is optional, and you may opt-out at any time.

## Why is telemetry collected?

Anonymous telemetry data makes up an important piece of our roadmap prioritization process. It gives our core team insight into broad trends in feature usage, pain points, and configuration which then helps us make more informed decisions about what to work on next.

## What is being collected?

We track general usage information about PPHLX and different configuration options that we support. Specifically, we track the following anonymous data:

- Command invoked (`pphlx build`, `pphlx dev`, `pphlx preview`, etc.)
- PPHLX version
- General machine information (e.g. number of CPUs, macOS/Windows/Linux, CI environment, etc.)
- Development environment (e.g. whether the CLI was invoked by an AI coding agent)
- General configuration information (Integrations, adapters, markdown options, etc.)
- Sanitized error information
- Toolbar usage (e.g. built-in apps being used)

This list is regularly audited to ensure its accuracy. You can audit telemetry yourself locally by settings the `DEBUG=”pphlx:telemetry”` environment variable when running the PPHLX CLI. While in debug mode, telemetry events are only logged to the console.

An example telemetry event might look like this:

```json
{
  "cliCommand": "dev",
  "config": {
    "markdownPlugins": [],
    "adapter": "@pphlx/adapter-default",
    "integrations": ["@pphlx/mdx", "@pphlx/rss"],
    "markdown": {
      "syntaxHighlight": "shiki"
    }
  }
}
```

## What about sensitive data?

We **do not** collect any metrics which may contain sensitive data. This includes, but is not limited to: environment variables, personally identifiable information, file paths, contents of files, logs, stack traces, git remote information, or unsanitized JavaScript error messages.

We take your privacy very seriously. You can read our [privacy policy](https://pphlx.org/privacy/) to learn more.

## How is my data protected?

The data collected by the PPHLX CLI is completely anonymous and only meaningful in aggregate form. The data that we store is not traceable to the source, and is only available to a small subset of the PPHLX core maintainer team to help inform our roadmap prioritization.

The PPHLX CLI does not track, collect, or store personally identifiable information (PII). The telemetry data that we do track has never and will never be sold or monetized in any form.

## How do I opt out?

You can always opt-out by running `pphlx telemetry disable` in the root of any PPHLX project directory. This will disable telemetry across your entire machine, not just the project directory that you run it in:

```bash
pphlx telemetry disable
```

You can re-enable telemetry at any time by running `pphlx telemetry enable` in the root of any PPHLX project directory. This will disable telemetry across your entire machine, not just the project that you run it in:

```bash
pphlx telemetry enable
```

You can also opt-out by setting the environment variable: `PPHLX_TELEMETRY_DISABLED=1`. If this environment variable is set when PPHLX runs, it will disable all telemetry.
