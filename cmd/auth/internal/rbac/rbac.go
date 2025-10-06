package rbac

// Role represents a coarse role mapping to a set of scopes.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// Scopes are fine-grained permissions checked by API Gateway/services.
var RoleScopes = map[Role][]string{
	RoleUser:  {"profile:read", "booking:create", "booking:read"},
	RoleAdmin: {"profile:read", "profile:write", "booking:*", "space:*", "admin:*"},
}

func ScopesFor(role Role) []string { return RoleScopes[role] }
