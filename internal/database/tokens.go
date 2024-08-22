package database

import "time"

type RefreshToken struct {
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (db *DB) SaveRefreshToken(userID int, token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	refreshToken := RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	dbStructure.RefeshTokens[token] = refreshToken

	err = db.writeDB(dbStructure)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UserForRefershToken(token string) (User, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	refreshToken, ok := dbStructure.RefeshTokens[token]
	if !ok {
		return User{}, ErrNotExist
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		return User{}, ErrNotExist
	}

	user, err := db.GetUser(refreshToken.UserID)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func (db *DB) RevokeToken(token string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	delete(dbStructure.RefeshTokens, token)

	return db.writeDB(dbStructure)
}
