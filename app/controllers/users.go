package controllers

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/hex"
	mailer "github.com/huntaub/list/app/email"
	"github.com/huntaub/list/app/models"
	"github.com/huntaub/list/app/routes"
	"github.com/revel/revel"
	"math/rand"
	"strings"
)

const api_chars = "abcdefghijklmnopqrstuvwxyz0123456789"

const api_key_length = 15
const email_verification_length = 30

// Generate A String of a Random Length
func generateKey(length int) string {
	r := make([]string, length)
	for i := 0; i < length; i++ {
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

// Verify User Email
func (u *Users) VerifyEmail(verification string, email string) revel.Result {
	// Grab the User Verifying
	var user *models.User
	err := users.Find(map[string]string{"email": email, "verificationkey": verification}).One(&user)
	if err != nil {
		u.Flash.Error("Incorrect verification key.")
		return u.Redirect(routes.App.Index())
	}

	// Update the User - They are Verified
	user.Verified = true
	err = users.Update(map[string]string{"email": email}, user)
	if err != nil {
		u.Flash.Error("Unable to verify you at this time.")
		return u.Redirect(routes.App.Index())
	}

	// Log them in
	u.Session["user"] = email

	// Show success
	u.Flash.Success("Email Successfully verified.")
	return u.Redirect(routes.App.Index())
}

// Login a User
func (u *Users) Login(email string, password string) revel.Result {
	// Grab User with Email
	var user *models.User
	err := users.Find(map[string]string{"email": email}).One(&user)
	if err != nil {
		u.Flash.Error("Incorrect username or password.")
		return u.Redirect(routes.App.Index())
	}

	// Check Passwords
	bytes, _ := hex.DecodeString(user.HashedPassword)
	if bcrypt.CompareHashAndPassword(bytes, []byte(password)) != nil {
		u.Flash.Error("Incorrect username or password.")
		return u.Redirect(routes.App.Index())
	}

	// Only login if they are verified
	if user.Verified {
		u.Session["user"] = email
	} else {
		u.Flash.Error("You cannot login until you verify your email.")
	}

	return u.Redirect(routes.App.Index())
}

// Register a User
func (u *Users) Register(name string, email string, password string, cpassword string) revel.Result {
	// Ensure that all required fields are provided
	if name == "" {
		u.Flash.Error("Name is required.")
		return u.Redirect(routes.App.Index())
	}
	if email == "" {
		u.Flash.Error("Email is required.")
		return u.Redirect(routes.App.Index())
	}
	if password == "" {
		u.Flash.Error("Password is required.")
		return u.Redirect(routes.App.Index())
	}
	// Ensure that email is UVa
	if !strings.HasSuffix(email, "@virginia.edu") {
		u.Flash.Error("Only @virginia.edu emails allowed.")
		return u.Redirect(routes.App.Index())
	}
	// Ensure that passwords match
	if password != cpassword {
		u.Flash.Error("Passwords do not match.")
		return u.Redirect(routes.App.Index())
	}

	// Ensure that Email is not already registered
	var existing *models.User
	if users.Find(map[string]string{"email": email}).One(&existing) == nil {
		u.Flash.Error("Someone with that email has already registered.")
		return u.Redirect(routes.App.Index())
	}

	// Generate Unique API Key
tryAgain:
	testKey := generateKey(api_key_length)
	if users.Find(map[string]string{"apikey": testKey}).One(&existing) == nil {
		goto tryAgain
	}

	// Generate Password Hash
	pass, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	// Create User Object
	newUser := &models.User{
		Email:           email,
		FullName:        name,
		HashedPassword:  hex.EncodeToString(pass),
		APIKey:          testKey,
		Verified:        false,
		VerificationKey: generateKey(email_verification_length),
	}

	// Send Verification Email
	err := mailer.SendVerificationEmail(newUser, revel.Config.StringDefault("server.baseurl", ""))
	if err != nil {
		u.Flash.Error("Unable to send verification email. Try registering again.")
		revel.ERROR.Printf("%v", err)
		return u.Redirect(routes.App.Index())
	}

	// Insert New User into Database
	users.Insert(newUser)

	// Success! Redirect
	u.Flash.Success("You have successfully registered. Now, we must verify your email. Click on the link that was just sent to you.")
	return u.Redirect(routes.App.Index())
}

// Log User Out of Application
func (u *Users) Logout() revel.Result {
	delete(u.Session, "user")
	return u.Redirect(routes.App.Index())
}
