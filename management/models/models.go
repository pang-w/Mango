package models

import "time"

// User definition
type User struct {
	Id         int       `orm:"auto;index"`
	Email      string    `orm:"size(64);index;unique"`
	Password   string    `orm:"size(128)"`
	Name       string    `orm:"size(64);index"`
	Nickname   string    `orm:"size(64);unique;index"`
	LastLogin  time.Time `orm:"auto_now_add"`
	DateJoined time.Time `orm:"auto_now_add"`
	IsActive   bool      `orm:"default(1)"`
	IsAdmin    bool      `orm:"default(0);index"`
    Profile  *UserProfile `orm:"reverse(one)"`
    Permissions []*Permission `orm:"rel(m2m);rel_table(user_permission)"`
    PasswordPermissions []*PasswordPermission `orm:"reverse(many)"`
    //Passwords []*PasswordInfo `orm:"rel(m2m)"`
}

// User Additional info
type UserProfile struct {
	Id         int    `orm:"auto;index"`
	Department string `orm:"index"`
	Title      string `orm:"null"`
	Mobile     string `orm:"index"`
	Phone      string `orm:"null;index"`
	User       *User  `orm:"rel(one)"`
	Salt       string
}

// Register invitation
type RegisterInvitation struct {
	Id        int `orm:"auto"`
	Token     string
	Email     string
	Expired   bool
	IssueDate time.Time     `orm:"auto_now_add"`
}

type Permission struct {
    Id int `orm:"auto"`
    ContentTypeId int
    Name string
    Codename string
    Users []*User `orm:"reverse(many)"`
}


type PasswordInfo struct {
    Id int `orm:"auto;index"`
    Account string
    Password string
    Desc string `orm:"null"`
    Permissions []*PasswordPermission `orm:"reverse(many)"`
    //Users []*User `orm:"reverse(many)"`
}

type PasswordPermission struct {
    Id int `orm:"auto;index`
    Password *PasswordInfo `orm:"rel(fk)"`
    User *User `orm:"rel(fk)"`
    Level int
}

func (this *PasswordPermission) TableUnique() [][]string {
    return [][]string {
        []string{"Password", "User"},
    }
}
