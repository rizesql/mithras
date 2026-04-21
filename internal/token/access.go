package token

import (
	"context"
	"time"

	"github.com/rizesql/mithras/internal/jws"
	"github.com/rizesql/mithras/internal/jwt"
	"github.com/rizesql/mithras/pkg/telemetry"
)

type Access struct {
	jwt.Token
}

type Issuer struct {
	jws    jws.Store
	issuer string
}

func NewIssuer(ks jws.Store, issuer string) Issuer {
	return Issuer{jws: ks, issuer: issuer}
}

type IssueConfig struct {
	IssuedAt time.Time
	Subject  string
	Duration time.Duration
	Roles    []string
}

func (i Issuer) Issue(ctx context.Context, cfg IssueConfig) (access Access, err error) {
	ctx, span := telemetry.Start(ctx, "token.issue")
	defer telemetry.End(span, &err)

	key, err := i.jws.SigningKey(ctx)
	if err != nil {
		return Access{}, err
	}

	tok, err := jwt.Sign(key, jwt.Claims{
		Issuer:    i.issuer,
		Subject:   cfg.Subject,
		IssuedAt:  cfg.IssuedAt.Unix(),
		ExpiresAt: cfg.IssuedAt.Add(cfg.Duration).Unix(),
		Roles:     cfg.Roles,
	})
	if err != nil {
		return Access{}, err
	}

	return Access{tok}, nil
}
