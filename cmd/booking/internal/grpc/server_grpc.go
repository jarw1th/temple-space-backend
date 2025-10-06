//go:build grpc

package grpc

import (
	"context"
	"log"
	"net"
	"time"

	gogrpc "google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"

	"templespace/cmd/booking/internal/service"
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
		ServiceName: "booking.BookingService",
		HandlerType: (*interface{})(nil),
		Methods: []gogrpc.MethodDesc{
			{MethodName: "CreateBooking", Handler: s.handleCreateBooking},
			{MethodName: "ConfirmPayment", Handler: s.handleConfirmPayment},
			{MethodName: "CancelBooking", Handler: s.handleCancelBooking},
		},
		Streams:  []gogrpc.StreamDesc{},
		Metadata: "booking.proto",
	}
	s.g.RegisterService(desc, s)
}

func (s *Server) handleCreateBooking(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.createBooking(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/booking.BookingService/CreateBooking"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.createBooking(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) handleConfirmPayment(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.confirmPayment(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/booking.BookingService/ConfirmPayment"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.confirmPayment(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) handleCancelBooking(_ interface{}, ctx context.Context, dec func(interface{}) error, interceptor gogrpc.UnaryServerInterceptor) (interface{}, error) {
	in := &structpb.Struct{}
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return s.cancelBooking(ctx, in)
	}
	info := &gogrpc.UnaryServerInfo{Server: s, FullMethod: "/booking.BookingService/CancelBooking"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.cancelBooking(ctx, req.(*structpb.Struct))
	}
	return interceptor(ctx, in, info, handler)
}

func (s *Server) createBooking(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	tok := in.GetFields()["access_token"].GetStringValue()
	spaceID := in.GetFields()["space_id"].GetStringValue()
	userID := in.GetFields()["user_id"].GetStringValue()
	startStr := in.GetFields()["slot_start"].GetStringValue()
	endStr := in.GetFields()["slot_end"].GetStringValue()
	start, _ := time.Parse(time.RFC3339, startStr)
	end, _ := time.Parse(time.RFC3339, endStr)
	b, err := s.svc.CreateBooking(ctx, tok, spaceID, userID, start, end)
	if err != nil {
		return structpb.NewStruct(map[string]interface{}{"error": err.Error()})
	}
	return structpb.NewStruct(map[string]interface{}{"id": b.ID, "status": string(b.Status)})
}

func (s *Server) confirmPayment(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	tok := in.GetFields()["access_token"].GetStringValue()
	id := in.GetFields()["booking_id"].GetStringValue()
	b, err := s.svc.ConfirmPayment(ctx, tok, id)
	if err != nil {
		return structpb.NewStruct(map[string]interface{}{"error": err.Error()})
	}
	return structpb.NewStruct(map[string]interface{}{"id": b.ID, "status": string(b.Status)})
}

func (s *Server) cancelBooking(ctx context.Context, in *structpb.Struct) (*structpb.Struct, error) {
	tok := in.GetFields()["access_token"].GetStringValue()
	id := in.GetFields()["booking_id"].GetStringValue()
	b, err := s.svc.CancelBooking(ctx, tok, id)
	if err != nil {
		return structpb.NewStruct(map[string]interface{}{"error": err.Error()})
	}
	return structpb.NewStruct(map[string]interface{}{"id": b.ID, "status": string(b.Status)})
}

func (s *Server) Listen(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("grpc listening on %s", addr)
	return s.g.Serve(ln)
}
