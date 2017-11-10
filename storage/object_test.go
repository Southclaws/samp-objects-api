package storage

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"bitbucket.org/Southclaws/samp-objects-api/types"
	minio "github.com/minio/minio-go"
)

func must(err error, rest ...interface{}) {
	if err != nil {
		panic(err)
	}
}

// Create some dummy users to own objects
func TestDatabase_ObjectOwners(t *testing.T) {
	must(db.CreateUser(types.User{"00000001-0000-0000-0000-000000000000", "owner1", "ownermail1", "pass1"}))
	must(db.CreateUser(types.User{"00000002-0000-0000-0000-000000000000", "owner2", "ownermail2", "pass2"}))
	must(db.CreateUser(types.User{"00000003-0000-0000-0000-000000000000", "owner3", "ownermail3", "pass3"}))
}

func TestDatabase_CreateObject(t *testing.T) {
	type args struct {
		object types.Object
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"v owner1 object1", args{types.Object{
			ID:          "00000000-0000-0000-0000-100000000000",
			Owner:       types.UserID("00000001-0000-0000-0000-000000000000"),
			Name:        "object1",
			Description: "object1",
			Category:    "category1",
			Tags:        []types.ObjectTag{"tag1"},
			ImageHash:   "123",
			ModelHash:   "456",
			TextureHash: "789",
		}}, false},
		{"v owner1 object2", args{types.Object{
			ID:          "00000000-0000-0000-0000-200000000000",
			Owner:       types.UserID("00000001-0000-0000-0000-000000000000"),
			Name:        "object2",
			Description: "object2",
			Category:    "category1",
			Tags:        []types.ObjectTag{"tag1"},
			ImageHash:   "123",
			ModelHash:   "456",
			TextureHash: "789",
		}}, false},
		{"v owner2 object1", args{types.Object{
			ID:          "00000000-0000-0000-0000-300000000000",
			Owner:       types.UserID("00000002-0000-0000-0000-000000000000"),
			Name:        "object1",
			Description: "object1",
			Category:    "category1",
			Tags:        []types.ObjectTag{"tag1"},
			ImageHash:   "123",
			ModelHash:   "456",
			TextureHash: "789",
		}}, false},
		{"i owner2 object1", args{types.Object{
			ID:          "00000000-0000-0000-0000-300000000000",
			Owner:       types.UserID("00000002-0000-0000-0000-000000000000"),
			Name:        "object1",
			Description: "object1",
			Category:    "category1",
			Tags:        []types.ObjectTag{"tag1"},
			ImageHash:   "123",
			ModelHash:   "456",
			TextureHash: "789",
		}}, true},
		{"i owner2 object1", args{types.Object{
			ID:          "not a uuid-0000-0000-0000-500000000000",
			Owner:       types.UserID("00000002-0000-0000-0000-000000000000"),
			Name:        "object1",
			Description: "object1",
			Category:    "category1",
			Tags:        []types.ObjectTag{"tag1"},
			ImageHash:   "123",
			ModelHash:   "456",
			TextureHash: "789",
		}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// read the test files into memory
			model, err := ioutil.ReadFile("./tests/test_model.dff")
			if err != nil {
				panic(err)
			}
			texture, err := ioutil.ReadFile("./tests/test_texture.txd")
			if err != nil {
				panic(err)
			}

			objectData := types.ObjectFiles{
				Models: []types.ObjectDFF{
					{Name: "test_model.dff", Data: model},
				},
				Textures: []types.ObjectTXD{
					{Name: "test_texture.txd", Data: texture},
				},
			}

			// do the test
			if err := db.CreateObject(tt.args.object, objectData); (err != nil) != tt.wantErr {
				t.Errorf("Database.CreateObject() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// ensure files are present in S3
				_, err = db.store.StatObject(db.StoreBucket, filepath.Join("/", string(tt.args.object.ID), "test_model.dff"), minio.StatObjectOptions{})
				if err != nil {
					panic(err)
				}

				_, err = db.store.StatObject(db.StoreBucket, filepath.Join("/", string(tt.args.object.ID), "test_texture.txd"), minio.StatObjectOptions{})
				if err != nil {
					panic(err)
				}
			}
		})
	}
}

// At this point the following objects are in the database:
// types.Object{ID: "00000000-0000-0000-0000-100000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object1",Description: "object1",Category: "category1",Tags: []types.ObjectTag{"tag1"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}
// types.Object{ID: "00000000-0000-0000-0000-200000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object2",Description: "object2",Category: "category1",Tags: []types.ObjectTag{"tag1"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}
// types.Object{ID: "00000000-0000-0000-0000-300000000000",Owner: types.UserID("00000002-0000-0000-0000-000000000000"),Name: "object1",Description: "object1",Category: "category1",Tags: []types.ObjectTag{"tag1"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}

func TestDatabase_UpdateObject(t *testing.T) {
	type args struct {
		object types.Object
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"v owner1 object1", args{types.Object{"00000000-0000-0000-0000-100000000000", types.UserID("00000001-0000-0000-0000-000000000000"), "object1", "object1", "category1", []types.ObjectTag{"tag2"}, "123", "456", "789"}}, false},
		{"v owner1 object2", args{types.Object{"00000000-0000-0000-0000-200000000000", types.UserID("00000001-0000-0000-0000-000000000000"), "object2", "object2", "category2", []types.ObjectTag{"tag1"}, "123", "456", "789"}}, false},
		{"v owner2 object1", args{types.Object{"00000000-0000-0000-0000-300000000000", types.UserID("00000002-0000-0000-0000-000000000000"), "object1", "object1", "category1", []types.ObjectTag{"tag1", "tag2"}, "123", "456", "789"}}, false},
		{"i owner2 object4", args{types.Object{"00000000-0000-0000-0000-400000000000", types.UserID("00000002-0000-0000-0000-000000000000"), "object1", "object1", "category1", []types.ObjectTag{"tag1"}, "123", "456", "789"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.UpdateObject(tt.args.object); (err != nil) != tt.wantErr {
				t.Errorf("Database.UpdateObject() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// At this point the following objects are in the database:
// types.Object{ID: "00000000-0000-0000-0000-100000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object1",Description: "object1",Category: "category1",Tags: []types.ObjectTag{"tag2"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}
// types.Object{ID: "00000000-0000-0000-0000-200000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object2",Description: "object2",Category: "category2",Tags: []types.ObjectTag{"tag1"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}
// types.Object{ID: "00000000-0000-0000-0000-300000000000",Owner: types.UserID("00000002-0000-0000-0000-000000000000"),Name: "object1",Description: "object1",Category: "category1",Tags: []types.ObjectTag{"tag1", "tag2"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}

func TestDatabase_DeleteObject(t *testing.T) {
	type args struct {
		objectID types.ObjectID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"v owner1 object1", args{"00000000-0000-0000-0000-300000000000"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.DeleteObject(tt.args.objectID); (err != nil) != tt.wantErr {
				t.Errorf("Database.DeleteObject() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// ensure files are present in S3
				_, err := db.store.StatObject(db.StoreBucket, filepath.Join("/", string(tt.args.objectID), "test_model.dff"), minio.StatObjectOptions{})
				if err == nil {
					panic(err)
				}

				_, err = db.store.StatObject(db.StoreBucket, filepath.Join("/", string(tt.args.objectID), "test_texture.txd"), minio.StatObjectOptions{})
				if err == nil {
					panic(err)
				}
			}
		})
	}
}

// At this point the following objects are in the database:
// types.Object{ID: "00000000-0000-0000-0000-100000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object1",Description: "object1",Category: "category1",Tags: []types.ObjectTag{"tag2"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}
// types.Object{ID: "00000000-0000-0000-0000-200000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object2",Description: "object2",Category: "category2",Tags: []types.ObjectTag{"tag1"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}

func TestDatabase_GetObject(t *testing.T) {
	type args struct {
		id types.ObjectID
	}
	tests := []struct {
		name       string
		args       args
		wantObject types.Object
		wantErr    bool
	}{
		{"v owner1 object1", args{types.ObjectID("00000000-0000-0000-0000-100000000000")}, types.Object{ID: "00000000-0000-0000-0000-100000000000", Owner: types.UserID("00000001-0000-0000-0000-000000000000"), Name: "object1", Description: "object1", Category: "category1", Tags: []types.ObjectTag{"tag2"}, ImageHash: "123", ModelHash: "456", TextureHash: "789"}, false},
		{"i no exist", args{types.ObjectID("00000000-0000-0000-0000-010000000000")}, types.Object{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotObject, err := db.GetObject(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotObject, tt.wantObject) {
				t.Errorf("Database.GetObject() = %v, want %v", gotObject, tt.wantObject)
			}
		})
	}
}

// At this point the following objects are in the database:
// types.Object{ID: "00000000-0000-0000-0000-100000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object1",Description: "object1",Category: "category1",Tags: []types.ObjectTag{"tag2"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}
// types.Object{ID: "00000000-0000-0000-0000-200000000000",Owner: types.UserID("00000001-0000-0000-0000-000000000000"),Name: "object2",Description: "object2",Category: "category2",Tags: []types.ObjectTag{"tag1"},ImageHash: "123",ModelHash: "456",TextureHash: "789",}

func TestDatabase_ObjectExists(t *testing.T) {
	type args struct {
		objectID types.ObjectID
	}
	tests := []struct {
		name       string
		args       args
		wantExists bool
		wantErr    bool
	}{
		{"v object1", args{types.ObjectID("00000000-0000-0000-0000-100000000000")}, true, false},
		{"v object2", args{types.ObjectID("00000000-0000-0000-0000-200000000000")}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExists, err := db.ObjectExists(tt.args.objectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.ObjectExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("Database.ObjectExists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}
