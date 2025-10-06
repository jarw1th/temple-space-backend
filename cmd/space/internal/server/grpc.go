//go:build grpc

package server

import (
	"context"
	"log"
	"net"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"templespace/cmd/space/internal/domain"
	"templespace/cmd/space/internal/service"
)

type Server struct {
	g   *gogrpc.Server
	svc *service.Service
}

func NewServer(svc *service.Service) *Server {
	srv := &Server{g: gogrpc.NewServer(), svc: svc}
	srv.register()
	return srv
}

func (s *Server) register() {
	desc := &gogrpc.ServiceDesc{
		ServiceName: "space.SpaceService",
		HandlerType: (*interface{})(nil),
		Methods: []gogrpc.MethodDesc{
			{MethodName: "GetSpace", Handler: s.handleGetSpace},
			{MethodName: "ListSpaces", Handler: s.handleListSpaces},
			{MethodName: "CreateSpace", Handler: s.handleCreateSpace},
			{MethodName: "UpdateSpace", Handler: s.handleUpdateSpace},
		},
		Streams:  []gogrpc.StreamDesc{},
		Metadata: "space.proto",
	}
	s.g.RegisterService(desc, s)
}

func (s *Server) handleGetSpace(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.getSpace(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/space.SpaceService/GetSpace"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.getSpace(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) handleListSpaces(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.listSpaces(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/space.SpaceService/ListSpaces"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.listSpaces(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) handleCreateSpace(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.createSpace(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/space.SpaceService/CreateSpace"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.createSpace(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) handleUpdateSpace(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.updateSpace(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/space.SpaceService/UpdateSpace"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.updateSpace(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) getSpace(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	idv, _ := in.Fields["id"]
	if idv == nil {
		return structpb.NewStruct(map[string]any{"error": "missing id"})
	}
	sp, err := s.svc.GetSpace(ctx, idv.GetStringValue())
	if err != nil {
		return structpb.NewStruct(map[string]any{"error": err.Error()})
	}
	return structpb.NewStruct(spaceToMap(sp))
}

func (s *Server) listSpaces(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	q := domain.Query{}
	if f := in.GetFields()["name"]; f != nil {
		q.Name = f.GetStringValue()
	}
	if f := in.GetFields()["location"]; f != nil {
		q.Location = f.GetStringValue()
	}
	res, err := s.svc.ListSpaces(ctx, q)
	if err != nil {
		return structpb.NewStruct(map[string]any{"error": err.Error()})
	}
	list := make([]any, 0, len(res))
	for _, sp := range res {
		list = append(list, spaceToMap(sp))
	}
	return structpb.NewStruct(map[string]any{"spaces": list})
}

func (s *Server) createSpace(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	access := ""
	if f := in.GetFields()["access_token"]; f != nil {
		access = f.GetStringValue()
	}
	sp := &domain.Space{
		Name:         getString(in, "name"),
		Location:     getString(in, "location"),
		PricePerHour: getFloat(in, "price_per_hour"),
	}
	out, err := s.svc.CreateSpace(ctx, access, sp)
	if err != nil {
		return structpb.NewStruct(map[string]any{"error": err.Error()})
	}
	return structpb.NewStruct(spaceToMap(out))
}

func (s *Server) updateSpace(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	access := ""
	if f := in.GetFields()["access_token"]; f != nil {
		access = f.GetStringValue()
	}
	sp := &domain.Space{
		ID:           getString(in, "id"),
		Name:         getString(in, "name"),
		Location:     getString(in, "location"),
		PricePerHour: getFloat(in, "price_per_hour"),
	}
	out, err := s.svc.UpdateSpace(ctx, access, sp)
	if err != nil {
		return structpb.NewStruct(map[string]any{"error": err.Error()})
	}
	return structpb.NewStruct(spaceToMap(out))
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("grpc listening on %s", addr)
	return s.g.Serve(ln)
}

func getString(in *structpb.Struct, key string) string {
	if f := in.GetFields()[key]; f != nil {
		return f.GetStringValue()
	}
	return ""
}

func getFloat(in *structpb.Struct, key string) float64 {
	if f := in.GetFields()[key]; f != nil {
		return f.GetNumberValue()
	}
	return 0
}

func spaceToMap(s *domain.Space) map[string]any {
	return map[string]any{
		"id":             s.ID,
		"name":           s.Name,
		"location":       s.Location,
		"tags":           s.Tags,
		"attributes":     s.Attributes,
		"price_per_hour": s.PricePerHour,
		"created_at":     s.CreatedAt.String(),
		"updated_at":     s.UpdatedAt.String(),
	}
}
