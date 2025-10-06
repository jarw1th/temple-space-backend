package server

// GRPCPlaceholder exists to keep main wiring compiling when grpc build tag is not enabled.
func GRPCPlaceholder() bool { return true }
