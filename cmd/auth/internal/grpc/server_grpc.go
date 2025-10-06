//go:build grpc

package grpc

import (
	"context"
	"log"
	"net"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"templespace/cmd/auth/internal/auth"
)

type Server struct {
	g      *gogrpc.Server
	auth   *auth.Service
	signer *auth.JWTSigner
}

func NewServer(a *auth.Service, s *auth.JWTSigner) *Server {
	srv := &Server{g: gogrpc.NewServer(), auth: a, signer: s}
	srv.register()
	return srv
}

func (s *Server) register() {
	desc := &gogrpc.ServiceDesc{
		ServiceName: "auth.AuthService",
		HandlerType: (*interface{})(nil),
		Methods: []gogrpc.MethodDesc{
			{MethodName: "VerifyToken", Handler: s.handleVerifyToken},
			{MethodName: "RegisterUser", Handler: s.handleRegisterUser},
		},
		Streams:  []gogrpc.StreamDesc{},
		Metadata: "auth.proto",
	}
	s.g.RegisterService(desc, s)
}

func (s *Server) handleVerifyToken(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.verifyToken(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/auth.AuthService/VerifyToken"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.verifyToken(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) handleRegisterUser(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.registerUser(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/auth.AuthService/RegisterUser"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.registerUser(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) verifyToken(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	tok, _ := in.Fields["access_token"]
	if tok == nil {
		return structpb.NewStruct(map[string]interface{}{"error": "missing access_token"})
	}
	claims, err := s.signer.ParseAndVerify(tok.GetStringValue())
	if err != nil {
		return structpb.NewStruct(map[string]interface{}{"error": err.Error()})
	}
	return structpb.NewStruct(map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"scopes":  claims.Scopes,
	})
}

func (s *Server) registerUser(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	emailv, _ := in.Fields["email"]
	if emailv == nil {
		return structpb.NewStruct(map[string]interface{}{"error": "missing email"})
	}
	email := emailv.GetStringValue()
	// noop: user repo is attached to auth service; Upsert happens on magic verify. For demo, return ID=email.
	return structpb.NewStruct(map[string]interface{}{"user_id": email})
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("grpc listening on %s", addr)
	return s.g.Serve(ln)
}
