package model

func (u User) IsCorpUser() bool {
	return u.TenantID != ""
}
