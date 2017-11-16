package storage

import (
	"testing"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

func TestDatabase_AddRating(t *testing.T) {
	type args struct {
		userID   types.UserID
		objectID types.ObjectID
		value    float64
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid", args{
			"00000003-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			3.4,
		}, false},
		{"valid", args{
			"00000002-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			4.4,
		}, false},
		{"valid", args{
			"00000001-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			4.6,
		}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.AddRating(tt.args.userID, tt.args.objectID, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Database.AddRating() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
