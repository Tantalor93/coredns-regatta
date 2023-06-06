package regatta

import (
	"github.com/coredns/caddy"
	"github.com/coredns/coredns/core/dnsserver"
	"github.com/coredns/coredns/plugin"
	clog "github.com/coredns/coredns/plugin/pkg/log"
)

var log = clog.NewWithPlugin(pluginName)

func init() { plugin.Register(pluginName, setup) }

func setup(c *caddy.Controller) error {
	r := Regatta{}
	if c.Next() {
		r.Zones = plugin.OriginsFromArgsOrServerBlock(c.RemainingArgs(), c.ServerBlockKeys)
		for c.NextBlock() {
		}
	}

	dnsserver.GetConfig(c).AddPlugin(func(next plugin.Handler) plugin.Handler {
		r.Next = next
		return r
	})

	log.Info("Regatta plugin registered")

	return nil
}
