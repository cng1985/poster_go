package app

import (
	"go.uber.org/fx"

	"poster.app/studio/internal/modules/account"
	"poster.app/studio/internal/modules/studio"
	"poster.app/studio/internal/platform/ai"
	"poster.app/studio/internal/platform/database"
	"poster.app/studio/internal/platform/middleware"
	"poster.app/studio/internal/platform/server"
	"poster.app/studio/internal/platform/storage"
	"poster.app/studio/web"
)

var Module = fx.Options(
	database.Module,
	storage.Module,
	ai.Module,
	middleware.Module,
	account.Module,
	studio.Module,
	server.Module,
	web.Module,
	fx.Invoke(Bootstrap),
)
