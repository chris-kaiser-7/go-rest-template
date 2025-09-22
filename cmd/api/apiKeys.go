package main

import (
	"github.com/chris-a-kaiser-7/go-rest-template/internal/data"
	"github.com/chris-a-kaiser-7/go-rest-template/internal/validator"
	"net/http"
)

type KeyReturn struct {
	data.ApiKeyData
	uses int
}

func (app *application) createApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		KeyName string `json:"keyName"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	user := app.contextGetUser(r)
	newKey := data.ApiKeyData{
		UserId:  user.ID,
		KeyName: input.KeyName,
	}

	err = app.models.ApiKeys.Create(&newKey)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"apiKey": newKey}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) validateApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Key string `json:"keyName"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	fetchedKey, err := app.models.ApiKeys.Validate([]byte(input.Key))
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"apiKey": fetchedKey}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	keyId, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	keyData, err := app.models.ApiKeys.Get(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	user := app.contextGetUser(r)
	if user.ID != keyData.UserId {
		app.invalidCredentialsResponse(w, r)
		return
	}

	usage, err := app.models.ApiKeyUsage.GetUsageDataByKey(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ApiKey": KeyReturn{ApiKeyData: keyData, uses: usage}, "uses": usage}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) getAllApiKeyUsageHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Sort = app.readStrings(qs, "sort", "id")
	input.SortSafeList = []string{"id", "-id"}

	user := app.contextGetUser(r)
	keys, metadata, err := app.models.ApiKeys.GetAll(user.ID, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	keyReturns := make([]KeyReturn, len(keys))
	for i, key := range keys {
		c, err := app.models.ApiKeyUsage.GetUsageDataByKey(key.Id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		keyReturns[i] = KeyReturn{ApiKeyData: key, uses: c}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "ApiKeys": keyReturns}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) deactivateApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	keyId, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user := app.contextGetUser(r)
	keyData, err := app.models.ApiKeys.Get(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if user.ID != keyData.UserId {
		app.invalidCredentialsResponse(w, r)
		return
	}

	err = app.models.ApiKeys.Deactivate(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "api succesfully deactivated."}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) adminGetApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	keyId, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	keyData, err := app.models.ApiKeys.Get(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	usage, err := app.models.ApiKeyUsage.GetUsageDataByKey(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"ApiKey": KeyReturn{ApiKeyData: keyData, uses: usage}, "uses": usage}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) adminGetAllApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		userId int64
		data.Filters
	}

	v := validator.New()
	qs := r.URL.Query()
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Sort = app.readStrings(qs, "sort", "id")
	input.SortSafeList = []string{"id", "-id"}

	keys, metadata, err := app.models.ApiKeys.GetAll(input.userId, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	keyReturns := make([]KeyReturn, len(keys))
	for i, key := range keys {
		c, err := app.models.ApiKeyUsage.GetUsageDataByKey(key.Id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		keyReturns[i] = KeyReturn{ApiKeyData: key, uses: c}
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "ApiKeys": keyReturns}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) adminDeactivateApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	keyId, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.ApiKeys.Deactivate(keyId)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "api succesfully deactivated."}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}
