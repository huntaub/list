package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/hex"
	"github.com/huntaub/list/app/routes"
	"github.com/robfig/revel"
	"math/rand"
	"strings"
)

type User struct {
	Email          string
	HashedPassword string
	FullName       string
	ClassBucket    []string
	APIKey         string
}

const api_chars = "abcdefghijklmnopqrstuvwxyz0123456789"

func generateKey() string {
	r := make([]string, 15)
	for i := 0; i < 15; i++ {
		l := rand.Intn(len(api_chars))
		if i != 0 && r[i-1] == string(api_chars[l]) {
			i--
			continue
		}
		r[i] = string(api_chars[l])
	}
	return strings.Join(r, "")
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

tryAgain:
	testKey := generateKey()
	if users.Find(map[string]string{"apikey": testKey}).One(&existing) == nil {
		goto tryAgain
	}

	pass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	users.Insert(&User{
		Email:          email,
		FullName:       name,
		HashedPassword: hex.EncodeToString(pass),
		APIKey:         testKey,
	})

	u.Session["user"] = email

	return u.Redirect(routes.App.Index())
}

func (u *Users) Logout() revel.Result {
	delete(u.Session, "user")
	return u.Redirect(routes.App.Index())
}
