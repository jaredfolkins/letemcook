package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"github.com/labstack/echo-contrib/session"
	"golang.org/x/crypto/bcrypt"
)

func LoginHandler(c LemcContext) error {
	lv, err := newLoginView(c)
	if err != nil {
		log.Println("newLoginView: ", err)
	}

	partial := c.QueryParam("partial")
	if partial != "true" && partial != "false" {
		log.Printf("Invalid 'partial' query param value received: %q. Defaulting to 'false'.", partial)
		partial = "false"
	}
	return renderLoginView(c, lv, partial)
}

type FormLogin struct {
	Username string `json:"username"`
	Password string `json:"Password"`
}

func (f FormLogin) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.Username, validation.Required, validation.Length(3, 64)),
		validation.Field(&f.Password, validation.Required, validation.Length(12, 64)),
	)
}

func PostLoginHandler(c LemcContext) error {

	form := &FormLogin{
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	if err := form.Validate(); err != nil {
		var errs validation.Errors
		if errors.As(err, &errs) {
			for field, errMsg := range errs {
				msg := fmt.Sprintf("%s: %s", field, errMsg)
				c.AddErrorFlash(field, msg)
			}
		} else {
			c.AddErrorFlash("setup", err.Error())
		}
		return c.NoContent(http.StatusNoContent)
	}

	user, err := models.ByUsernameAndSquid(form.Username, c.QueryParam("squid"))
	if err != nil {
		log.Println("ByUsernameAndSquid: ", err)
		c.AddErrorFlash("login", "Invalid username or password.")
		return c.NoContent(http.StatusNoContent)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Hash), []byte(form.Password))
	if err != nil {
		log.Println("CompareHashAndPassword: ", err)
		c.AddErrorFlash("login", "Invalid username or password.")
		return c.NoContent(http.StatusNoContent)
	}

	sess, err := session.Get("session", c)
	if err != nil {
		c.AddErrorFlash("login", "Session error.")
		return c.NoContent(http.StatusNoContent)
	}

	lv, err := newLoginView(c)
	if err != nil {
		log.Println("newLoginView: ", err)
		c.AddErrorFlash("login", "View initialization error.")
		return renderLoginView(c, lv, "true")
	}

	page := 1
	limit := 10 // Default limit

	totalapps, err := models.Countapps(user.ID, user.Account.ID)
	if err != nil {
		log.Printf("Error counting apps post-login for user %d, account %d: %v", user.ID, user.Account.ID, err)
		totalapps = 0 // Fallback to 0 if count fails
	}

	totalPages := 0
	if totalapps > 0 {
		totalPages = int(math.Ceil(float64(totalapps) / float64(limit)))
	}

	its, err := models.Apps(user.ID, user.Account.ID, page, limit)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("Error fetching apps post-login: %v", err)
			c.AddErrorFlash("login", "Error loading apps after login.")
			return c.NoContent(http.StatusNoContent)
		}
	}

	// We can only update the user context in the session,
	// so we rely on the session handling in Before middleware
	sess.Values["logged_in_user_id"] = user.ID
	sess.Values["logged_in_account_id"] = user.Account.ID
	sess.Values["acting_as_user_id"] = user.ID            // Initially the same
	sess.Values["acting_as_account_id"] = user.Account.ID // Initially the same
	sess.Options.MaxAge = 86400 * 7                       // Example: 7 days

	err = sess.Save(c.Request(), c.Response())
	if err != nil {
		log.Printf("Error saving session: %v", err) // Log the specific error
		return err
	}

	// Create a user context with the newly logged-in user
	userCtx := &models.UserContext{
		LoggedInAs: user,
		ActingAs:   user, // Initially, acting as self
	}

	newSquid, newName, err := util.SquidAndNameByAccountID(user.Account.ID)
	if err != nil {
		return c.String(http.StatusInternalServerError, "Failed to generate account identifier")
	}
	// Manually construct BaseView instead of using NewBaseViewWithSquidAndAccountName
	// This ensures the UserContext is populated correctly for *this* request.
	baseViewForResponse := models.BaseView{
		AccountSquid: newSquid,
		AccountName:  newName,
		Theme:        c.Theme(),       // Get theme from context
		CacheBuster:  c.CacheBuster(), // Get cache buster from context
		UserContext:  userCtx,         // Use the manually created UserContext
		ActiveNav:    "",
		// TODO: Potentially set IsDev, RegistrationEnabled, nav flags if needed immediately
	}

	v := models.AppsView{
		Apps:        its,
		BaseView:    baseViewForResponse,
		CurrentPage: page,
		TotalPages:  totalPages,
		Limit:       limit,
	}

	c.Response().Header().Set("HX-Replace-Url", "/lemc/apps")
	c.AddSuccessFlash("login", "Login successful.")

	cv := pages.Apps(v)
	return HTML(c, cv)

}
