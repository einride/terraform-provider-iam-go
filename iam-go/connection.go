package iam_go

import (
	"context"
	"crypto/x509"
	"fmt"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func Connect(ctx context.Context, address string, token string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithUnaryInterceptor(unaryClientInterceptor))
	opts = append(opts, grpc.WithPerRPCCredentials(tokenCredentials(token)))

	const tlsPort = 443
	address = withDefaultPort(address, tlsPort)

	systemCertPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(systemCertPool, "")))

	return grpc.DialContext(ctx, address, opts...)
}

func unaryClientInterceptor(
	ctx context.Context,
	method string,
	req, reply interface{},
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
		return &printDetailsError{err: err}
	}
	return nil
}

type tokenCredentials string

func (t tokenCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"authorization": "Bearer " + string(t),
	}, nil
}

func (t tokenCredentials) RequireTransportSecurity() bool {
	return false
}

type printDetailsError struct {
	err error
}

func (e *printDetailsError) Error() string {
	s, ok := status.FromError(e.err)
	if !ok {
		return e.err.Error()
	}
	details := s.Details()
	if len(details) == 0 {
		return e.err.Error()
	}
	var result strings.Builder
	_, _ = result.WriteString(e.err.Error())
	for _, details := range details {
		_ = result.WriteByte('\n')
		if protoDetails, ok := details.(proto.Message); ok {
			_, _ = result.WriteString(protojson.Format(protoDetails))
		} else {
			_, _ = result.WriteString(fmt.Sprintf("%v", details))
		}
	}
	return result.String()
}

func withDefaultPort(target string, port int) string {
	parts := strings.Split(target, ":")
	if len(parts) == 1 {
		return target + ":" + strconv.Itoa(port)
	}
	return target
}
