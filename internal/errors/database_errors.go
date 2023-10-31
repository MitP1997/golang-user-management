package errors

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ValidationError = func(e error) *Error { return &Error{Type: typeBadRequest, Err: e, DisplayString: e.Error()} }
	DatabaseError   = func(e error) *Error {
		return &Error{Type: typeInternalServer, Err: e, DisplayString: "Internal server error"}
	}
	NoDocumentsError = func(e error, document string) *Error {
		return &Error{Type: typeNotFound, Err: e, DisplayString: fmt.Sprintf("No %s found", document)}
	}
	DuplicateKeyError = func(we mongo.WriteException) *Error {
		var doc map[string]interface{}
		err := bson.Unmarshal([]byte(we.WriteErrors[0].Raw), &doc)
		if err != nil {
			fmt.Println(err)
		}
		keyValue := doc["keyValue"].(map[string]interface{})
		var displayString string
		for key, value := range keyValue {
			displayString = fmt.Sprintf("%s %s: %s", displayString, key, value)
		}
		return &Error{Type: typeBadRequest, Err: we, DisplayString: fmt.Sprintf("%s already exists in the database", displayString)}
	}
	IncorrectIdFormatError = func(e error) *Error {
		return &Error{Type: typeBadRequest, Err: e, DisplayString: "Incorrect id format"}
	}
)
