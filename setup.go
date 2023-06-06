package regatta

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
	"github.com/coredns/coredns/plugin/pkg/upstream"
)

var log = clog.NewWithPlugin(pluginName)

func init() { plugin.Register(pluginName, setup) }

func setup(c *caddy.Controller) error {
	r := Regatta{}

	var endpoint string
	var insecure bool
	if c.Next() {
		r.Zones = plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
		for c.NextBlock() {
			switch c.Val() {
			case "fallthrough":
				r.Fall.SetZonesFromArgs(c.RemainingArgs())
			case "endpoint":
				if !c.NextArg() {
					return c.ArgErr()
				}
				endpoint = c.Val()
			case "insecure":
				insecure = true
			case "table":
				if !c.NextArg() {
					return c.ArgErr()
				}
				r.table = c.Val()
			}
		}
	}

	if len(r.table) == 0 {
		return c.Err("missing Regatta table configuration")
	}

	client, err := createClient(endpoint, insecure)
	if err != nil {
		return c.Errf("failed to create Regatta client due to '%v'", err)
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		r.client = client
		r.Upstream = upstream.New()
		return &r
	})

	return nil
}
