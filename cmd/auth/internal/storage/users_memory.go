package storage

type User struct {
	ID    string
	Email string
	Role  string
}

type UserRepo interface {
	FindByEmail(email string) (*User, bool)
	Upsert(u User) error
}

type InMemoryUsers struct { InMemoryKV }

func NewInMemoryUsers() *InMemoryUsers { return &InMemoryUsers{InMemoryKV: *NewInMemoryKV()} }

func (s *InMemoryUsers) FindByEmail(email string) (*User, bool) {
	if v, ok := s.get("email:"+email); ok {
		return &User{ID: v, Email: email, Role: "user"}, true
	}
	return nil, false
}

func (s *InMemoryUsers) Upsert(u User) error {
	if u.ID == "" { u.ID = u.Email }
	s.set("email:"+u.Email, u.ID, 0)
	return nil
}
