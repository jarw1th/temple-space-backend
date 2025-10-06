package auth

import (
	"crypto"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"strings"
	"time"
)

type JWTSigner struct {
	Issuer     string
	Secret     []byte // HS256 secret (fallback when RSA not configured)
	RSAPrivate *rsa.PrivateKey
	RSAPublic  *rsa.PublicKey
	KeyID      string
}

type Claims struct {
	Issuer   string   `json:"iss"`
	Subject  string   `json:"sub"`
	Audience string   `json:"aud,omitempty"`
	IssuedAt int64    `json:"iat"`
	Expires  int64    `json:"exp"`
	Scopes   []string `json:"scopes,omitempty"`
	UserID   string   `json:"uid,omitempty"`
	Email    string   `json:"email,omitempty"`
}

func base64urlEncode(b []byte) string          { return base64.RawURLEncoding.EncodeToString(b) }
func base64urlDecode(s string) ([]byte, error) { return base64.RawURLEncoding.DecodeString(s) }

func (j *JWTSigner) sign(data string) string {
	h := hmac.New(sha256.New, j.Secret)
	h.Write([]byte(data))
	return base64urlEncode(h.Sum(nil))
}

func (j *JWTSigner) encodeHeader() string {
	h := map[string]string{"alg": "HS256", "typ": "JWT"}
	if j.RSAPrivate != nil {
		h["alg"] = "RS256"
		if j.KeyID != "" {
			h["kid"] = j.KeyID
		}
	}
	b, _ := json.Marshal(h)
	return base64urlEncode(b)
}

func (j *JWTSigner) Sign(cl Claims) (string, error) {
	cl.Issuer = j.Issuer
	if cl.IssuedAt == 0 {
		cl.IssuedAt = time.Now().Unix()
	}
	h := j.encodeHeader()
	payloadBytes, err := json.Marshal(cl)
	if err != nil {
		return "", err
	}
	p := base64urlEncode(payloadBytes)
	signingInput := h + "." + p
	var sig string
	if j.RSAPrivate != nil {
		sum := sha256.Sum256([]byte(signingInput))
		s, err := rsa.SignPKCS1v15(rand.Reader, j.RSAPrivate, crypto.SHA256, sum[:])
		if err != nil {
			return "", err
		}
		sig = base64urlEncode(s)
	} else {
		sig = j.sign(signingInput)
	}
	return signingInput + "." + sig, nil
}

func (j *JWTSigner) ParseAndVerify(token string) (Claims, error) {
	var cl Claims
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return cl, errors.New("invalid token format")
	}
	signingInput := parts[0] + "." + parts[1]
	if j.RSAPublic != nil {
		sigBytes, err := base64urlDecode(parts[2])
		if err != nil {
			return cl, err
		}
		sum := sha256.Sum256([]byte(signingInput))
		if err := rsa.VerifyPKCS1v15(j.RSAPublic, crypto.SHA256, sum[:], sigBytes); err != nil {
			return cl, errors.New("invalid signature")
		}
	} else {
		expected := j.sign(signingInput)
		if !hmac.Equal([]byte(expected), []byte(parts[2])) {
			return cl, errors.New("invalid signature")
		}
	}
	payload, err := base64urlDecode(parts[1])
	if err != nil {
		return cl, err
	}
	if err := json.Unmarshal(payload, &cl); err != nil {
		return cl, err
	}
	if time.Now().Unix() > cl.Expires {
		return cl, errors.New("token expired")
	}
	if cl.Issuer != j.Issuer {
		return cl, errors.New("invalid issuer")
	}
	return cl, nil
}

// RSA helpers and JWKS
type JWKS struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid,omitempty"`
	Use string `json:"use,omitempty"`
	Alg string `json:"alg,omitempty"`
	N   string `json:"n,omitempty"`
	E   string `json:"e,omitempty"`
}

func RSAPublicToJWK(pub *rsa.PublicKey, kid string) JWK {
	n := base64urlEncode(pub.N.Bytes())
	// standard exponent e=65537 encodes to AQAB
	eBytes := []byte{0x01, 0x00, 0x01}
	if pub.E != 65537 {
		// generic encoding
		eBytes = bigIntBytes(pub.E)
	}
	return JWK{Kty: "RSA", Kid: kid, Use: "sig", Alg: "RS256", N: n, E: base64urlEncode(eBytes)}
}

func bigIntBytes(e int) []byte {
	// minimal big-endian bytes for small ints
	if e <= 0xFF {
		return []byte{byte(e)}
	}
	if e <= 0xFFFF {
		return []byte{byte(e >> 8), byte(e)}
	}
	return []byte{byte(e >> 16), byte(e >> 8), byte(e)}
}

func ParseRSAPrivateFromPEM(pemBytes []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}
	pkcs8, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rk, ok := pkcs8.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not RSA private key")
	}
	return rk, nil
}

func ParseRSAPublicFromPEM(pemBytes []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("invalid PEM")
	}
	if pk, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		if rk, ok := pk.(*rsa.PublicKey); ok {
			return rk, nil
		}
		return nil, errors.New("not RSA public key")
	}
	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		if rk, ok := cert.PublicKey.(*rsa.PublicKey); ok {
			return rk, nil
		}
		return nil, errors.New("not RSA public key in cert")
	}
	return nil, errors.New("unsupported public key format")
}
