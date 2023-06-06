package regatta

import (
	"context"
	"encoding/json"
	"time"

	"github.com/coredns/coredns/plugin"
	"github.com/coredns/coredns/plugin/etcd/msg"
	"github.com/coredns/coredns/plugin/pkg/fall"
	"github.com/coredns/coredns/plugin/pkg/upstream"
	"github.com/coredns/coredns/request"
	"github.com/jamf/regatta/proto"
	"github.com/miekg/dns"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const pluginName = "regatta"

// Regatta is a plugin that serves zone data from Regatta data store.
type Regatta struct {
	Next  plugin.Handler
	Zones []string

	Upstream *upstream.Upstream
	Fall     fall.F

	client proto.KVClient
	table  string
}

func (r *Regatta) Services(ctx context.Context, state request.Request, exact bool, _ plugin.Options) ([]msg.Service, error) {
	return r.Records(ctx, state, exact)
}

func (r *Regatta) Reverse(ctx context.Context, state request.Request, exact bool, opt plugin.Options) ([]msg.Service, error) {
	return r.Services(ctx, state, exact, opt)
}

func (r *Regatta) Lookup(ctx context.Context, state request.Request, name string, typ uint16) (*dns.Msg, error) {
	return r.Upstream.Lookup(ctx, state, name, typ)
}

func (r *Regatta) Records(ctx context.Context, state request.Request, _ bool) ([]msg.Service, error) {
	name := state.Name()

	key := Key(name)
	rangeRequest := proto.RangeRequest{
		Table: []byte(r.table),
		Key:   []byte(key),
	}

	resp, err := r.client.Range(ctx, &rangeRequest)
	if err != nil {
		return nil, err
	}

	var svcs []msg.Service
	for _, v := range resp.Kvs {
		var entry msg.Service
		err := json.Unmarshal(v.Value, &entry)
		if err != nil {
			return svcs, err
		}
		svcs = append(svcs, entry)
	}
	return svcs, nil
}

func (r *Regatta) IsNameError(err error) bool {
	if st := status.Convert(err); st != nil {
		switch st.Code() {
		case codes.NotFound:
			return true
		}
	}
	return false
}

func (r *Regatta) Serial(_ request.Request) uint32 {
	return uint32(time.Now().Unix())
}

func (r *Regatta) MinTTL(_ request.Request) uint32 {
	return 30
}

func (r *Regatta) ServeDNS(ctx context.Context, w dns.ResponseWriter, m *dns.Msg) (int, error) {
	req := request.Request{W: w, Req: m}
	opt := plugin.Options{}

	zone := plugin.Zones(r.Zones).Matches(req.Name())
	if zone == "" {
		return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, m)
	}

	var (
		records, extra []dns.RR
		truncated      bool
		err            error
	)
	switch req.QType() {
	case dns.TypeA:
		records, truncated, err = plugin.A(ctx, r, zone, req, nil, opt)
	}

	if err != nil && r.IsNameError(err) {
		if r.Fall.Through(req.Name()) {
			return plugin.NextOrFailure(r.Name(), r.Next, ctx, w, m)
		}
		// Make err nil when returning here, so we don't log spam for NXDOMAIN.
		return plugin.BackendError(ctx, r, zone, dns.RcodeNameError, req, nil, opt)
	}

	if err != nil {
		return plugin.BackendError(ctx, r, zone, dns.RcodeServerFailure, req, err, opt)
	}

	resp := new(dns.Msg)
	resp.SetReply(m)
	resp.Truncated = truncated
	resp.Authoritative = true
	resp.Answer = append(m.Answer, records...)
	resp.Extra = append(m.Extra, extra...)

	w.WriteMsg(resp)
	return dns.RcodeSuccess, nil
}

func (r *Regatta) Name() string {
	return pluginName
}
