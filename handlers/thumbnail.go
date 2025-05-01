package handlers

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/jaredfolkins/letemcook/embedded"
	"github.com/jaredfolkins/letemcook/models"
	"github.com/jaredfolkins/letemcook/util"
	"github.com/labstack/echo/v4"
	"gopkg.in/yaml.v3"
)

func GetCookbookThumbnailImage(c LemcContext) error {
	uuid := c.Param("uuid")

	cb := &models.Cookbook{}
	if err := cb.ByUUID(uuid); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to to open Cookbook")
	}

	yamlDefault := models.YamlDefault{}
	if err := yaml.Unmarshal([]byte(cb.YamlIndividual), &yamlDefault); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "Failed to unmarshal user YAML for use with thumbnail")
	}

	var encodedStr string
	if len(yamlDefault.Cookbook.Storage.Thumbnail.B64) == 0 {
		return serveDefaultThumbnail(c)
	}

	var decoded []byte
	var err error

	if len(yamlDefault.Cookbook.Storage.Thumbnail.B64) == 0 {
		decoded, err = base64.StdEncoding.DecodeString(encodedStr)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to decode base64 encoded file")
		}
	} else {
		decoded, err = base64.StdEncoding.DecodeString(yamlDefault.Cookbook.Storage.Thumbnail.B64)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to decode base64 encoded file")
		}
	}

	return c.Blob(http.StatusOK, http.DetectContentType(decoded), decoded)
}

func PostCookbookThumbnailImage(c echo.Context) error {
	uuid := c.Param("uuid")

	file, err := c.FormFile("file")
	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusBadRequest, "Failed to retrieve the file")
	}

	src, err := file.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to open the file")
	}
	defer src.Close()

	img, format, err := image.Decode(src)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "Failed to decode the image")
	}

	var buf bytes.Buffer
	if format == "gif" {
		gifImg, err := gif.DecodeAll(src)
		if err != nil {
			return c.JSON(http.StatusBadRequest, "Failed to decode the GIF image")
		}

		for i, frame := range gifImg.Image {
			croppedFrame := imaging.Fill(frame, 400, 400, imaging.Center, imaging.Lanczos)
			palettedFrame := image.NewPaletted(croppedFrame.Bounds(), frame.Palette)
			draw.Draw(palettedFrame, palettedFrame.Rect, croppedFrame, image.Point{}, draw.Over)
			gifImg.Image[i] = palettedFrame
		}

		err = gif.EncodeAll(&buf, gifImg)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to encode the GIF image")
		}
	} else {
		croppedImg := imaging.Fill(img, 400, 400, imaging.Center, imaging.Lanczos)

		switch format {
		case "jpeg":
			err = jpeg.Encode(&buf, croppedImg, &jpeg.Options{Quality: 75})
		case "png":
			encoder := png.Encoder{CompressionLevel: png.BestCompression}
			err = encoder.Encode(&buf, croppedImg)
		default:
			return c.JSON(http.StatusBadRequest, "Unsupported image format")
		}
		if err != nil {
			return c.JSON(http.StatusInternalServerError, "Failed to encode the image")
		}
	}

	if _, err = io.Copy(&buf, src); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to copy the file to buffer")
	}

	cb := &models.Cookbook{}
	if err := cb.ByUUID(uuid); err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to to open Cookbook")
	}

	yamlUser := models.YamlDefault{}
	if err := yaml.Unmarshal([]byte(cb.YamlIndividual), &yamlUser); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "Failed to unmarshal user YAML")
	}

	thumbnail := models.Thumbnail{
		B64:       base64.StdEncoding.EncodeToString(buf.Bytes()),
		Type:      format,
		Timestamp: strconv.FormatInt(time.Now().UnixNano(), 10),
	}

	yamlUser.UUID = cb.UUID
	yamlUser.Cookbook.Storage.Thumbnail = thumbnail
	prettyUserYAML, err := yaml.Marshal(yamlUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to marshal YAML")
	}
	cb.YamlIndividual = string(prettyUserYAML)

	yamlAdmin := models.YamlDefault{}
	if err := yaml.Unmarshal([]byte(cb.YamlShared), &yamlAdmin); err != nil {
		c.Logger().Error(err)
		return c.JSON(http.StatusInternalServerError, "Failed to unmarshal user YAML")
	}
	yamlAdmin.UUID = cb.UUID
	yamlAdmin.Cookbook.Storage.Thumbnail = thumbnail
	prettyAdminYAML, err := yaml.Marshal(yamlAdmin)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to marshal YAML")
	}
	cb.YamlShared = string(prettyAdminYAML)

	err = cb.Update()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, "Failed to update Cookbook")
	}

	image := &Image{
		Url: fmt.Sprintf("/lemc/cookbook/thumbnail/%s", uuid),
	}

	return c.JSON(http.StatusOK, image)
}

func GetAppThumbnail(c LemcContext) error {
	appUUID := c.Param("uuid")
	accountID := c.UserContext().ActingAs.Account.ID

	log.Printf("Attempting to fetch thumbnail for app UUID: %s, Account ID: %d", appUUID, accountID)

	app, err := models.AppByUUIDAndAccountID(appUUID, accountID)
	if err != nil {
		log.Printf("[DEBUG] Error in AppByUUIDAndAccountID: %v", err)
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("app not found (UUID: %s, Account: %d). Serving default thumbnail.", appUUID, accountID)
			return serveDefaultThumbnail(c)
		}
		log.Printf("Error fetching app (UUID: %s, Account: %d): %v. Serving default thumbnail.", appUUID, accountID, err)
		return serveDefaultThumbnail(c)
	}

	if app.YAMLIndividual == "" {
		log.Printf("App YAML is empty (AppID: %d). Serving default thumbnail.", app.ID)
		return serveDefaultThumbnail(c)
	}

	var yamlDefault models.YamlDefault
	err = yaml.Unmarshal([]byte(app.YAMLIndividual), &yamlDefault)
	if err != nil {
		log.Printf("[DEBUG] Error in yaml.Unmarshal: %v", err)
		log.Printf("Error unmarshalling app YAML (AppID: %d): %v. Serving default thumbnail.", app.ID, err)
		return serveDefaultThumbnail(c)
	}

	thumbnailDataB64 := yamlDefault.Cookbook.Storage.Thumbnail.B64
	if thumbnailDataB64 == "" {
		log.Printf("Thumbnail data is empty in app YAML (AppID: %d). Serving default thumbnail.", app.ID)
		return serveDefaultThumbnail(c)
	}

	imageData, err := base64.StdEncoding.DecodeString(thumbnailDataB64)
	if err != nil {
		log.Printf("[DEBUG] Error in base64.StdEncoding.DecodeString: %v", err)
		log.Printf("Error decoding base64 thumbnail data (AppID: %d): %v. Serving default thumbnail.", app.ID, err)
		return serveDefaultThumbnail(c)
	}

	contentType := http.DetectContentType(imageData)
	if !strings.HasPrefix(contentType, "image/") {
		log.Printf("[DEBUG] Decoded data not image type: %s", contentType)
		log.Printf("Decoded thumbnail data is not a valid image type ('%s', AppID: %d). Serving default thumbnail.", contentType, app.ID)
		return serveDefaultThumbnail(c)
	}

	log.Printf("Serving actual thumbnail for app %s (AppID: %d, Type: %s)", appUUID, app.ID, contentType)
	c.Response().Header().Set("Content-Type", contentType)
	return c.Blob(http.StatusOK, contentType, imageData)
}

func serveDefaultThumbnail(c LemcContext) error {
	themeName := c.Theme()
	if themeName == "" {
		themeName = util.DefaultTheme
	}
	filePath := fmt.Sprintf("themes/%s/public/imgs/placeholder.png", themeName)

	placeholderBytes, err := embedded.ReadAsset(filePath)
	if err != nil {
		log.Printf("[DEBUG] Error reading embedded asset %s: %v", filePath, err)
		log.Printf("Error reading default thumbnail file '%s' from embedded FS: %v", filePath, err)
		return c.String(http.StatusInternalServerError, "Error serving default image: placeholder missing or unreadable")
	}

	contentType := "image/png"
	c.Response().Header().Set("Content-Type", contentType)
	return c.Blob(http.StatusOK, contentType, placeholderBytes)
}
