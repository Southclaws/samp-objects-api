package main

import (
	"encoding/json"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
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
	"github.com/nfnt/resize"
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

// ObjectsList handles the /objects endpoint, it returns a query result of objects
func (app App) ObjectsList(w http.ResponseWriter, r *http.Request) {
	objects, err := app.Storage.GetObjects()
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, err)
		return
	}

	payload, err := json.Marshal(objects)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, err)
		return
	}

	w.Write(payload)
}

// Objects handles the /objects/:objectid endpoint, it returns the metadata for a specific object
func (app App) Objects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawID, ok := vars["objectid"]
	if !ok {
		WriteResponseError(w, http.StatusBadRequest, errors.New("no object ID specified"))
		return
	}
	objectID := types.ObjectID(rawID)

	objects, err := app.Storage.GetObject(objectID)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, err)
		return
	}

	payload, err := json.Marshal(objects)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, err)
		return
	}

	w.Write(payload)
}

// ObjectImage handles requests for object image thumbails
func (app App) ObjectImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rawID, ok := vars["objectid"]
	if !ok {
		WriteResponseError(w, http.StatusBadRequest, errors.New("no object ID specified"))
		return
	}
	objectID := types.ObjectID(rawID)

	err := app.Storage.GetObjectThumb(objectID, w)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to get object image"))
		return
	}
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

	user, err := app.Storage.GetUser(userID)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.New("failed to get user details by ID"))
		return
	}

	object.OwnerID = user.ID
	object.OwnerName = user.Name

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
	resp := UploadResponse{}
	defer func() {
		payload, err := json.Marshal(&resp)
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

				switch filepath.Ext(filename) {
				case ".dff":
					object.Models = append(object.Models, types.File(filename))

					err = app.Storage.PutObjectFile(objectID, filepath.Base(filename), int64(size), p)
					if err != nil {
						logger.Error("failed to write object file to block store",
							zap.Error(err))
						resp.Error = err.Error()
						return
					}
				case ".txd":
					object.Textures = append(object.Textures, types.File(filename))

					err = app.Storage.PutObjectFile(objectID, filepath.Base(filename), int64(size), p)
					if err != nil {
						logger.Error("failed to write object file to block store",
							zap.Error(err))
						resp.Error = err.Error()
						return
					}
				default:
					if len(object.Images) == 0 {
						img, _, err := image.Decode(p)
						if err != nil {
							logger.Error("unsupported image format",
								zap.Error(err))
							resp.Error = err.Error()
							return
						}

						reader, writer := io.Pipe()

						go func() {
							img = resize.Thumbnail(200, 200, img, resize.NearestNeighbor)
							err = jpeg.Encode(writer, img, &jpeg.Options{64})
							if err != nil {
								logger.Error("failed to encode thumbnail",
									zap.Error(err))
								resp.Error = err.Error()
								err = writer.CloseWithError(err)
								if err != nil {
									logger.Error("failed to close pipe",
										zap.Error(err))
									return
								}
								return
							}
							err = writer.Close()
							if err != nil {
								logger.Error("failed to close writer",
									zap.Error(err))
								resp.Error = err.Error()
								return
							}
						}()

						err = app.Storage.PutObjectFile(objectID, filepath.Base(filename), -1, reader)
						if err != nil {
							logger.Error("failed to write object file to block store",
								zap.Error(err))
							resp.Error = err.Error()
							err = reader.CloseWithError(err)
							if err != nil {
								logger.Error("failed to close pipe",
									zap.Error(err))
								return
							}
							return
						}
						err = reader.Close()
						if err != nil {
							logger.Error("failed to close reader",
								zap.Error(err))
							resp.Error = err.Error()
							return
						}
					} else {
						err = app.Storage.PutObjectFile(objectID, filepath.Base(filename), int64(size), p)
						if err != nil {
							logger.Error("failed to write object file to block store",
								zap.Error(err))
							resp.Error = err.Error()
							return
						}
					}
					object.Images = append(object.Images, types.File(filename))
				}

				uploadedFile = true
			} else {
				raw, err := ioutil.ReadAll(p)
				if err != nil {
					logger.Error("failed to read upload metadata chunk",
						zap.Error(err))
					resp.Error = err.Error()
					return
				}

				size, err = strconv.Atoi(string(raw))
				if err != nil {
					continue
				}
			}
		}
	}
	app.Pending.Set(string(object.ID), object, time.Minute*30)

	if uploadedFile && len(object.Images) > 0 && len(object.Models) > 0 && len(object.Textures) > 0 {
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
