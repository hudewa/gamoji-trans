# gamoji-trans

A Go command-line tool for scanning, translating, and replacing Chinese text in code files.

## Features

- 🔍 Scan code files for Chinese text wrapped in quotes
- 🌐 Translate Chinese to multiple languages using Doubao AI
- 📝 Generate SQL files for i18n database entries
- 🔄 Replace original Chinese text with `module.identification` format

## Installation

```bash
go install github.com/yourusername/gamoji-trans/cmd/gamoji-trans@latest
```

Or clone and build locally:

```bash
git clone https://github.com/yourusername/gamoji-trans.git
cd gamoji-trans
go build -o gamoji-trans ./cmd/gamoji-trans
```

## Quick Start

1. **Create a configuration file:**

```bash
gamoji-trans init -o config.yaml
```

2. **Edit the configuration file** with your Doubao API key.

3. **Run the full process:**

```bash
gamoji-trans process -d ./src -c ./config.yaml --replace
```

## Commands

### `scan`

Scan directory for Chinese text without translating:

```bash
gamoji-trans scan -d ./src
```

### `translate`

Scan and translate Chinese text, generate SQL file:

```bash
gamoji-trans translate -d ./src -o ./sql -k YOUR_API_KEY
```

### `process`

Full workflow: scan, translate, generate SQL, and optionally replace:

```bash
# Preview replacements without making changes
gamoji-trans process -d ./src --dry-run

# Full process with replacement
gamoji-trans process -d ./src -o ./sql --replace

# Use configuration file
gamoji-trans process -c ./config.yaml --replace
```

### `init`

Create an example configuration file:

```bash
gamoji-trans init -o config.yaml
```

## Configuration

Configuration can be provided via:

1. **Command-line flags** (highest priority)
2. **Configuration file** (`-c` flag or default locations)
3. **Environment variables** (e.g., `DOUBAO_API_KEY`)

### Configuration File Example

```yaml
# Doubao API Configuration
doubao:
  api_key: "your-api-key-here"
  base_url: "https://ark.cn-beijing.volces.com/api/v3"
  model: "doubao-1.5-pro-32k"

# Scanner Configuration
scan:
  include_ext:
    - ".go"
    - ".js"
    - ".vue"
    - ".ts"
    - ".tsx"
    - ".jsx"
    - ".json"
  exclude_dirs:
    - "node_modules"
    - ".git"
    - "dist"
    - "build"
  exclude_patterns:
    - "*_test.go"
    - "*.min.js"

# Output Configuration
output:
  sql_dir: "./sql"
  module_name: "doubao"
  updated_by: "doubao"
```

## Supported Languages

The tool translates Chinese text to the following languages:

- `en` - English
- `id` - Indonesian
- `th` - Thai
- `vi` - Vietnamese
- `ms` - Malay

## SQL Output Format

Generated SQL files follow this format:

```sql
INSERT INTO gamoji.i18n (type, module, identification, zh_lan, en_lan, id_lan, th_lan, vi_lan, ms_lan, updated_by)
VALUES (1, 'doubao', 'a1b2c3d4', '你好世界', 'Hello World', 'Halo Dunia', 'สวัสดีชาวโลก', 'Xin chào thế giới', 'Hai Dunia', 'doubao');
```

## How It Works

1. **Scanning**: Recursively scans specified directories for files with configured extensions
2. **Chinese Detection**: Uses regex patterns to find Chinese text within single or double quotes
3. **Filtering**: Automatically skips image file paths (`.png`, `.webp`, `.jpg`, etc.)
4. **Translation**: Sends unique Chinese texts to Doubao AI for batch translation
5. **SQL Generation**: Creates timestamped SQL files with all translations
6. **Replacement**: Optionally replaces original Chinese text with `module.hash` format

## Development

```bash
# Run tests
go test ./...

# Build
go build -o gamoji-trans ./cmd/gamoji-trans

# Run locally
go run ./cmd/gamoji-trans scan -d ./test
```

## License

MIT
