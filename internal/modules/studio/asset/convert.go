package asset

func AssetEntityToSimple(entity *Asset) *AssetSimple {
	return &AssetSimple{
		ID: entity.ID, Name: entity.OriginalName, ContentType: entity.ContentType,
		ByteSize: entity.ByteSize, Width: entity.Width, Height: entity.Height,
		CreatedAt: entity.CreatedAt.UTC(),
	}
}

func AssetEntityToResponse(entity *Asset) *AssetResponse {
	return &AssetResponse{
		ID: entity.ID, Name: entity.OriginalName, ContentType: entity.ContentType,
		ByteSize: entity.ByteSize, Width: entity.Width, Height: entity.Height,
		CreatedAt: entity.CreatedAt.UTC(),
	}
}
