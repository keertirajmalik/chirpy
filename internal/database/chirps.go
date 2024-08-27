package database

type Chirp struct {
	ID       int    `json:"id"`
	Body     string `json:"body"`
	AuthorID int    `json:"author_id"`
}

func (db *DB) CreateChirp(body string, userId int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	var chripID int = 1

	totalChirp := len(dbStructure.Chirps)
	if totalChirp > 0 {
		chirp, err := db.GetChirp(totalChirp)
		if err != nil {
			return Chirp{}, err
		}
		chripID = chirp.ID + 1
	}

	chirp := Chirp{
		ID:       chripID,
		Body:     body,
		AuthorID: userId,
	}
	dbStructure.Chirps[chripID] = chirp

	err = db.writeDB(dbStructure)

	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	dBStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	chirps := make([]Chirp, 0, len(dBStructure.Chirps))
	for _, chirp := range dBStructure.Chirps {
		chirps = append(chirps, chirp)
	}
	return chirps, nil
}

func (db *DB) GetChirp(id int) (Chirp, error) {
	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	chirp, ok := dbStructure.Chirps[id]
	if !ok {
		return Chirp{}, ErrNotExist
	}

	return chirp, nil
}

func (db *DB) DeleteChirp(chripId, userID int) error {
	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	chirp, ok := dbStructure.Chirps[chripId]
	if !ok {
		return ErrNotExist
	}

	delete(dbStructure.Chirps, chirp.ID)

	return db.writeDB(dbStructure)
}
