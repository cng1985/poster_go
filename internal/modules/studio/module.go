package studio

import (
	"go.uber.org/fx"

	"poster.app/studio/internal/modules/studio/analysis"
	"poster.app/studio/internal/modules/studio/asset"
	"poster.app/studio/internal/modules/studio/generation"
	"poster.app/studio/internal/modules/studio/poster"
	"poster.app/studio/internal/modules/studio/template"
)

var Module = fx.Module(
	"studio",
	template.Module,
	asset.Module,
	poster.Module,
	generation.Module,
	analysis.Module,
)
