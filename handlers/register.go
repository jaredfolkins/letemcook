package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/views/pages"
	"golang.org/x/crypto/bcrypt"
)

func RegisterHandler(c LemcContext) error {
	partial := c.QueryParam("partial")
	squid := c.QueryParam("squid")
	// accountName := c.QueryParam("account") // Removed: declared and not used

	account, err := models.AccountBySquid(squid)
	if err != nil {
		log.Printf("Error finding account by squid '%s': %v", squid, err)
		return c.String(http.StatusNotFound, "Account not found")
	}

	newsquid, name, err := models.SquidAndNameByAccountID(account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	bv := NewBaseViewWithSquidAndAccountName(c, newsquid, name)

	// Check if registration is enabled
	if !bv.RegistrationEnabled {
		c.AddErrorFlash("registration_disabled", "User registration is currently disabled for this account.")
		loginRedirectURL := fmt.Sprintf("%s?squid=%s&account=%s", paths.Login, newsquid, name)
		// Redirect immediately, no need to prepare RegisterView
		return c.Redirect(http.StatusSeeOther, loginRedirectURL)
	}

	dv := models.RegisterView{
		BaseView: bv,
	}

	registerView := pages.Register(dv)
	if partial == "true" {
		return HTML(c, registerView)
	}
	ri := pages.RegisterIndex(dv, registerView)
	return HTML(c, ri)
}

func PostRegisterHandler(c LemcContext) error {
	squid := c.QueryParam("squid")
	// accountName := c.QueryParam("account") // Removed: declared and not used

	account, err := models.AccountBySquid(squid)
	if err != nil {
		log.Printf("Invalid squid '%s' submitted during registration: %v", squid, err)
		return c.String(http.StatusNotFound, "Account not found")
	}

	newSquid, newName, err := models.SquidAndNameByAccountID(account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	validate := validator.New()

	user := &models.User{
		Email:    c.FormValue("email"),
		Username: c.FormValue("username"),
		Password: c.FormValue("password"),
	}

	if err := validate.Struct(user); err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			log.Println(err.StructNamespace())
			log.Println(err.Tag())
			log.Println(err.Param())
			log.Println(err.Value())
		}
	}

	tx, err := db.Db().Beginx()
	if err != nil {
		return err
	}

	dv := models.RegisterView{BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName)}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Hash = string(hashedPassword)

	// Create user and get ID
	userID, err := models.CreateUserWithAccountID(user, account.ID, tx)

	if err != nil {
		err0 := tx.Rollback()
		if err0 != nil {
			log.Println(err0)
		}

		c.AddErrorFlash("register0", "Failed to create user")
		c.AddErrorFlash("register1", err.Error())
		registerView := pages.Register(dv)
		return HTML(c, registerView)
	}

	// Assign user to apps with on_register set
	appsToAssign, err := models.GetAppsForRegistrationByAccountID(tx, account.ID)
	if err != nil {
		log.Printf("Error fetching apps to assign for user %d, account %d: %v", userID, account.ID, err)
		err0 := tx.Rollback()
		if err0 != nil {
			log.Println("Rollback error after failing to fetch apps: ", err0)
		}
		c.AddErrorFlash("register0", "Failed during registration process (assigning apps)")
		registerView := pages.Register(dv)
		return HTML(c, registerView)
	}

	for _, app := range appsToAssign {
		err = models.AssignAppPermissionOnRegister(tx, userID, app.AccountID, app.ID, app.CookbookID)
		if err != nil {
			log.Printf("Error assigning app %d to user %d via model function: %v", app.ID, userID, err)
			terr := tx.Rollback()
			if terr != nil {
				log.Println("Rollback error after failing AssignAppPermissionOnRegister: ", terr)
			}
			c.AddErrorFlash("register0", "Failed during registration process (assigning app permissions)")
			registerView := pages.Register(dv)
			return HTML(c, registerView)
		}
	}

	err = tx.Commit()
	if err != nil {
		// Log commit error, but the user might already be created
		log.Printf("Error committing transaction after user creation and app assignment: %v", err)
		c.AddErrorFlash("register0", "Registration completed but encountered an issue finalizing.")
		// Proceed to login page anyway, as user likely exists
	}

	// Existing code to fetch user and redirect
	_, err = models.ByUsernameAndAccountID(user.Username, account.ID)
	if err != nil {
		// User was created but fetching failed? Log and redirect to login anyway.
		log.Printf("Error fetching newly created user %s for account %d after successful registration: %v", user.Username, account.ID, err)
	}

	c.Response().Header().Set("HX-Replace-Url", fmt.Sprintf("%s?squid=%s&account=%s", paths.Login, newSquid, newName))
	lv := models.LoginView{BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName)}
	c.AddSuccessFlash("register", "Your account has been created")
	loginView := pages.Login(lv)
	return HTML(c, loginView)
}
