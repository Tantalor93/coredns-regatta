package regatta

import (
	"crypto/tls"
	"crypto/x509"
	"path"

	"github.com/jamf/regatta/proto"
	"github.com/miekg/dns"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func createClient(endpoint string, insecure bool) (proto.KVClient, error) {
	pool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	connOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{RootCAs: pool, InsecureSkipVerify: insecure})),
	}

	conn, err := grpc.Dial(endpoint, connOpts...)
	if err != nil {
		return nil, err
	}

	return proto.NewKVClient(conn), nil
}

// Key converts a domain name to a key in Regatta.
func Key(s string) string {
	l := dns.SplitDomainName(s)
	for i, j := 0, len(l)-1; i < j; i, j = i+1, j-1 {
		l[i], l[j] = l[j], l[i]
	}
	return path.Join(append([]string{"/"}, l...)...)
}
