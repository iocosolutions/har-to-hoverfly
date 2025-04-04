# har-to-hoverfly
## Converts a HAR file into a Hoverfly simulation file (v5) or summarises request/response data

This tool reads a HAR file (HTTP Archive) and either:
- Outputs an equivalent [Hoverfly](https://hoverfly.io) simulation file, or
- Summarises the HTTP traffic by grouping request/response pairs by host

The generated Hoverfly simulation supports exact matchers for method, path, query parameters, headers, and body. The summary mode provides a quick overview of API activity across hosts.

### Usage:
```bash
har-to-hoverfly [--output file.json] [--max-response-bytes N] [--skip-non-text] [--host hostname] [--summarise] <input.har>
```

### Options:
- `--output`: Write simulation output to a file (default is stdout)
- `--max-response-bytes`: Omit or replace large response bodies (e.g. images, binaries)
- `--skip-non-text`: Replace non-text responses with `NON_TEXT_RESPONSE_SKIPPED`
- `--host`: Only include traffic for a specific destination host
- `--summarise`: Print a summary table instead of a Hoverfly simulation file

### Example: Convert a HAR to Hoverfly simulation
```bash
har-to-hoverfly --output simulation.json input.har
```

### Example: Summarise request/response activity
```bash
har-to-hoverfly --summarise input.har
```

### Example: Convert a HAR to Hoverfly simulation but only include text based content type responses that come from requests made to hoverfly.io and only include responses less than 10240 bytes
```bash
har-to-hoverfly --output simulation.json --skip-non-text --host hoverfly.io --max-response-bytes 10240 input.har
```

### Example: Summarise request/response activity but only include text based content type responses that come from requests made to hoverfly.io and only include responses less than 10240 bytes
```bash
har-to-hoverfly --summarise --skip-non-text --host hoverfly.io --max-response-bytes 10240 input.har
```


Will return to the console:

```
HOST: api.example.com
  METHOD     PATH                                               QUERY                                            
  GET        /api/products                                      category=shoes&limit=5                          
  POST       /api/shoe                                          product=shirt                                   
```


