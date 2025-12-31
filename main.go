package main

import (
	"fmt"
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	caddycmd "github.com/caddyserver/caddy/v2/cmd"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/jpillora/ipfilter"
	"go.uber.org/zap"
)

func main() {
	caddycmd.Main()
}

// Interface guards
var (
	_ caddy.Module             = (*IPFilterGeolocation)(nil)
	_ caddyhttp.RequestMatcher = (*IPFilterGeolocation)(nil)
	_ caddy.Provisioner        = (*IPFilterGeolocation)(nil)
	_ caddy.CleanerUpper       = (*IPFilterGeolocation)(nil)
	_ caddyfile.Unmarshaler    = (*IPFilterGeolocation)(nil)
)

func init() {
	caddy.RegisterModule(IPFilterGeolocation{})
}

// IPFilterGeolocation allows filtering requests based on source IP country using jpillora/ipfilter.
type IPFilterGeolocation struct {
	// A list of countries that the filter will allow.
	// If you specify this, you should not specify DenyCountries.
	// If both are specified, DenyCountries will take precedence.
	// All countries that are not in this list will be denied.
	// You can specify the special value "UNK" to match unrecognized countries.
	AllowCountries []string `json:"allow_countries,omitempty"`

	// A list of countries that the filter will deny.
	// If you specify this, you should not specify AllowCountries.
	// If both are specified, DenyCountries will take precedence.
	// All countries that are not in this list will be allowed.
	// You can specify the special value "UNK" to match unrecognized countries.
	DenyCountries []string `json:"deny_countries,omitempty"`

	// BlockByDefault sets the default behavior when no allow/deny rules match.
	// When true, requests are blocked by default unless explicitly allowed.
	// When false, requests are allowed by default unless explicitly denied.
	BlockByDefault bool `json:"block_by_default,omitempty"`

	filter *ipfilter.IPFilter
	logger *zap.Logger
}

// CaddyModule returns the Caddy module information.
func (IPFilterGeolocation) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.matchers.ipfilter_geolocation",
		New: func() caddy.Module { return new(IPFilterGeolocation) },
	}
}

// Provision sets up the module.
func (m *IPFilterGeolocation) Provision(ctx caddy.Context) error {
	m.logger = ctx.Logger(m)

	// Initialize ipfilter with configured options
	opts := ipfilter.Options{
		BlockByDefault: m.BlockByDefault,
	}

	// Add allowed countries
	if len(m.AllowCountries) > 0 {
		for _, country := range m.AllowCountries {
			opts.AllowedCountries = append(opts.AllowedCountries, country)
		}
	}

	// Add blocked countries
	if len(m.DenyCountries) > 0 {
		for _, country := range m.DenyCountries {
			opts.BlockedCountries = append(opts.BlockedCountries, country)
		}
	}

	m.filter = ipfilter.New(opts)
	return nil
}

// Validate validates the module configuration.
func (m *IPFilterGeolocation) Validate() error {
	// Ensure we don't have conflicting allow/deny configurations
	if len(m.AllowCountries) > 0 && len(m.DenyCountries) > 0 {
		return fmt.Errorf("cannot specify both allow_countries and deny_countries")
	}
	return nil
}

// Cleanup cleans up the module resources.
func (m *IPFilterGeolocation) Cleanup() error {
	// ipfilter doesn't require explicit cleanup
	return nil
}

// Match checks if the request matches the geolocation criteria.
func (m *IPFilterGeolocation) Match(r *http.Request) bool {
	// Get client IP from Caddy's context
	clientIP, ok := caddyhttp.GetVar(r.Context(), caddyhttp.ClientIPVarKey).(string)
	if !ok {
		m.logger.Warn("failed getting client IP from context")
		return false
	}

	// Check if IP is allowed by the filter
	allowed := m.filter.Allowed(clientIP)

	// Get country for logging
	country := m.filter.IPToCountry(clientIP)

	m.logger.Debug("IPFilter geolocation check",
		zap.String("ip", clientIP),
		zap.String("country", country),
		zap.Bool("allowed", allowed))

	return allowed
}

// UnmarshalCaddyfile parses the Caddyfile configuration.
/*
The matcher configuration will have a single block with the following parameters:

- `allow_countries`: a space-separated list of allowed countries
- `deny_countries`: a space-separated list of denied countries
- `block_by_default`: whether to block by default (true/false)

Examples:
	ipfilter_geolocation {
		allow_countries AU US CA
		block_by_default true
	}

	ipfilter_geolocation {
		deny_countries RU CN
	}
*/
func (m *IPFilterGeolocation) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	current := 0
	for d.Next() {
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "allow_countries":
				current = 1
			case "deny_countries":
				current = 2
			case "block_by_default":
				current = 3
			default:
				switch current {
				case 1:
					m.AllowCountries = append(m.AllowCountries, d.Val())
				case 2:
					m.DenyCountries = append(m.DenyCountries, d.Val())
				case 3:
					if d.Val() == "true" {
						m.BlockByDefault = true
					}
					current = 0
				default:
					return fmt.Errorf("unexpected config parameter %s", d.Val())
				}
			}
		}
	}
	return nil
}
