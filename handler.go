package regatta

import (
	"context"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/request"
	"github.com/jamf/regatta/proto"
	"github.com/miekg/dns"
)

const pluginName = "regatta"

// Regatta is a plugin that serves zone data from Regatta data store.
type Regatta struct {
	Next  plugin.Handler
	Zones []string

	client proto.KVClient
	table  string
}

func (r Regatta) ServeDNS(ctx context.Context, w dns.ResponseWriter, m *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: m}

	zone := plugin.Zones(r.Zones).Matches(req.Name())
	if zone == "" {
		log.Info("Regatta plugin not matched request")
		return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, m)
	}
	log.Info("Regatta plugin matched request")
	return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, m)
}

func (r Regatta) Name() string {
	return pluginName
}
