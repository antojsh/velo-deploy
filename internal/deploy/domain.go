package deploy

import (
	"fmt"
	"os"

	"velo-deploy/internal/config"
)

// DomainMode identifies how a deployed app's public domain was resolved.
type DomainMode string

const (
	// ModeCustom: --domain was provided explicitly. Uses Let's Encrypt HTTP-01.
	ModeCustom DomainMode = "custom"
	// ModeDuckDNS: app will be served at <app>.<sub>.duckdns.org with
	// a Let's Encrypt cert obtained via DNS-01 (requires caddy-dns/duckdns).
	ModeDuckDNS DomainMode = "duckdns"
	// ModeSelfSigned: no public domain. App is served via the shared
	// path-based config with a Caddy self-signed cert.
	ModeSelfSigned DomainMode = "self-signed"
)

// ResolvedDomain is the outcome of deciding how an app will be reached.
type ResolvedDomain struct {
	// Domain is the public domain for the app, or "" when self-signed.
	Domain string
	// Alias is the local /etc/hosts alias used as the Caddy upstream.
	// Always set, regardless of mode.
	Alias string
	// Mode indicates which routing strategy applies.
	Mode DomainMode
}

// resolveDomain decides which domain (if any) an app should use, applying
// the documented priority:
//
//  1. explicitDomain (from --domain)           → ModeCustom
//  2. useDuckDNS && cfg.DuckDNSSubdomain+Token → ModeDuckDNS
//  3. none of the above                        → ModeSelfSigned
//
// The local alias (`<appName>.local`) is always returned, since Caddy
// uses it as the upstream for reverse-proxying requests to the systemd
// service, regardless of whether the public traffic is domain-based or
// path-based.
func resolveDomain(cfg *config.Config, appName, explicitDomain string, useDuckDNS bool) ResolvedDomain {
	alias := appName + ".local"

	if explicitDomain != "" {
		return ResolvedDomain{
			Domain: explicitDomain,
			Alias:  alias,
			Mode:   ModeCustom,
		}
	}

	if useDuckDNS && cfg.DuckDNSSubdomain != "" && cfg.DuckDNSToken != "" {
		return ResolvedDomain{
			Domain: fmt.Sprintf("%s.%s.duckdns.org", appName, cfg.DuckDNSSubdomain),
			Alias:  alias,
			Mode:   ModeDuckDNS,
		}
	}

	return ResolvedDomain{
		Domain: "",
		Alias:  alias,
		Mode:   ModeSelfSigned,
	}
}

// shouldWriteSharedFallback returns true when the app must NOT have a
// dedicated Caddy vhost (i.e. no public domain), so it should be served
// from the shared catch-all config and use Caddy's path-based routing.
func (r ResolvedDomain) shouldWriteSharedFallback() bool {
	return r.Mode == ModeSelfSigned
}

// requiresDuckDNSPlugin reports whether the resolved domain depends on
// the caddy-dns/duckdns module being present in the Caddy binary.
func (r ResolvedDomain) requiresDuckDNSPlugin() bool {
	return r.Mode == ModeDuckDNS
}

// warnUnreachable prints a friendly notice when a custom domain is set
// but the user has not provided a way to make it reachable from the public
// internet (no DNS-01 provider). The function is a no-op in production
// paths where the domain is already correctly resolvable.
func (r ResolvedDomain) warnUnreachable() {
	if r.Mode != ModeCustom {
		return
	}
	if os.Getenv("DEPLOY_SKIP_DOMAIN_HINT") == "1" {
		return
	}
	fmt.Printf("  → Make sure %s points to this server (A record).\n", r.Domain)
}
