package main

import (
	"encoding/json"
	"image"
	_ "image/gif"
	"image/jpeg"
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
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

// UploadResponse is the response returned when a client successfully uploads a file
// this is for FineUploader
type UploadResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// ObjectsList handles the /objects endpoint, it returns a query result of objects
func (app *App) ObjectsList(w http.ResponseWriter, r *http.Request) {
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

// Objects handles the /objects/:userName/:objectName endpoint, it returns the metadata for a specific object
func (app *App) Objects(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userName := types.UserName(vars["userName"])
	objectName := types.ObjectName(vars["objectName"])

	objects, err := app.Storage.GetUserObject(userName, objectName)
	if err != nil {
		if err.Error() == "not found" {
			WriteResponse(w, http.StatusNotFound, err.Error())
			return
		}
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

// ObjectThumb handles requests for object image thumbails
func (app *App) ObjectThumb(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectID := types.ObjectID(vars["objectid"])

	err := app.Storage.GetObjectThumb(objectID, w)
	if err != nil {
		err = jpeg.Encode(w, image.NewGray(image.Rect(0, 0, 200, 200)), &jpeg.Options{50})
		if err != nil {
			WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to get object image"))
			return
		}
		return
	}
}

// ObjectFiles handles requests for object files by name
func (app *App) ObjectFiles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectID := types.ObjectID(vars["objectid"])
	fileName := types.File(vars["fileName"])

	err := app.Storage.GetObjectFile(objectID, fileName, w)
	if err != nil {
		WriteResponseError(w, http.StatusInternalServerError, errors.Wrap(err, "failed to get object image"))
		return
	}
}

// ObjectPrepare receives a types.Object and caches it while responding with the generated unique ID
// so the client can begin uploading files for that object.
func (app *App) ObjectPrepare(w http.ResponseWriter, r *http.Request) {
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

	app.UploadWaiter(object)

	logger.Debug("prepared new object upload waiter",
		zap.String("objectid", string(object.ID)))

	payload, err = json.Marshal(object)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(payload)
}

// ObjectUpload is the file upload endpoint
func (app *App) ObjectUpload(w http.ResponseWriter, r *http.Request) {
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

	logger.Debug("received upload request",
		zap.Any("objectid", objectID))

	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil {
		logger.Error("failed to parse media type from Content-Type header",
			zap.Error(err))
		resp.Error = err.Error()
		return
	}
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

				err = app.AddFile(objectID, filename, p)
				if err != nil {
					logger.Error("failed to update object upload cache",
						zap.Error(err))
					resp.Error = err.Error()
					return
				}

				logger.Debug("completed upload",
					zap.Any("objectid", objectID))

				resp.Success = true
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
}

// UploadWaiter is called when an upload is prepared and awaits files from ObjectUpload
func (app *App) UploadWaiter(object types.Object) {
	ch := make(chan types.ObjectFile, 16)

	app.Uploads.Store(string(object.ID), ActiveUpload{
		ch:     ch,
		object: object,
	})

	logger.Debug("created new object upload waiter",
		zap.String("objectid", string(object.ID)))

	duration := time.Minute * 30

	go func() {
		timeout := time.NewTimer(duration)
		success := true

		// make sure we don't leak map entries
		defer app.Uploads.Delete(string(object.ID))

		for {
			select {
			case file, ok := <-ch:
				if !ok {
					ch = nil
				}

				logger.Debug("waiter received file",
					zap.String("name", string(file.Name)),
					zap.String("type", string(file.Type)),
					zap.String("objectid", string(object.ID)))

				uploadRaw, ok := app.Uploads.Load(string(object.ID))
				if !ok {
					// todo: handle this error somehow
					logger.Error("failed to load object upload state while accepting a new file",
						zap.String("name", string(file.Name)),
						zap.String("type", string(file.Type)),
						zap.String("objectid", string(object.ID)))
					return
				}
				upload, ok := uploadRaw.(ActiveUpload)
				if !ok {
					// todo: handle this error somehow
					logger.Error("failed to decode object upload state while accepting a new file",
						zap.String("name", string(file.Name)),
						zap.String("type", string(file.Type)),
						zap.String("objectid", string(object.ID)))
					return
				}

				switch file.Type {
				case "image":
					upload.object.Images = append(upload.object.Images, types.File(file.Name))
				case "model":
					upload.object.Models = append(upload.object.Models, types.File(file.Name))
				case "texture":
					upload.object.Textures = append(upload.object.Textures, types.File(file.Name))
				}

				app.Uploads.Store(string(object.ID), upload)

				timeout.Reset(duration)
			case <-timeout.C:
				success = false
				close(ch)
			}
			if ch == nil {
				break
			}
		}

		if success {
			logger.Debug("object upload cache closed successfully, attempting to write to db",
				zap.String("objectid", string(object.ID)))

			err := app.Storage.CreateObject(object)
			if err != nil {
				logger.Error("failed to create object metadata in database",
					zap.Error(err),
					zap.String("objectid", string(object.ID)))
				return
			}
		} else {
			logger.Debug("dropped cache waiter for upload",
				zap.String("objectid", string(object.ID)))

			// todo: purge S3 bucket
		}
	}()
}

// AddFile adds a file from an upload to a pending waiter and writes the file to the file store
func (app *App) AddFile(objectID types.ObjectID, filename string, p io.Reader) (err error) {
	uploadRaw, ok := app.Uploads.Load(string(objectID))
	if !ok {
		return errors.New("no upload cache open with that ID")
	}
	upload, ok := uploadRaw.(ActiveUpload)
	if !ok {
		return errors.New("upload cache corrupted")
	}

	filetype := "image"

	// todo: use mimetypes/first bytes to determine file type
	switch filepath.Ext(filename) {
	case ".dff":
		filetype = "model"
	case ".txd":
		filetype = "texture"
	}

	r, w := io.Pipe()

	go func() {
		if filetype == "image" {
			img, format, err := image.Decode(p)
			if err != nil {
				w.CloseWithError(errors.Wrapf(err, "failed to decode image file of format '%s'", format))
			}

			err = jpeg.Encode(w, img, &jpeg.Options{64})
			if err != nil {
				w.CloseWithError(errors.Wrap(err, "failed to re-encode image as JPEG"))
			}
		} else {
			_, err := io.Copy(w, p)
			if err != nil {
				w.CloseWithError(err)
			}
		}
		err = w.Close()
		if err != nil {
			logger.Fatal("failed to close upload cache image writer",
				zap.String("objectid", string(objectID)),
				zap.String("filename", filename))
		}
	}()

	err = app.Storage.PutObjectFile(objectID, filename, r)
	if err != nil {
		return errors.Wrap(err, "failed to write object to store")
	}
	err = r.Close()
	if err != nil {
		return errors.Wrap(err, "failed to close reader")
	}

	upload.ch <- types.ObjectFile{
		Name: filename,
		Type: filetype,
	}

	return
}

// ObjectFinish finishes an upload process and closes the channel resulting in the object getting
// created in the database.
func (app *App) ObjectFinish(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objectID := types.ObjectID(vars["objectid"])
	if err := objectID.Validate(); err != nil {
		WriteResponseError(w, http.StatusBadRequest, err)
		return
	}

	logger.Debug("received request to finish upload",
		zap.String("objectid", string(objectID)))

	uploadRaw, ok := app.Uploads.Load(string(objectID))
	if !ok {
		WriteResponseError(w, http.StatusNotFound, errors.New("no upload cache open with that ID"))
		return
	}
	upload, ok := uploadRaw.(ActiveUpload)
	if !ok {
		WriteResponseError(w, http.StatusInternalServerError, errors.New("upload cache corrupted"))
		return
	}

	if err := upload.object.Validate(); err != nil {
		WriteResponse(w, http.StatusBadRequest, errors.Wrap(err, "object not valid").Error())
		return
	}

	close(upload.ch)

	logger.Debug("finished upload for object files",
		zap.String("objectid", string(objectID)))
}
