package synthient

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

// DefaultGRPCEndpoint is the default Synthient gRPC server address.
const DefaultGRPCEndpoint = "grpc.synthient.com:443"

// GRPCSchemaOptions configures a GRPCSchema call.
type GRPCSchemaOptions struct {
	// Endpoint is the gRPC server address (host:port or grpc://host:port).
	// Defaults to DefaultGRPCEndpoint when empty.
	Endpoint string
	// Plaintext disables TLS. Useful for local development against a plaintext server.
	Plaintext bool
	// IncludeReflection includes the gRPC reflection service itself in the output.
	IncludeReflection bool
	// Symbols is the list of fully-qualified service or message names to resolve.
	// When empty, all services exposed by the server are fetched.
	Symbols []string
}

// GRPCSchemaResult holds the output of a GRPCSchema call.
type GRPCSchemaResult struct {
	// Endpoint is the normalized gRPC address that was queried.
	Endpoint string
	// Symbols lists the service symbols that were resolved.
	Symbols []string
	// DescriptorSet contains all resolved protobuf file descriptors in topological
	// dependency order.
	DescriptorSet *descriptorpb.FileDescriptorSet
}

// GRPCSchema uses gRPC server reflection to fetch protobuf file descriptors from a
// Synthient gRPC endpoint. If options is nil or options.Endpoint is empty, it connects
// to DefaultGRPCEndpoint. If options.Symbols is empty, all services exposed by the
// server are fetched.
//
// The client's token is forwarded as the x-api-key metadata header when non-empty.
//
// Example (all services):
//
//	result, err := client.GRPCSchema(ctx, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for _, f := range result.DescriptorSet.File {
//		fmt.Printf("%s  %s\n", f.GetName(), f.GetPackage())
//	}
//
// Example (specific symbol):
//
//	result, err := client.GRPCSchema(ctx, &synthient.GRPCSchemaOptions{
//		Symbols: []string{"synthient.lookup.v1.LookupService"},
//	})
func (client *Client) GRPCSchema(ctx context.Context, options *GRPCSchemaOptions) (GRPCSchemaResult, error) {
	if options == nil {
		options = &GRPCSchemaOptions{}
	}

	endpoint, host, err := NormalizeGRPCEndpoint(options.Endpoint)
	if err != nil {
		return GRPCSchemaResult{}, err
	}

	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
			ServerName: host,
			MinVersion: tls.VersionTLS12,
		})),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(32 << 20)),
	}
	if options.Plaintext {
		dialOptions[0] = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.NewClient(endpoint, dialOptions...)
	if err != nil {
		return GRPCSchemaResult{}, fmt.Errorf("creating grpc client: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if strings.TrimSpace(client.Token) != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", client.Token)
	}

	rc := reflectpb.NewServerReflectionClient(conn)
	symbols := grpcCleanSymbols(options.Symbols)
	if len(symbols) == 0 {
		symbols, err = grpcListServices(ctx, rc, host, options.IncludeReflection)
		if err != nil {
			return GRPCSchemaResult{}, err
		}
	}
	if len(symbols) == 0 {
		return GRPCSchemaResult{}, errors.New("no protobuf services returned by reflection")
	}

	files := map[string]*descriptorpb.FileDescriptorProto{}
	for _, symbol := range symbols {
		rawFiles, err := grpcFileContainingSymbol(ctx, rc, host, symbol)
		if err != nil {
			return GRPCSchemaResult{}, err
		}
		err = grpcAddFiles(files, rawFiles)
		if err != nil {
			return GRPCSchemaResult{}, err
		}
	}
	err = grpcResolveDependencies(ctx, rc, host, files)
	if err != nil {
		return GRPCSchemaResult{}, err
	}

	return GRPCSchemaResult{
		Endpoint:      endpoint,
		Symbols:       symbols,
		DescriptorSet: &descriptorpb.FileDescriptorSet{File: grpcOrderedFiles(files)},
	}, nil
}

// NormalizeGRPCEndpoint normalizes a raw gRPC endpoint string to host:port form.
// Accepts grpc:// or https:// URL schemes, bare hostnames (port 443 assumed), and
// host:port strings. If raw is empty, DefaultGRPCEndpoint is used.
func NormalizeGRPCEndpoint(raw string) (endpoint string, host string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = DefaultGRPCEndpoint
	}
	if strings.Contains(raw, "://") {
		parsed, parseErr := url.Parse(raw)
		if parseErr != nil {
			return "", "", fmt.Errorf("parsing grpc endpoint: %w", parseErr)
		}
		raw = parsed.Host
		if raw == "" {
			raw = strings.Trim(parsed.Path, "/")
		}
	}
	raw = strings.TrimSuffix(raw, "/")
	if raw == "" {
		return "", "", errors.New("empty grpc endpoint")
	}
	h, _, splitErr := net.SplitHostPort(raw)
	if splitErr == nil {
		return raw, strings.Trim(h, "[]"), nil
	}
	if !strings.Contains(raw, ":") {
		return net.JoinHostPort(raw, "443"), raw, nil
	}
	return raw, strings.Trim(raw, "[]"), nil
}

// ExplainGRPCError returns a human-readable explanation for common gRPC status errors
// returned by GRPCSchema. Returns an empty string for unrecognized errors.
func ExplainGRPCError(err error) string {
	s, ok := status.FromError(err)
	if !ok {
		return ""
	}
	switch s.Code() {
	case codes.Unauthenticated:
		return "Missing or invalid API key."
	case codes.PermissionDenied:
		return "This API key does not have permission to inspect gRPC schemas."
	case codes.NotFound:
		return "No protobuf schema found for the requested symbol. Call GRPCSchema without symbols to list available services."
	case codes.Unimplemented:
		return "The gRPC endpoint does not expose server reflection."
	case codes.Unavailable:
		return "Could not reach the gRPC endpoint. Check the address, network access, and TLS settings."
	case codes.DeadlineExceeded:
		return "Timed out reading protobuf schemas from gRPC reflection. Use a longer context deadline."
	default:
		return ""
	}
}

func grpcListServices(ctx context.Context, rc reflectpb.ServerReflectionClient, host string, includeReflection bool) ([]string, error) {
	resp, err := grpcReflectionRequest(ctx, rc, host, &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{ListServices: ""},
	})
	if err != nil {
		return nil, err
	}
	list := resp.GetListServicesResponse()
	if list == nil {
		return nil, errors.New("reflection response did not include service list")
	}
	services := []string{}
	for _, svc := range list.Service {
		name := strings.TrimSpace(svc.GetName())
		if name == "" {
			continue
		}
		if !includeReflection && strings.HasPrefix(name, "grpc.reflection.") {
			continue
		}
		services = append(services, name)
	}
	sort.Strings(services)
	return services, nil
}

func grpcFileContainingSymbol(ctx context.Context, rc reflectpb.ServerReflectionClient, host string, symbol string) ([][]byte, error) {
	resp, err := grpcReflectionRequest(ctx, rc, host, &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileContainingSymbol{FileContainingSymbol: symbol},
	})
	if err != nil {
		return nil, err
	}
	dr := resp.GetFileDescriptorResponse()
	if dr == nil {
		return nil, fmt.Errorf("reflection response did not include descriptors for %s", symbol)
	}
	return dr.FileDescriptorProto, nil
}

func grpcFileByName(ctx context.Context, rc reflectpb.ServerReflectionClient, host string, name string) ([][]byte, error) {
	resp, err := grpcReflectionRequest(ctx, rc, host, &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_FileByFilename{FileByFilename: name},
	})
	if err != nil {
		return nil, err
	}
	dr := resp.GetFileDescriptorResponse()
	if dr == nil {
		return nil, fmt.Errorf("reflection response did not include descriptors for %s", name)
	}
	return dr.FileDescriptorProto, nil
}

func grpcReflectionRequest(ctx context.Context, rc reflectpb.ServerReflectionClient, host string, req *reflectpb.ServerReflectionRequest) (*reflectpb.ServerReflectionResponse, error) {
	stream, err := rc.ServerReflectionInfo(ctx)
	if err != nil {
		return nil, err
	}
	req.Host = host
	err = stream.Send(req)
	if err != nil {
		return nil, err
	}
	resp, err := stream.Recv()
	_ = stream.CloseSend()
	if err != nil {
		return nil, err
	}
	if errResp := resp.GetErrorResponse(); errResp != nil {
		return nil, status.Error(codes.Code(errResp.ErrorCode), errResp.ErrorMessage)
	}
	return resp, nil
}

func grpcAddFiles(files map[string]*descriptorpb.FileDescriptorProto, rawFiles [][]byte) error {
	for _, raw := range rawFiles {
		var file descriptorpb.FileDescriptorProto
		err := proto.Unmarshal(raw, &file)
		if err != nil {
			return fmt.Errorf("decoding file descriptor: %w", err)
		}
		if file.GetName() == "" {
			return errors.New("reflection returned a file descriptor without a name")
		}
		if _, ok := files[file.GetName()]; !ok {
			files[file.GetName()] = &file
		}
	}
	return nil
}

func grpcResolveDependencies(ctx context.Context, rc reflectpb.ServerReflectionClient, host string, files map[string]*descriptorpb.FileDescriptorProto) error {
	for {
		missing := grpcMissingDependencies(files)
		if len(missing) == 0 {
			return nil
		}
		for _, name := range missing {
			rawFiles, err := grpcFileByName(ctx, rc, host, name)
			if err != nil {
				return err
			}
			err = grpcAddFiles(files, rawFiles)
			if err != nil {
				return err
			}
			if _, ok := files[name]; !ok {
				return fmt.Errorf("reflection did not return missing dependency %s", name)
			}
		}
	}
}

func grpcMissingDependencies(files map[string]*descriptorpb.FileDescriptorProto) []string {
	seen := map[string]bool{}
	missing := []string{}
	for _, file := range files {
		for _, dep := range file.Dependency {
			if _, ok := files[dep]; ok || seen[dep] {
				continue
			}
			seen[dep] = true
			missing = append(missing, dep)
		}
	}
	sort.Strings(missing)
	return missing
}

func grpcOrderedFiles(files map[string]*descriptorpb.FileDescriptorProto) []*descriptorpb.FileDescriptorProto {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)

	ordered := []*descriptorpb.FileDescriptorProto{}
	visited := map[string]bool{}
	visiting := map[string]bool{}
	var visit func(string)
	visit = func(name string) {
		if visited[name] || visiting[name] {
			return
		}
		file, ok := files[name]
		if !ok {
			return
		}
		visiting[name] = true
		deps := append([]string{}, file.Dependency...)
		sort.Strings(deps)
		for _, dep := range deps {
			visit(dep)
		}
		visiting[name] = false
		visited[name] = true
		ordered = append(ordered, file)
	}
	for _, name := range names {
		visit(name)
	}
	return ordered
}

func grpcCleanSymbols(symbols []string) []string {
	seen := map[string]bool{}
	cleaned := []string{}
	for _, s := range symbols {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		cleaned = append(cleaned, s)
	}
	return cleaned
}
