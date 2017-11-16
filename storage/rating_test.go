package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

func TestDatabase_AddRating(t *testing.T) {
	type args struct {
		userID   types.UserID
		objectID types.ObjectID
		value    float64
	}
	tests := []struct {
		name       string
		args       args
		wantExists bool
		wantErr    bool
	}{
		{"valid", args{
			"00000003-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			3.4,
		}, false, false},
		{"valid", args{
			"00000002-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			4.4,
		}, false, false},
		{"valid", args{
			"00000001-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			4.6,
		}, false, false},
		{"valid", args{
			"00000001-0000-0000-0000-000000000000",
			"00000000-0000-0000-0000-200000000000",
			4.6,
		}, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExists, err := db.AddRating(tt.args.userID, tt.args.objectID, tt.args.value)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantExists, gotExists)
		})
	}
}
