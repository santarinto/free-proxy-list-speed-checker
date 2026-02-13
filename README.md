# Free Proxy List Speed Checker

A Go application for checking the speed and availability of free proxy servers from [gfpcom/free-proxy-list](https://github.com/gfpcom/free-proxy-list).

## Features

- **Configuration Management**: Supports TOML configuration files with local override support
- **Cache System**: Built-in caching mechanism with a configurable directory
- **Proxy Collection**: Fetches proxy lists from various sources (SOCKS5 supported)
- **Config Patching**: Apply local configuration patches without modifying the main config file

## Installation

```bash
go get -u free-proxy-list-speed-checker
```

## Configuration

The application uses TOML configuration files located in the `config/` directory:

- `config.toml` - Main configuration file
- `config.local.toml` - Optional local overrides (git-ignored)

### Configuration Structure

```toml
app_name = "free-proxy-list-speed-checker"
source_repo_url = "https://github.com/gfpcom/free-proxy-list"

[proxy_collection_list]
socks5 = "https://raw.githubusercontent.com/wiki/gfpcom/free-proxy-list/lists/socks5.txt"

[options]
cache_dir = "var/cache"
```

### Configuration Options

- `app_name`: Application name
- `source_repo_url`: Source repository URL
- `proxy_collection_list.socks5`: URL to SOCKS5 proxy list
- `options.cache_dir`: Directory for caching data

## Usage

Run with the default configuration:

```bash
go run main.go
```

Run with custom configuration file:

```bash
go run main.go -config path/to/config.toml
```

## Project Structure

```
.
├── config/              # Configuration files
│   ├── config.toml      # Main configuration
│   └── config.local.toml # Local overrides (git-ignored)
├── internal/
│   ├── cache/           # Cache management
│   └── config/          # Configuration loading and patching
├── var/                 # Variable data (cache, etc.)
├── main.go              # Application entry point
└── go.mod               # Go module definition
```

## Requirements

- Go 1.25 or higher
- Dependencies:
    - `github.com/BurntSushi/toml` v1.6.0

## License

See the source repository for license information.

## Source

Based on proxy lists from [gfpcom/free-proxy-list](https://github.com/gfpcom/free-proxy-list)