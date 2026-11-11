package poster

import (
	"go.uber.org/fx"

	"poster.app/studio/internal/platform/routing"
)

var Module = fx.Module(
	"poster",
	fx.Provide(NewRepository),
	fx.Provide(NewResource),
	fx.Provide(
		fx.Annotate(
			NewController,
			fx.As(new(routing.Registrar)),
			fx.ResultTags(`group:"routes"`),
		),
	),
)
