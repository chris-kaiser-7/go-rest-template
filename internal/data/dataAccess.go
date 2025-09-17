package data

import (
	"database/sql"
	"errors"
	"log"
	"os"
)

var (
	// ErrRecordNotFound is returned when a movie record doesn't exist in database.
	ErrRecordNotFound = errors.New("record not found")

	// ErrEditConflict is returned when a there is a data race, and we have an edit conflict.
	ErrEditConflict = errors.New("edit conflict")
)

// DataAccessWrapers struct is a single convenient container to hold and represent all our database access wrappers.
type DataAccessWrapers struct {
	ApiKeys     ApiKeyDataAccess
	Users       UserDataAccess
	Tokens      TokenDataAccess
	Permissions PermissionDataAccess
}

func InitDataAccess(db *sql.DB) DataAccessWrapers {
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	return DataAccessWrapers{
		ApiKeys: ApiKeyDataAccess{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Users: UserDataAccess{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Tokens: TokenDataAccess{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
		Permissions: PermissionDataAccess{
			DB:       db,
			InfoLog:  infoLog,
			ErrorLog: errorLog,
		},
	}
}
