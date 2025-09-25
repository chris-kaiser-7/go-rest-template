package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/chris-a-kaiser-7/go-rest-template/internal/data"
	"github.com/chris-a-kaiser-7/go-rest-template/internal/validator"
)

type formatedApiKey struct {
	Id        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	UserId    int64     `json:"user_id"`
	KeyName   string    `json:"key_name"`
	Key       string    `json:"key_value"`
	Uses      int       `json:"uses"`
}

func formatApiKey(key data.ApiKeyData, useCount int) formatedApiKey {
	return formatedApiKey{
		Id:        key.Id,
		CreatedAt: key.CreatedAt,
		UserId:    key.UserId,
		KeyName:   key.KeyName,
		Key:       string(key.Key),
		Uses:      useCount,
	}
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

	err = app.writeJSON(w, http.StatusOK, envelope{"apiKey": formatApiKey(newKey, 0)}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
}

func (app *application) validateApiKeyHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Key string `json:"key"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	fetchedKey, err := app.models.ApiKeys.Validate([]byte(input.Key))
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.invalidApiKeyValidation(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	uses, err := app.models.ApiKeyUsage.GetUsageDataByKey(fetchedKey.Id)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"apiKey": formatApiKey(fetchedKey, uses)}, nil)
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

	err = app.writeJSON(w, http.StatusOK, envelope{"ApiKey": formatApiKey(keyData, usage)}, nil)
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
	keyReturns := make([]formatedApiKey, len(keys))
	for i, key := range keys {
		c, err := app.models.ApiKeyUsage.GetUsageDataByKey(key.Id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		keyReturns[i] = formatApiKey(key, c)
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
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "API key succesfully deactivated."}, nil)
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

	err = app.writeJSON(w, http.StatusOK, envelope{"ApiKey": formatApiKey(keyData, usage)}, nil)
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

	keyReturns := make([]formatedApiKey, len(keys))
	for i, key := range keys {
		c, err := app.models.ApiKeyUsage.GetUsageDataByKey(key.Id)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		keyReturns[i] = formatApiKey(key, c)
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
