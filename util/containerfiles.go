package util

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	PRIVATE              = "private"
	PUBLIC               = "public"
	CACHE                = "cache"
	GLOBAL_DIR           = "global"
	SCOPE_INDIVIDUAL_DIR = "individual"
	SCOPE_SHARED_DIR     = "shared"

	CACHE_HTML = "cache.html"
	CACHE_CSS  = "cache.css"
	CACHE_JS   = "cache.js"
)

type ContainerFiles struct {
	Html *os.File
	Css  *os.File
	Js   *os.File

	InternalPerUserPublicDir  string
	InternalPerUserPrivateDir string
	InternalPerUserCacheDir   string
	InternalSharedDir         string
	InternalGlobalDir         string

	BindPerUserPublicDir  string
	BindPerUserPrivateDir string
	BindPerUserCacheDir   string
	BindSharedDir         string
	BindGlobalDir         string
}

func NewContainerFiles(jm *JobMeta, is_admin bool) (*ContainerFiles, error) {
	cf := &ContainerFiles{}

	basedir := SCOPE_INDIVIDUAL_DIR
	if is_admin {
		basedir = SCOPE_SHARED_DIR
	}

	fm, err := NewFileMeta(jm, is_admin)
	if err != nil {
		return cf, err
	}

	// set the bind directories to the host data path
	// so when lemc interacts with unix:///var/run/docker.sock
	// it can read and write to the host data path
	if os.Getenv("LEMC_HOST_LOCKER_PATH") != "" {
		cf.BindPerUserPublicDir = filepath.Join(os.Getenv("LEMC_HOST_LOCKER_PATH"), jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, PUBLIC)
		cf.BindPerUserPrivateDir = filepath.Join(os.Getenv("LEMC_HOST_LOCKER_PATH"), jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, PRIVATE)
		cf.BindPerUserCacheDir = filepath.Join(os.Getenv("LEMC_HOST_LOCKER_PATH"), jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, CACHE)
		cf.BindSharedDir = filepath.Join(os.Getenv("LEMC_HOST_LOCKER_PATH"), jm.UUID, SCOPE_SHARED_DIR)
		cf.BindGlobalDir = filepath.Join(os.Getenv("LEMC_HOST_LOCKER_PATH"), jm.UUID, GLOBAL_DIR)
	} else {
		cf.BindPerUserPublicDir = filepath.Join(fm.LemcLocker, jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, PUBLIC)
		cf.BindPerUserPrivateDir = filepath.Join(fm.LemcLocker, jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, PRIVATE)
		cf.BindPerUserCacheDir = filepath.Join(fm.LemcLocker, jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, CACHE)
		cf.BindSharedDir = filepath.Join(fm.LemcLocker, jm.UUID, SCOPE_SHARED_DIR)
		cf.BindGlobalDir = filepath.Join(fm.LemcLocker, jm.UUID, GLOBAL_DIR)
	}

	cf.InternalPerUserPublicDir = filepath.Join(fm.LemcLocker, jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, PUBLIC)
	if _, err := os.Stat(cf.InternalPerUserPublicDir); os.IsNotExist(err) {
		err := os.MkdirAll(cf.InternalPerUserPublicDir, 0755)
		if err != nil {
			return cf, err
		}
	}

	cf.InternalPerUserPrivateDir = filepath.Join(fm.LemcLocker, jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, PRIVATE)
	if _, err := os.Stat(cf.InternalPerUserPrivateDir); os.IsNotExist(err) {
		err := os.MkdirAll(cf.InternalPerUserPrivateDir, 0755)
		if err != nil {
			return cf, err
		}
	}

	cf.InternalPerUserCacheDir = filepath.Join(fm.LemcLocker, jm.UUID, basedir, fm.IndividualUsernameOrSharedUsername, fm.PageString, CACHE)
	if _, err := os.Stat(cf.InternalPerUserCacheDir); os.IsNotExist(err) {
		err := os.MkdirAll(cf.InternalPerUserCacheDir, 0755)
		if err != nil {
			return cf, err
		}
	}

	cf.InternalSharedDir = filepath.Join(fm.LemcLocker, jm.UUID, SCOPE_SHARED_DIR)
	if _, err := os.Stat(cf.InternalSharedDir); os.IsNotExist(err) {
		err := os.MkdirAll(cf.InternalSharedDir, 0755)
		if err != nil {
			return cf, err
		}
	}

	cf.InternalGlobalDir = filepath.Join(fm.LemcLocker, jm.UUID, GLOBAL_DIR)
	if _, err := os.Stat(cf.InternalGlobalDir); os.IsNotExist(err) {
		err := os.MkdirAll(cf.InternalGlobalDir, 0755)
		if err != nil {
			return cf, err
		}
	}

	return cf, nil
}

func (cf *ContainerFiles) OpenFiles() error {
	var err error
	cf.Html, err = os.OpenFile(filepath.Join(cf.InternalPerUserCacheDir, CACHE_HTML), os.O_RDWR|os.O_CREATE|os.O_SYNC, 0755)
	if err != nil {
		return err
	}

	cf.Css, err = os.OpenFile(filepath.Join(cf.InternalPerUserCacheDir, CACHE_CSS), os.O_RDWR|os.O_CREATE|os.O_SYNC, 0755)
	if err != nil {
		return err
	}

	cf.Js, err = os.OpenFile(filepath.Join(cf.InternalPerUserCacheDir, CACHE_JS), os.O_RDWR|os.O_CREATE|os.O_SYNC, 0755)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ContainerFiles) CloseFiles() {
	cf.Html.Close()
	cf.Css.Close()
	cf.Js.Close()
}

func (cf *ContainerFiles) Append(msg string, file *os.File) error {
	/*
		origin, err := cf.Read(file)
		if err != nil {
			return err
		}

		s := fmt.Sprintf("%s\n%s\n", origin, msg)

	*/

	_, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = file.WriteString(msg + "\n")
	if err != nil {
		return err
	}
	return nil
}

func (cf *ContainerFiles) Trunc(msg string, file *os.File) error {
	err := file.Truncate(0)
	if err != nil {
		return err
	}

	_, err = file.WriteAt([]byte(msg), 0)
	if err != nil {
		return err
	}

	return nil
}

func (cf *ContainerFiles) Read(file *os.File) (string, error) {
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
