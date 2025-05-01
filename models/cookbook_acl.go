package models

type CookbookAcl struct {
	UserID   int64  `db:"user_id"`
	Username string `db:"username"`
	Email    string `db:"email"`
	CanEdit  bool   `db:"can_edit"`
	CanView  bool   `db:"can_view"`
	IsOwner  bool   `db:"is_owner"`
}
