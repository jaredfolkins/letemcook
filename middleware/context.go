package middleware

import (
	"encoding/json"
	"io/fs"
	"log"

	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/labstack/echo/v4"
)

const (
	X_LEMC_FLASH_ERROR   = "X-Lemc-Flash-Error"
	X_LEMC_FLASH_SUCCESS = "X-Lemc-Flash-Success"
)

// LemcContext is the interface for our custom context
type LemcContext interface {
	echo.Context
	UserContext() *models.UserContext
	AddErrorFlash(key, value string)
	AddSuccessFlash(key, value string)
	Theme() string
	CacheBuster() string
	AssetsFS() (fs.FS, error)
}

// NewCustomContext creates a new custom context
func NewCustomContext(c echo.Context) *lemcContext {
	return &lemcContext{
		Context:        c,
		userContext:    &models.UserContext{},
		errorFlashes:   make(map[string]string),
		successFlashes: make(map[string]string),
	}
}

// lemcContext is our concrete implementation of LemcContext
type lemcContext struct {
	echo.Context
	userContext    *models.UserContext
	errorFlashes   map[string]string
	successFlashes map[string]string
	theme          string
	cacheBuster    string
}

// UserContext returns the user context
func (c *lemcContext) UserContext() *models.UserContext {
	return c.userContext
}

// AddErrorFlash adds an error flash message
func (c *lemcContext) AddErrorFlash(key string, value string) {
	s := c.Response().Header().Get(X_LEMC_FLASH_ERROR)
	m := make(map[string]string)
	if len(s) > 0 {
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			log.Println("error: ", err)
			return
		}
	}

	m[key] = value
	jsonData, err := json.Marshal(m)
	if err != nil {
		log.Println("error: ", err)
		return
	}

	c.Response().Header().Set(X_LEMC_FLASH_ERROR, string(jsonData))
}

// AddSuccessFlash adds a success flash message
func (c *lemcContext) AddSuccessFlash(key string, value string) {
	s := c.Response().Header().Get(X_LEMC_FLASH_SUCCESS)
	m := make(map[string]string)
	if len(s) > 0 {
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			log.Println("error: ", err)
			return
		}
	}

	m[key] = value
	jsonData, err := json.Marshal(m)
	if err != nil {
		log.Println("error: ", err)
		return
	}

	c.Response().Header().Set(X_LEMC_FLASH_SUCCESS, string(jsonData))
}

// Theme returns the current theme
func (c *lemcContext) Theme() string {
	return c.theme
}

// CacheBuster returns the cache buster string
func (c *lemcContext) CacheBuster() string {
	return c.cacheBuster
}

// AssetsFS returns the embedded filesystem for assets
func (c *lemcContext) AssetsFS() (fs.FS, error) {
	return embedded.GetAssetsFS()
}
