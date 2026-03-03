# q

A CLI tool to quickly query an LLM from your terminal.

```
$ q what is a goroutine
A goroutine is a lightweight thread managed by the Go runtime, started with the go keyword.
```

## Install

Requires Go 1.25+.

```sh
go install github.com/yannistarfa/q@latest
```

Or build from source:

```sh
git clone https://github.com/ytarfa/q.git
cd q
go build -o q .
```

## Setup

Run `q init` to create a config file interactively:

```
$ q init
Provider (openai/ollama) [openai]: ollama
Model [llama3]: llama3.2:1b
Config written to /home/user/.config/q/config.json
```

This creates `~/.config/q/config.json` with 0600 permissions.

### Providers

**OpenAI** — requires an API key:

```
$ q init
Provider (openai/ollama) [openai]: openai
Model [gpt-4o-mini]: gpt-4o-mini
API key: sk-...
```

**Ollama** — uses your local Ollama instance at `localhost:11434`, no API key needed.

## Usage

```
q [flags] [question...]
```

### Examples

```sh
# Ask a question
q what is a closure in javascript

# Longer response
q -l 500 explain the CAP theorem

# Unlimited response length
q -l 0 write a haiku about recursion

# Pipe stdin as context
cat error.log | q what does this mean

# Stdin with no question defaults to "Explain this"
echo "func main() {}" | q
```

### Flags

| Flag | Description |
|---|---|
| `-l`, `--limit` | Max response length in characters (default: 200, 0 = unlimited) |
| `-v`, `--version` | Print version |
| `-h`, `--help` | Print help |

### Subcommands

| Command | Description |
|---|---|
| `init` | Create config file interactively |

## Configuration

### Config file

Located at `~/.config/q/config.json`:

```json
{
  "limit": 200,
  "provider": {
    "type": "ollama",
    "model": "llama3.2:1b",
    "api_key": "",
    "base_url": ""
  }
}
```

Leave `base_url` empty to use the provider default (`https://api.openai.com/v1` for OpenAI, `http://localhost:11434/v1` for Ollama).

### Environment variables

| Variable | Description |
|---|---|
| `Q_API_KEY` | API key (overrides config) |
| `Q_MODEL` | Model name (overrides config) |
| `Q_PROVIDER` | Provider type (overrides config) |
| `Q_LIMIT` | Character limit (overrides config) |

### Precedence

CLI flag > environment variable > config file > built-in default.

## Exit codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | Usage error |
| 2 | Config error |
| 3 | API error |

## License

MIT
