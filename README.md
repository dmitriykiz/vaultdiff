# vaultdiff

A CLI tool to compare two HashiCorp Vault secret paths and output structured diffs.

---

## Installation

```bash
go install github.com/yourusername/vaultdiff@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultdiff.git
cd vaultdiff
go build -o vaultdiff .
```

---

## Usage

Ensure your Vault environment variables are set (`VAULT_ADDR`, `VAULT_TOKEN`), then run:

```bash
vaultdiff <path-a> <path-b>
```

**Example:**

```bash
vaultdiff secret/data/app/staging secret/data/app/production
```

**Sample output:**

```
~ db_password   : "hunter2" → "c0rrectH0rse"
+ new_feature   : "enabled"
- deprecated_key: "old_value"
```

| Symbol | Meaning              |
|--------|----------------------|
| `~`    | Value changed        |
| `+`    | Key added in path-b  |
| `-`    | Key removed in path-b|

### Flags

| Flag         | Description                        | Default |
|--------------|------------------------------------|---------|
| `--format`   | Output format: `text`, `json`      | `text`  |
| `--no-color` | Disable colored output             | `false` |
| `--version`  | Print version and exit             |         |

---

## Requirements

- Go 1.21+
- HashiCorp Vault with a valid token and network access

---

## License

MIT © 2024 Your Name