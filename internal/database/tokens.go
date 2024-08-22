package database

import "time"

type Token struct {
	RefreshToken string        `json:"refresh_token"`
	Token        string        `json:"token"`
	ExpiresAt    time.Duration `json:"expires_at"`
}

func (db *DB) CreateRefreshToken(token, refreshToken string) (Token, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Token{}, err
	}

	newToken := Token{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Duration(time.Now().UTC().AddDate(0, 0, 60).Second()),
	}
	dbStructure.Tokens[refreshToken] = newToken

	err = db.writeDB(dbStructure)

	if err != nil {
		return Token{}, err
	}

	return newToken, nil
}

func (db *DB) GetToken(refreshToken string) (Token, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Token{}, err
	}

	token, ok := dbStructure.Tokens[refreshToken]
	if !ok {
		return Token{}, ErrNotExist
	}

	return token, nil
}
func (db *DB) DeleteToken(refreshToken string) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	token, ok := dbStructure.Tokens[refreshToken]
	if !ok {
		return ErrNotExist
	}

    delete(dbStructure.Tokens, token.RefreshToken)

    return db.writeDB(dbStructure)
}
