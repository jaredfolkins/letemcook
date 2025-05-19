package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/jaredfolkins/letemcook/db"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/paths"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/jaredfolkins/letemcook/views/pages"
	"golang.org/x/crypto/bcrypt"
)

func SetupHandler(c LemcContext) error {
	bv := NewBaseView(c)
	bv.IsSetup = true
	sv := models.SetupView{
		BaseView: bv,
	}
	setupView := pages.Setup(sv)
	si := pages.SetupIndex(sv, setupView)
	return HTML(c, si)
}

type FormSetup struct {
	SiteName string `json:"Site Name"`
	Email    string `json:"Email"`
	Username string `json:"Username"`
	Password string `json:"Password"`
}

func (f FormSetup) Validate() error {
	pattern := `^[a-zA-Z0-9-]+$`
	return validation.ValidateStruct(&f,
		validation.Field(&f.SiteName, validation.Required, validation.Length(3, 64), validation.Match(regexp.MustCompile(pattern)).Error("can only contain alphanumeric characters and dashes")),
		validation.Field(&f.Email, validation.Required, is.Email),
		validation.Field(&f.Username, validation.Required, validation.Length(3, 64), validation.Match(regexp.MustCompile(pattern)).Error("can only contain alphanumeric characters and dashes")),
		validation.Field(&f.Password, validation.Required, validation.Length(12, 64)),
	)
}

func PostSetupHandler(c LemcContext) error {
	form := FormSetup{
		SiteName: util.Sanitize(c.FormValue("site_name")),
		Email:    util.Sanitize(c.FormValue("email")),
		Username: util.Sanitize(c.FormValue("username")),
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

	tx, err := db.Db().Beginx()
	if err != nil {
		return err
	}

	account, err := models.AccountCreate(form.SiteName, tx)
	if err != nil {
		return err
	}

	user := &models.User{
		Email:    form.Email,
		Username: form.Username,
		Password: form.Password,
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Hash = string(hashedPassword)

	err = models.CreateSuperUserWithAccountID(user, account.ID, tx)
	if err != nil {
		err0 := tx.Rollback()
		if err0 != nil {
			log.Println(err0)
		}

		c.AddErrorFlash("register0", "Failed to create user")
		c.AddErrorFlash("register1", err.Error())
		return c.NoContent(http.StatusNoContent)
	}

	err = tx.Commit()
	if err != nil {
		c.AddErrorFlash("setup", err.Error())
		return c.NoContent(http.StatusNoContent)
	}

	_, err = models.ByUsernameAndAccountID(user.Username, account.ID)
	if err != nil {
		return err
	}

	newSquid, newName, err := util.SquidAndNameByAccountID(account.ID)
	if err != nil {
		return c.String(http.StatusNotFound, "Can't create Squid")
	}

	c.Response().Header().Set("HX-Replace-Url", fmt.Sprintf("%s?squid=%s&account=%s", paths.Login, newSquid, newName))
	lv := models.LoginView{BaseView: NewBaseViewWithSquidAndAccountName(c, newSquid, newName)}
	c.AddSuccessFlash("setup", "You're all setup and ready to cook!")
	loginView := pages.Login(lv)
	return HTML(c, loginView)
}
