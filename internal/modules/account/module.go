package account

import (
	"go.uber.org/fx"

	"poster.app/studio/internal/platform/auth"
	"poster.app/studio/internal/platform/routing"
)

var Module = fx.Module(
	"account",
	fx.Provide(NewRepository),
	fx.Provide(NewResource),
	fx.Provide(NewSessionResolver),
	fx.Provide(func(resolver *SessionResolver) auth.Resolver { return resolver }),
	fx.Provide(
		fx.Annotate(
			NewController,
			fx.As(new(routing.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
