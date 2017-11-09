package storage

import (
	"reflect"
	"testing"

	"bitbucket.org/Southclaws/samp-objects-api/types"
)

// Note: these tests are sequential and each one depends on the state of the database left by the
// previous. Comments between each test function indicate the database state for the reader.

func TestDatabase_CreateUser(t *testing.T) {
	type args struct {
		user types.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"v user1", args{types.User{"10000000-0000-0000-0000-000000000000", "user1", "mail1", "pass1", []types.ObjectID{}}}, false},
		{"v user2", args{types.User{"20000000-0000-0000-0000-000000000000", "user2", "mail2", "pass2", []types.ObjectID{}}}, false},
		{"v user3", args{types.User{"30000000-0000-0000-0000-000000000000", "user3", "mail3", "pass3", []types.ObjectID{}}}, false},

		// already used name
		{"i user1 again", args{types.User{"40000000-0000-0000-0000-000000000000", "user1", "mail4", "pass4", []types.ObjectID{}}}, true},

		// already used mail
		{"i user5", args{types.User{"50000000-0000-0000-0000-000000000000", "user5", "mail3", "pass5", []types.ObjectID{}}}, true},

		// invalid fielss
		{"i user6", args{types.User{"60000000-0000-0000-0000-000000000000", "", "mail6", "pass6", []types.ObjectID{}}}, true},
		{"i user7", args{types.User{"70000000-0000-0000-0000-000000000000", "user7", "", "pass7", []types.ObjectID{}}}, true},
		{"i user8", args{types.User{"80000000-0000-0000-0000-000000000000", "user8", "mail8", "", []types.ObjectID{}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.CreateUser(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("Database.CreateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// At this point, the following users exist in the database:
// {"10000000-0000-0000-0000-000000000000", "user1", "mail1", "pass1", []types.ObjectID{}}
// {"20000000-0000-0000-0000-000000000000", "user2", "mail2", "pass2", []types.ObjectID{}}
// {"30000000-0000-0000-0000-000000000000", "user3", "mail3", "pass3", []types.ObjectID{}}

func TestDatabase_UpdateUser(t *testing.T) {
	type args struct {
		user types.User
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"v user1", args{types.User{"10000000-0000-0000-0000-000000000000", "user1", "mail1", "pass1new", []types.ObjectID{}}}, false},
		{"v user2", args{types.User{"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}}, false},
		{"v user3", args{types.User{"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}}, false},
		{"i id", args{types.User{"01000000-0000-0000-0000-000000000000", "user4", "mail4", "pass4", []types.ObjectID{}}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.UpdateUser(tt.args.user); (err != nil) != tt.wantErr {
				t.Errorf("Database.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// At this point, the following users exist in the database:
// {"10000000-0000-0000-0000-000000000000", "user1", "mail1", "pass1new", []types.ObjectID{}}
// {"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}
// {"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}

func TestDatabase_DeleteUser(t *testing.T) {
	type args struct {
		userID types.UserID
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"v user1", args{types.UserID("10000000-0000-0000-0000-000000000000")}, false},
		{"v user1", args{types.UserID("10000000-0000-0000-0000-000000000000")}, true},
		{"v user1", args{types.UserID("")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := db.DeleteUser(tt.args.userID); (err != nil) != tt.wantErr {
				t.Errorf("Database.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// At this point, the following users exist in the database:
// {"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}
// {"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}

func TestDatabase_GetUser(t *testing.T) {
	type args struct {
		id types.UserID
	}
	tests := []struct {
		name     string
		args     args
		wantUser types.User
		wantErr  bool
	}{
		{"i user1", args{types.UserID("10000000-0000-0000-0000-000000000000")}, types.User{}, true},
		{"v user2", args{types.UserID("20000000-0000-0000-0000-000000000000")}, types.User{"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}, false},
		{"v user3", args{types.UserID("30000000-0000-0000-0000-000000000000")}, types.User{"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}, false},
		{"i user4", args{types.UserID("40000000-0000-0000-0000-000000000000")}, types.User{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := db.GetUser(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("Database.GetUser() = %v, want %v", gotUser, tt.wantUser)
			}
		})
	}
}

// At this point, the following users exist in the database:
// {"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}
// {"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}

func TestDatabase_GetUserByName(t *testing.T) {
	type args struct {
		username types.UserName
	}
	tests := []struct {
		name     string
		args     args
		wantUser types.User
		wantErr  bool
	}{
		{"i user1", args{types.UserName("user1")}, types.User{}, true},
		{"v user2", args{types.UserName("user2")}, types.User{"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}, false},
		{"v user3", args{types.UserName("user3new")}, types.User{"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}, false},
		{"i user3", args{types.UserName("user3")}, types.User{}, true},
		{"i user4", args{types.UserName("user4")}, types.User{}, true},
		{"i blank", args{types.UserName("")}, types.User{}, true},
		{"i invalid", args{types.UserName("user_1")}, types.User{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := db.GetUserByName(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.GetUserByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotUser, tt.wantUser) {
				t.Errorf("Database.GetUserByName() = %v, want %v", gotUser, tt.wantUser)
			}
		})
	}
}

// At this point, the following users exist in the database:
// {"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}
// {"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}

func TestDatabase_UserExists(t *testing.T) {
	type args struct {
		id types.UserID
	}
	tests := []struct {
		name       string
		args       args
		wantExists bool
		wantErr    bool
	}{
		{"v user1", args{types.UserID("10000000-0000-0000-0000-000000000000")}, false, false},
		{"v user2", args{types.UserID("20000000-0000-0000-0000-000000000000")}, true, false},
		{"v user3", args{types.UserID("30000000-0000-0000-0000-000000000000")}, true, false},
		{"v user4", args{types.UserID("40000000-0000-0000-0000-000000000000")}, false, false},
		{"i bad id", args{types.UserID("not a valid uuid")}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExists, err := db.UserExists(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.UserExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("Database.UserExists() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}

// At this point, the following users exist in the database:
// {"20000000-0000-0000-0000-000000000000", "user2", "mail2new", "pass2", []types.ObjectID{}}
// {"30000000-0000-0000-0000-000000000000", "user3new", "mail3", "pass3", []types.ObjectID{}}

func TestDatabase_UserExistsByName(t *testing.T) {
	type args struct {
		username types.UserName
	}
	tests := []struct {
		name       string
		args       args
		wantExists bool
		wantErr    bool
	}{
		{"v user1", args{types.UserName("user1")}, false, false},
		{"v user2", args{types.UserName("user2")}, true, false},
		{"v user3", args{types.UserName("user3new")}, true, false},
		{"v user3", args{types.UserName("user3")}, false, false},
		{"v user4", args{types.UserName("user4")}, false, false},
		{"i blank", args{types.UserName("")}, false, true},
		{"i invalid", args{types.UserName("user_1")}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExists, err := db.UserExistsByName(tt.args.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("Database.UserExistsByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotExists != tt.wantExists {
				t.Errorf("Database.UserExistsByName() = %v, want %v", gotExists, tt.wantExists)
			}
		})
	}
}
