## Auth Service (bootstrap)

Features:
- Passwordless login via magic link token
- JWT issuance (HS256 by default; RS256 + JWKS plumbed)
- RBAC scopes embedded in access token
- REST endpoints: POST `/auth/login`, GET `/auth/verify`
- JWKS: GET `/.well-known/jwks.json` (returns empty when RSA not configured)

### Run

```bash
go run ./cmd/auth
```

Environment:
- `PORT` (default: 8080)

Health:
- GET `http://localhost:8080/healthz` → `ok`

### Test passwordless flow

1) Start login
```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"user@example.com"}'
# {"magic_token":"..."}
```

2) Verify token and receive JWTs
```bash
MAGIC=... # from step 1
curl -s "http://localhost:8080/auth/verify?token=$MAGIC"
# {"access_token":"...","refresh_token":"...","expires_in":3600}
```

3) Use Authorization: Bearer
```bash
ACCESS=... # from step 2
curl -s http://localhost:8080/healthz -H "Authorization: Bearer $ACCESS"
```

Notes:
- Current JWT is HS256 using a dev secret. RS256 + JWKS are wired; provide RSA keys via env and Gateway can verify via JWKS endpoint.

### JWKS

- GET `http://localhost:8080/.well-known/jwks.json`
- When RSA keys are configured in env, returns the RS256 public JWK.

### gRPC (optional, behind build tag)

To enable a minimal gRPC server without codegen (uses `structpb` as messages):

1) Install deps (if your Go proxy is flaky, try later or set `GOPROXY=direct`):
```bash
go get google.golang.org/grpc@v1.54.0
go get google.golang.org/protobuf@v1.31.0
```

2) Build/run with tag `grpc`:
```bash
go run -tags grpc ./cmd/auth
```

Server listens on `:9090` by default (env `GRPC_PORT`).

Service (logical contract):
```
service AuthService {
  rpc VerifyToken(TokenRequest) returns (TokenResponse);
  rpc RegisterUser(UserRequest) returns (UserResponse);
}

message TokenRequest   { string access_token = 1; }
message TokenResponse  { string user_id = 1; string email = 2; repeated string scopes = 3; }
message UserRequest    { string email = 1; }
message UserResponse   { string user_id = 1; }
```

grpcurl examples (plaintext for local):
```bash
# VerifyToken
grpcurl -plaintext -d '{"access_token":"'$ACCESS'"}' localhost:9090 auth.AuthService/VerifyToken

# RegisterUser
grpcurl -plaintext -d '{"email":"user@example.com"}' localhost:9090 auth.AuthService/RegisterUser
```

### RSA keys via env (optional)

- `JWT_PRIVATE_KEY_PEM` – RSA private key (PEM, PKCS1/PKCS8)
- `JWT_PUBLIC_KEY_PEM`  – RSA public key or certificate (PEM)
- `JWT_ISSUER`          – Issuer value in JWT (default: `templespace`)

# temple-space-backend
## Booking Service (core)

Functions:
- Create and manage bookings
- Availability check
- CQRS (write/read models)
- Integrations: Auth (token verify), Payment (charge), Kafka events

### Run

```bash
go run ./cmd/booking
```

Env (defaults in code):
- `BOOKING_HTTP_ADDR` (":8081")
- `BOOKING_GRPC_ADDR` (":9091")
- `BOOKING_POSTGRES_URL` (write model)
- `BOOKING_REDIS_URL` (read model cache)
- `BOOKING_KAFKA_BROKERS`
- `AUTH_GRPC_ADDR` (gRPC to Auth Service)

### REST endpoints (stubs)

- POST `/booking` – create booking
```json
{
  "space_id": "uuid-123",
  "user_id": "uuid-456",
  "slot_start": "2025-10-10T10:00:00Z",
  "slot_end": "2025-10-10T12:00:00Z"
}
```

- POST `/booking/{id}/pay` – confirm payment
- POST `/booking/{id}/cancel` – cancel booking

Flow:
1. Verify access token via Auth Service
2. Check availability (write repo / read cache)
3. Create/Update in write DB
4. Publish Kafka event: `booking_created` | `booking_paid` | `booking_cancelled`

### gRPC (internal)

Build with tag `grpc`:
```bash
go run -tags grpc ./cmd/booking
```

Service shape (no codegen, uses `structpb`):
```
service BookingService {
  rpc CreateBooking(CreateBookingRequest) returns (BookingResponse);
  rpc ConfirmPayment(PaymentConfirmation) returns (BookingResponse);
  rpc CancelBooking(CancelBookingRequest) returns (BookingResponse);
}
```

### Storage models

Write (Postgres):
```
CREATE TABLE bookings (
    id UUID PRIMARY KEY,
    space_id UUID NOT NULL,
    user_id UUID NOT NULL,
    slot_start TIMESTAMPTZ NOT NULL,
    slot_end TIMESTAMPTZ NOT NULL,
    status TEXT CHECK (status IN ('pending','confirmed','paid','cancelled')),
    version INT DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
```

Read:
- Redis – cache available slots
- Elasticsearch – fast search on spaces/slots

## Space Service (catalog)

Functions:
- CRUD for spaces and photos
- Search and filtering by name, location, tags, capacity, price
- Read model for fast search (in-memory now; swap to Elasticsearch later)
- Events: `space_created`, `space_updated`

### Run

```bash
go run ./cmd/space
```

Defaults (dev):
- HTTP `:8082`

Health:
- GET `http://localhost:8082/healthz` → `ok`

### REST endpoints

- POST `/spaces`
```json
{
  "name": "One Dance Studio",
  "location": "Belgrade",
  "tags": ["dance", "studio", "morning-light"],
  "attributes": {"size": 50, "capacity": 20},
  "price_per_hour": 15.0
}
```

- GET `/spaces?tags=dance&min_capacity=10&name=&location=&min_price=&max_price=`
  - served from read model (in-memory; replaceable with Elasticsearch)

- PUT `/spaces/{id}`
```json
{
  "name": "One Dance Studio",
  "location": "Belgrade",
  "tags": ["dance"],
  "attributes": {"capacity": 25},
  "price_per_hour": 18.0
}
```

Auth:
- Endpoints expect `Authorization` header; a stub verifier is used in dev.

Events:
- In-memory publisher logs events; swap to Kafka in production.

### gRPC (internal, optional)

Build with tag `grpc`:
```bash
go run -tags grpc ./cmd/space
```

Service shape (logical contract, no codegen):
```
service SpaceService {
  rpc GetSpace(GetSpaceRequest) returns (SpaceResponse);
  rpc ListSpaces(ListSpacesRequest) returns (ListSpacesResponse);
  rpc CreateSpace(CreateSpaceRequest) returns (SpaceResponse);
  rpc UpdateSpace(UpdateSpaceRequest) returns (SpaceResponse);
}
```
