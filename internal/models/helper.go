package models

import (
	"github.com/MitP1997/golang-user-management/internal/errors"
	"go.mongodb.org/mongo-driver/mongo"
)

func getErrorToReturn(e error, collection string) *errors.Error {
	if e == mongo.ErrNoDocuments {
		return errors.NoDocumentsError(e, collection)
	}
	writeException := e.(mongo.WriteException)
	if writeException.HasErrorCode(11000) {
		return errors.DuplicateKeyError(writeException)
	}
	return errors.DatabaseError(e)
}
