# IPFilter Caddy Plugin

[![Go Reference](https://pkg.go.dev/badge/github.com/jpillora/ipfilter-caddy.svg)](https://pkg.go.dev/github.com/jpillora/ipfilter-caddy)
[![Go Report Card](https://goreportcard.com/badge/github.com/jpillora/ipfilter-caddy)](https://goreportcard.com/report/github.com/jpillora/ipfilter-caddy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Caddy v2 plugin that provides geolocation-based request filtering using the jpillora/ipfilter library with IP2Location LITE data. Perfect for restricting access to specific countries without requiring external databases or API keys.

## ✨ Features

- 🚀 **Zero Configuration**: No database downloads, API keys, or external dependencies required
- 🌍 **Country-Based Filtering**: Allow or deny requests based on visitor country
- 🔒 **Free Geolocation Data**: Uses IP2Location LITE database (completely free)
- ⚡ **High Performance**: Embedded geolocation data with no network calls
- 🛡️ **Thread-Safe**: Safe for concurrent request handling
- 📝 **Full Caddy Support**: JSON and Caddyfile configuration support
- 🔄 **Easy Integration**: Drop-in Caddy HTTP matcher

## 📦 Installation

### Option 1: Download Pre-built Binary

Download a Caddy binary with this plugin pre-installed from the [releases page](https://github.com/jpillora/ipfilter-caddy/releases).

### Option 2: Build from Source

1. **Install xcaddy** (if not already installed):
   ```bash
   go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
   ```

2. **Build Caddy with the plugin**:
   ```bash
   xcaddy build --with github.com/jpillora/ipfilter-caddy
   ```

3. **Replace your existing Caddy binary**:
   ```bash
   sudo cp caddy /usr/bin/caddy
   sudo systemctl restart caddy
   ```

### Option 3: Add to Existing Installation

If you have Caddy installed via package manager:

```bash
caddy add-package github.com/jpillora/ipfilter-caddy
```

## 🚀 Usage

### JSON Configuration

```json
{
  "apps": {
    "http": {
      "servers": {
        "example": {
          "listen": [":443"],
          "routes": [
            {
              "match": [
                {
                  "ipfilter_geolocation": {
                    "allow_countries": ["AU", "US", "CA"],
                    "block_by_default": true
                  }
                }
              ],
              "handle": [
                {
                  "handler": "static_response",
                  "body": "Welcome from allowed countries!"
                }
              ]
            },
            {
              "handle": [
                {
                  "handler": "static_response",
                  "status_code": 403,
                  "body": "Access restricted"
                }
              ]
            }
          ]
        }
      }
    }
  }
}
```

### Caddyfile Configuration

```caddyfile
{
    order ipfilter_geolocation first
}

example.com {
    @allowed_countries {
        ipfilter_geolocation {
            allow_countries AU US CA
            block_by_default true
        }
    }

    handle @allowed_countries {
        respond "Welcome from Australia, USA, or Canada!"
    }

    handle {
        respond "Access restricted to specified countries" 403
    }
}

# Block specific countries
api.example.com {
    @blocked_countries {
        ipfilter_geolocation {
            deny_countries RU CN
        }
    }

    handle @blocked_countries {
        respond "Access denied from restricted countries" 403
    }

    # Continue with normal processing for allowed countries
    reverse_proxy localhost:8080
}
```

## ⚙️ Configuration Options

### JSON Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `allow_countries` | `[]string` | `[]` | List of ISO country codes to allow |
| `deny_countries` | `[]string` | `[]` | List of ISO country codes to deny |
| `block_by_default` | `bool` | `false` | Block all requests by default unless explicitly allowed |

### Caddyfile Directives

```caddyfile
ipfilter_geolocation {
    allow_countries <country_codes...>
    deny_countries <country_codes...>
    block_by_default <true|false>
}
```

## 🌍 Country Codes

Uses standard [ISO 3166-1 alpha-2](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2) country codes:

- `AU` - Australia
- `US` - United States
- `CA` - Canada
- `GB` - United Kingdom
- `DE` - Germany
- `FR` - France
- `JP` - Japan
- `CN` - China
- `RU` - Russia
- And [many more](https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2)...

**Note**: Unknown or private IPs (like `127.0.0.1`, `::1`) return country code `ZZ`.

## 📋 Examples

### 1. Australia-Only Access

```caddyfile
australia-only.example.com {
    @australia {
        ipfilter_geolocation {
            allow_countries AU
            block_by_default true
        }
    }

    handle @australia {
        respond "G'day! Welcome from Australia 🇦🇺"
    }

    handle {
        respond "This site is only accessible from Australia" 403
    }
}
```

### 2. Block Bad Actors

```caddyfile
api.example.com {
    @restricted {
        ipfilter_geolocation {
            deny_countries RU CN KP
        }
    }

    handle @restricted {
        respond "Access denied" 403
    }

    reverse_proxy localhost:3000
}
```

### 3. Multi-Country Support

```caddyfile
global.example.com {
    @eu_countries {
        ipfilter_geolocation {
            allow_countries DE FR GB IT ES NL
            block_by_default true
        }
    }

    handle @eu_countries {
        respond "Welcome from Europe! 🇪🇺"
    }

    handle {
        respond "This content is geo-restricted to European countries" 403
    }
}
```

### 4. Debug Logging

Enable debug logging to see IP-to-country mappings:

```json
{
  "logging": {
    "logs": {
      "default": {
        "level": "debug"
      }
    }
  }
}
```

This will show logs like:
```
{"level":"debug","logger":"http.matchers.ipfilter_geolocation","msg":"IPFilter geolocation check","ip":"1.2.3.4","country":"AU","allowed":true}
```

## 🔧 Technical Details

### Plugin Architecture

- **Module ID**: `http.matchers.ipfilter_geolocation`
- **Type**: HTTP Request Matcher
- **Dependencies**: jpillora/ipfilter (with embedded IP2Location LITE data)

### How It Works

1. **IP Extraction**: Gets the client IP from Caddy's request context
2. **Geolocation Lookup**: Uses embedded IP2Location database to find country
3. **Rule Evaluation**: Applies allow/deny rules based on configuration
4. **Request Routing**: Continues or blocks request processing

### Performance

- **Lookup Time**: Sub-millisecond geolocation lookups
- **Memory Usage**: ~50MB for embedded database
- **Thread Safety**: Safe for concurrent access
- **No External Calls**: All data is local

### Accuracy Notes

IP geolocation is inherently approximate:
- Mobile networks may show incorrect locations
- VPNs and proxies can mask real locations
- Corporate networks often appear in headquarters country
- Accuracy is typically 95%+ for broadband connections

## 🧪 Testing

### Manual Testing

1. **Start your server**:
   ```bash
   caddy run --config your-config.json
   ```

2. **Test from different locations**:
   ```bash
   # Test from current location
   curl https://your-domain.com/

   # Test with VPN (if available)
   # Connect to VPN in different country and test again
   ```

3. **Check logs** for geolocation data:
   ```bash
   caddy run --config your-config.json 2>&1 | grep "IPFilter geolocation"
   ```

### Automated Testing

```go
package main

import (
    "net/http"
    "net/http/httptest"
    "github.com/caddyserver/caddy/v2/modules/caddyhttp"
    "github.com/jpillora/ipfilter-caddy"
)

func TestAustraliaFilter(t *testing.T) {
    // Create matcher
    matcher := &IPFilterGeolocation{
        AllowCountries: []string{"AU"},
        BlockByDefault: true,
    }

    // Mock Australian IP request
    req := httptest.NewRequest("GET", "/", nil)
    req = req.WithContext(caddyhttp.WithRemoteAddr(req.Context(), "1.1.1.1:12345")) // Cloudflare AU IP

    if !matcher.Match(req) {
        t.Error("Expected Australian IP to be allowed")
    }
}
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

### Development Setup

1. **Clone the repository**:
   ```bash
   git clone https://github.com/jpillora/ipfilter-caddy.git
   cd ipfilter-caddy
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Build and test**:
   ```bash
   go build
   go test ./...
   ```

4. **Build with Caddy**:
   ```bash
   xcaddy build --with ./...
   ```

### Guidelines

- Follow Go best practices
- Add tests for new features
- Update documentation
- Ensure thread safety
- Keep dependencies minimal

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Caddy](https://caddyserver.com/) - The extensible web server
- [jpillora/ipfilter](https://github.com/jpillora/ipfilter) - The underlying filtering library
- [IP2Location](https://www.ip2location.com/) - Free geolocation database

## 📞 Support

- 📖 [Documentation](https://github.com/jpillora/ipfilter-caddy#readme)
- 🐛 [Issues](https://github.com/jpillora/ipfilter-caddy/issues)
- 💬 [Discussions](https://github.com/jpillora/ipfilter-caddy/discussions)

---

**Made with ❤️ for the Caddy community**