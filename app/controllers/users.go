package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/hex"
	"github.com/huntaub/list/app/routes"
	"github.com/robfig/revel"
	"labix.org/v2/mgo"
	"strings"
)

type User struct {
	Email          string
	HashedPassword string
	FullName       string
	ClassBucket    []string
}

var users *mgo.Collection

func init() {
	session, _ := mgo.Dial("mongodb://leath:hunter0813@oceanic.mongohq.com:10000/list")
	users = session.DB("list").C("users")
}

type Users struct {
	*revel.Controller
}

func (u *Users) Login(email string, password string) revel.Result {
	var user *User
	err := users.Find(map[string]string{"email": email}).One(&user)
	if err != nil {
		u.Flash.Error("Incorrect username or password.")
		return u.Redirect(routes.App.Index())
	}

	bytes, _ := hex.DecodeString(user.HashedPassword)
	if bcrypt.CompareHashAndPassword(bytes, []byte(password)) != nil {
		u.Flash.Error("Incorrect username or password.")
		return u.Redirect(routes.App.Index())
	}

	u.Session["user"] = email

	return u.Redirect(routes.App.Index())
}

func (u *Users) Register(name string, email string, password string, cpassword string) revel.Result {
	if name == "" {
		u.Flash.Error("Name is required.")
		return u.Redirect(routes.App.Index())
	}
	if email == "" {
		u.Flash.Error("Email is required.")
		return u.Redirect(routes.App.Index())
	}
	if !strings.HasSuffix(email, "@virginia.edu") {
		u.Flash.Error("Only @virginia.edu emails allowed.")
		return u.Redirect(routes.App.Index())
	}
	if password == "" {
		u.Flash.Error("Password is required.")
		return u.Redirect(routes.App.Index())
	}
	if password != cpassword {
		u.Flash.Error("Passwords do not match.")
		return u.Redirect(routes.App.Index())
	}

	var existing *User
	if users.Find(map[string]string{"email": email}).One(&existing) == nil {
		u.Flash.Error("Someone with that email has already registered.")
		return u.Redirect(routes.App.Index())
	}

	pass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	users.Insert(&User{
		Email:          email,
		FullName:       name,
		HashedPassword: hex.EncodeToString(pass),
	})

	u.Session["user"] = email

	return u.Redirect(routes.App.Index())
}

func (u *Users) Logout() revel.Result {
	delete(u.Session, "user")
	return u.Redirect(routes.App.Index())
}
