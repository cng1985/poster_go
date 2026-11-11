package account

func AccountEntityToResponse(entity *Account) *AccountResponse {
	return &AccountResponse{
		ID:          entity.ID,
		Username:    entity.Username,
		DisplayName: entity.DisplayName,
		CreatedAt:   entity.CreatedAt.UTC(),
	}
}
