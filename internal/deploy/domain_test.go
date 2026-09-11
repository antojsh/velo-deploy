package deploy

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"velo-deploy/internal/config"
)

func TestResolveDomain_Custom(t *testing.T) {
	cfg := &config.Config{
		DuckDNSSubdomain: "pepito",
		DuckDNSToken:     "some-token",
	}

	r := resolveDomain(cfg, "myapi", "api.example.com", false)

	assert.Equal(t, "api.example.com", r.Domain)
	assert.Equal(t, "myapi.local", r.Alias)
	assert.Equal(t, ModeCustom, r.Mode)
}

func TestResolveDomain_CustomBeatsDuckDNS(t *testing.T) {
	// Even when DuckDNS is configured, --domain must take priority.
	cfg := &config.Config{
		DuckDNSSubdomain: "pepito",
		DuckDNSToken:     "some-token",
	}

	r := resolveDomain(cfg, "myapi", "api.example.com", true)

	assert.Equal(t, ModeCustom, r.Mode)
	assert.Equal(t, "api.example.com", r.Domain)
}

func TestResolveDomain_DuckDNS(t *testing.T) {
	cfg := &config.Config{
		DuckDNSSubdomain: "pepito",
		DuckDNSToken:     "tok",
	}

	r := resolveDomain(cfg, "myapi", "", true)

	assert.Equal(t, "myapi.pepito.duckdns.org", r.Domain)
	assert.Equal(t, "myapi.local", r.Alias)
	assert.Equal(t, ModeDuckDNS, r.Mode)
	assert.True(t, r.requiresDuckDNSPlugin())
}

func TestResolveDomain_DuckDNSWithoutFlag(t *testing.T) {
	// DuckDNS configured but user didn't pass --duckdns → fallback to self-signed.
	cfg := &config.Config{
		DuckDNSSubdomain: "pepito",
		DuckDNSToken:     "tok",
	}

	r := resolveDomain(cfg, "myapi", "", false)

	assert.Equal(t, ModeSelfSigned, r.Mode)
	assert.Equal(t, "", r.Domain)
	assert.Equal(t, "myapi.local", r.Alias)
}

func TestResolveDomain_DuckDNSIncomplete(t *testing.T) {
	// Token missing → cannot use DuckDNS even if flag is true.
	cfg := &config.Config{
		DuckDNSSubdomain: "pepito",
		DuckDNSToken:     "",
	}

	r := resolveDomain(cfg, "myapi", "", true)

	assert.Equal(t, ModeSelfSigned, r.Mode)
}

func TestResolveDomain_SelfSigned(t *testing.T) {
	cfg := &config.Config{}

	r := resolveDomain(cfg, "myapi", "", false)

	assert.Equal(t, "", r.Domain)
	assert.Equal(t, "myapi.local", r.Alias)
	assert.Equal(t, ModeSelfSigned, r.Mode)
	assert.True(t, r.shouldWriteSharedFallback())
	assert.False(t, r.requiresDuckDNSPlugin())
}

func TestResolvedDomain_ShouldWriteSharedFallback(t *testing.T) {
	cases := []struct {
		mode DomainMode
		want bool
	}{
		{ModeSelfSigned, true},
		{ModeCustom, false},
		{ModeDuckDNS, false},
	}
	for _, c := range cases {
		r := ResolvedDomain{Mode: c.mode}
		assert.Equal(t, c.want, r.shouldWriteSharedFallback(), "mode=%s", c.mode)
	}
}

func TestResolvedDomain_RequiresDuckDNSPlugin(t *testing.T) {
	cases := []struct {
		mode DomainMode
		want bool
	}{
		{ModeDuckDNS, true},
		{ModeCustom, false},
		{ModeSelfSigned, false},
	}
	for _, c := range cases {
		r := ResolvedDomain{Mode: c.mode}
		assert.Equal(t, c.want, r.requiresDuckDNSPlugin(), "mode=%s", c.mode)
	}
}
