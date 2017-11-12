package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/Southclaws/samp-objects-api/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// UploadResponse is the response returned when a client successfully uploads a file
// this is for FineUploader
type UploadResponse struct {
	Success bool         `json:"success"`
	Error   string       `json:"error"`
	Object  types.Object `json:"object"`
}

// Objects handles the /objects/:object endpoint, it includes actions for listing all objects,
// viewing info for a specific object and uploading a new object.
func (app App) Objects(w http.ResponseWriter, r *http.Request) {
}

// PrepareObject receives a types.Object and caches it while responding with the generated unique ID
// so the client can begin uploading files for that object.
func (app App) PrepareObject(w http.ResponseWriter, r *http.Request) {
	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return
	}

	var object types.Object
	err = json.Unmarshal(payload, &object)
	if err != nil {
		return
	}

	err = object.ValidatePartial()
	if err != nil && err.Error() != "id does not match pattern" {
		return
	}

	session, err := app.Sessions.Get(r, UserSessionCookie)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.New("failed to read session cookies"))
		return
	}
	userRaw, ok := session.Values["UserID"]
	if !ok {
		WriteResponseError(w, http.StatusInternalServerError, errors.New("no UserID field in request"))
		return
	}
	userID, ok := userRaw.(types.UserID)
	if !ok {
		WriteResponseError(w, http.StatusInternalServerError, errors.New("failed to decode user ID from request"))
		return
	}

	object.Owner = userID

	exists, err := app.Storage.UserObjectExists(object)
	if err != nil {
		return
	}
	if exists {
		WriteResponseError(w, http.StatusConflict, errors.New("object name already in use by user"))
		return
	}

	object.ID = types.ObjectID(uuid.New().String())

	// object ID stores the object metadata
	app.Pending.Set(string(object.ID), object, time.Minute*30)
	// object ID + PENDING indicates whether or not the object has been created in the database yet
	app.Pending.Set(string(object.ID)+"-PENDING", false, time.Minute*30)

	payload, err = json.Marshal(object)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(payload)
}

// ObjectUpload is the file upload endpoint
func (app App) ObjectUpload(w http.ResponseWriter, r *http.Request) {
	resp := &UploadResponse{}
	defer func() {
		payload, err := json.Marshal(resp)
		if err != nil {
			WriteResponseError(w, http.StatusBadRequest, err)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(payload))
	}()

	vars := mux.Vars(r)
	objectID := types.ObjectID(vars["objectid"])
	if err := objectID.Validate(); err != nil {
		logger.Error("failed to validate object ID",
			zap.Error(err))
		resp.Error = err.Error()
		return
	}

	objRaw, hit := app.Pending.Get(string(objectID))
	if !hit {
		resp.Error = "no cached object metadata by that ID"
		return
	}
	object, ok := objRaw.(types.Object)
	if !ok {
		resp.Error = "failed to decode cached object ID, please try again"
		app.Pending.Delete(string(objectID))
		return
	}

	if object.ID != objectID {
		resp.Error = "request object ID and cached object ID do not match, please try again"
		app.Pending.Delete(string(objectID))
		return
	}

	logger.Debug("received upload request", zap.String("objectID", string(objectID)))

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		logger.Error("failed to parse media type from Content-Type header",
			zap.Error(err))
		resp.Error = err.Error()
		return
	}
	uploadedFile := false
	if strings.HasPrefix(mediaType, "multipart/") {
		size := 0
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				logger.Error("failed to read next part of multipart body",
					zap.Error(err))
				resp.Error = err.Error()
				return
			}
			filename := p.FileName()
			if filename != "" {
				if size == 0 {
					logger.Error("no size found in multipart request",
						zap.Error(err))
					resp.Error = err.Error()
					return
				}

				logger.Debug("received actual file",
					zap.String("filename", filename))

				err = app.Storage.PutObjectFile(objectID, filepath.Base(filename), int64(size), p)
				if err != nil {
					logger.Error("failed to write object file to block store",
						zap.Error(err))
					resp.Error = err.Error()
					return
				}

				switch filepath.Ext(filename) {
				case ".dff":
					object.Models = append(object.Models, types.File(filename))
				case ".txd":
					object.Textures = append(object.Textures, types.File(filename))
				default:
					object.Images = append(object.Images, types.File(filename))
				}

				logger.Debug("uploaded file successfully",
					zap.Any("object", object))

				uploadedFile = true
			} else {
				raw, err := ioutil.ReadAll(p)
				if err != nil {
					logger.Error("failed to read upload metadata chunk",
						zap.Error(err))
					resp.Error = err.Error()
					return
				}
				logger.Debug("received upload metadata",
					zap.ByteString("data", raw))

				size, err = strconv.Atoi(string(raw))
				if err != nil {
					continue
				}
			}
		}
	}
	app.Pending.Set(string(object.ID), object, time.Minute*30)

	if uploadedFile && len(object.Models) > 0 && len(object.Textures) > 0 {
		_, hit = app.Pending.Get(string(object.ID) + "-PENDING")
		if hit { // object pending creation in DB
			err = app.Storage.CreateObject(object)
			if err != nil {
				logger.Error("failed to create object metadata in database",
					zap.Error(err))
				resp.Error = err.Error()
				return
			}
			app.Pending.Delete(string(object.ID) + "-PENDING")
		} else { // object is already in DB, update it
			err = app.Storage.UpdateObject(object)
			if err != nil {
				logger.Error("failed to update object metadata in databasea",
					zap.Error(err))
				resp.Error = err.Error()
				return
			}
		}
	}

	resp.Object = object
	resp.Success = true
}

// get object
func (app App) getObject(w http.ResponseWriter, r *http.Request) {
	// objectID := types.ObjectID(vars["objectid"])
	// if err := objectID.Validate(); err != nil {
	// 	// todo: error
	// 	return
	// }
	w.WriteHeader(http.StatusNotImplemented)
}

// update object
func (app App) updateObject(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
}
