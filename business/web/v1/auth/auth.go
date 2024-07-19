package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/open-policy-agent/opa/rego"
	"github.com/rmishgoog/starter-go-service/foundations/logger"
)

var ErrForbidden = errors.New("attempted action is not allowed")

type Config struct {
	Log *logger.Logger
	//DB        *sqlx.DB
	KeyLookup KeyLookup
	Issuer    string
}

type Auth struct {
	log       *logger.Logger
	keyLookup KeyLookup
	method    jwt.SigningMethod
	parser    *jwt.Parser
	issuer    string
	mu        sync.RWMutex
	cache     map[string]string
}

type Claims struct {
	jwt.RegisteredClaims
	Roles []string `json:"roles"`
}

type KeyLookup interface {
	PrivateKey(kid string) (key string, err error)
	PublicKey(kid string) (key string, err error)
}

func New(cfg Config) (*Auth, error) {

	// If a database connection is not provided, we won't perform the
	// user enabled check.
	//var usrCore *user.Core
	// if cfg.DB != nil {
	// 	usrCore = user.NewCore(cfg.Log, userdb.NewStore(cfg.Log, cfg.DB))
	// }

	a := Auth{
		log:       cfg.Log,
		keyLookup: cfg.KeyLookup,
		//usrCore:   usrCore,
		method: jwt.GetSigningMethod(jwt.SigningMethodRS256.Name),
		parser: jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name})),
		issuer: cfg.Issuer,
		cache:  make(map[string]string),
	}

	return &a, nil
}

// GenerateToken generates a signed JWT token string representing the user Claims.
func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {
	token := jwt.NewWithClaims(a.method, claims)
	token.Header["kid"] = kid

	privateKeyPEM, err := a.keyLookup.PrivateKey(kid)
	if err != nil {
		return "", fmt.Errorf("private key: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return "", fmt.Errorf("parsing private pem: %w", err)
	}

	str, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}

	return str, nil
}

// Authenticate processes the token to validate the sender's token is valid.
func (a *Auth) Authenticate(ctx context.Context, bearerToken string) (Claims, error) {
	parts := strings.Split(bearerToken, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return Claims{}, errors.New("expected authorization header format: Bearer <token>")
	}

	var claims Claims
	token, _, err := a.parser.ParseUnverified(parts[1], &claims)
	if err != nil {
		return Claims{}, fmt.Errorf("error parsing token: %w", err)
	}

	// Perform an extra level of authentication verification with OPA.

	kidRaw, exists := token.Header["kid"]
	if !exists {
		return Claims{}, fmt.Errorf("kid missing from header: %w", err)
	}

	kid, ok := kidRaw.(string)
	if !ok {
		return Claims{}, fmt.Errorf("kid malformed: %w", err)
	}

	pem, err := a.publicKeyLookup(kid)
	if err != nil {
		return Claims{}, fmt.Errorf("failed to fetch public key: %w", err)
	}

	input := map[string]any{
		"Key":   pem,
		"Token": parts[1],
		"ISS":   a.issuer,
	}

	if err := a.opaPolicyEvaluation(ctx, opaAuthentication, RuleAuthenticate, input); err != nil {
		return Claims{}, fmt.Errorf("authentication failed : %w", err)
	}

	// Check the database for this user to verify they are still enabled.

	// if err := a.isUserEnabled(ctx, claims); err != nil {
	// 	return Claims{}, fmt.Errorf("user not enabled : %w", err)
	// }

	return claims, nil
}

func (a *Auth) publicKeyLookup(kid string) (string, error) {
	pem, err := func() (string, error) {
		a.mu.RLock()
		defer a.mu.RUnlock()

		pem, exists := a.cache[kid]
		if !exists {
			return "", errors.New("not found")
		}
		return pem, nil
	}()
	if err == nil {
		return pem, nil
	}

	pem, err = a.keyLookup.PublicKey(kid)
	if err != nil {
		return "", fmt.Errorf("fetching public key: %w", err)
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	a.cache[kid] = pem

	return pem, nil
}

// opaPolicyEvaluation asks opa to evaulate the token against the specified token
// policy and public key.
func (a *Auth) opaPolicyEvaluation(ctx context.Context, opaPolicy string, rule string, input any) error {
	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaPolicy),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

func (a *Auth) Authorize(ctx context.Context, claims Claims, userID uuid.UUID, rule string) error {
	input := map[string]any{
		"Roles":   claims.Roles,
		"Subject": claims.Subject,
		"UserID":  userID,
	}

	if err := a.opaPolicyEvaluation(ctx, opaAuthorization, rule, input); err != nil {
		return fmt.Errorf("rego evaluation failed : %w", err)
	}

	return nil
}

// func (a *Auth) isUserEnabled(ctx context.Context, claims Claims) error {
// 	if a.usrCore == nil {
// 		return nil
// 	}

// 	userID, err := uuid.Parse(claims.Subject)
// 	if err != nil {
// 		return fmt.Errorf("parse user: %w", err)
// 	}

// 	if _, err := a.usrCore.QueryByID(ctx, userID); err != nil {
// 		return fmt.Errorf("query user: %w", err)
// 	}

// 	return nil
// }
