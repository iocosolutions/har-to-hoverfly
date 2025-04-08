# har-to-hoverfly

## Converts a HAR file to a Hoverfly simulation file for API mocking and testing

This CLI tool ingests a HAR (HTTP Archive) file and produces a Hoverfly-compatible simulation. Useful for simulating real traffic and testing systems in isolation.

### Features

- Filters entries by MIME type
- Skips or includes only specific content types via `--allowed-content-types`
- Optionally omits non-text entries
- Summarise mode shows traffic structure
- Supports limiting body size
- Allows host restriction

### Usage

```bash
har-to-hoverfly --input <file.har> [flags]
```

### Flags

| Flag                      | Description                                                                 |
|---------------------------|-----------------------------------------------------------------------------|
| `--input`                | Path to the input HAR file (required)                                       |
| `--output`               | Output simulation JSON file path (optional, defaults to stdout)             |
| `--max-body-bytes`       | Max body size for responses; truncate if exceeded                           |
| `--ignore-non-text`      | Completely ignore non-text MIME types                                       |
| `--allowed-content-types`| Comma-separated list of allowed substrings in MIME types                    |
| `--host`                 | Restrict processing to entries for a specific destination host              |
| `--summarise`            | Outputs a summary table grouped by host, method, path                       |

### Example

```bash
har-to-hoverfly --input session.har --output simulation.json --ignore-non-text --allowed-content-types json,xml
```

This processes a HAR file, includes only JSON/XML responses, and outputs a Hoverfly simulation file.

---

© 2024 IOCO Solutions 
