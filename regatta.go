package regatta

import (
	"crypto/tls"
	"crypto/x509"

	"github.com/jamf/regatta/proto"
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
