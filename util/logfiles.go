package util

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type LogFile struct {
	EventID string
	File    *os.File
	Writer  *bufio.Writer
	Dir     string
}

const (
	LOGS                      = "logs"
	SHARED_SINGLETON_USERNAME = "shared-user"
)

type FileMeta struct {
	Scope                              string
	SafeUsername                       string
	PageString                         string
	LemcLocker                         string
	IndividualUsernameOrSharedUsername string
	ImageName                          string
	UUID                               string
}

const SCOPE_SHARED = "shared"
const SCOPE_INDIVIDUAL = "individual"

func (fm *FileMeta) ConainterName() string {
	return fmt.Sprintf("uuid-%s-page-%s-scope-%s-username-%s-imagename-%s-", fm.UUID, fm.PageString, fm.Scope, fm.IndividualUsernameOrSharedUsername, fm.ImageName)
}

func NewFileMeta(jm *JobMeta, is_admin bool) (*FileMeta, error) {
	fm := &FileMeta{}

	fm.SafeUsername = fmt.Sprintf("%v-%v", AlphaNumHyphen(jm.Username), jm.UserID)
	fm.PageString = fmt.Sprintf("page-%v", jm.PageID)
	fm.LemcLocker = LockerPath()
	fm.UUID = jm.UUID

	if is_admin {
		fm.Scope = SCOPE_SHARED
		fm.IndividualUsernameOrSharedUsername = SHARED_SINGLETON_USERNAME
	} else {
		fm.Scope = SCOPE_INDIVIDUAL
		fm.IndividualUsernameOrSharedUsername = fm.SafeUsername
	}

	return fm, nil
}

func (lf *LogFile) StepWriteToLog(stepid, msg, imageHash, imageName string) {
	timestamp := time.Now().Format("Mon Jan 2 15:04:05 MST 2006")
	lf.Writer.WriteString(fmt.Sprintf("[%s] [image:%s] [name:%s] [event:%s] [step:%s] %s\n", timestamp, imageHash, imageName, lf.EventID, stepid, msg))
}

func (fm *FileMeta) OpenLogFile(jm *JobMeta) (*LogFile, error) {
	var err error
	lf := &LogFile{
		EventID: fmt.Sprintf("%d", time.Now().Unix()),
	}

	lf.Dir = filepath.Join(fm.LemcLocker, jm.UUID, fm.Scope, fm.IndividualUsernameOrSharedUsername, LOGS)
	if _, err := os.Stat(lf.Dir); os.IsNotExist(err) {
		err := os.MkdirAll(lf.Dir, DirPerm)
		if err != nil {
			return lf, err
		}
	}

	fileName := fmt.Sprintf("%s-%s.log", AlphaNumHyphen(fm.PageString), AlphaNumHyphen(jm.RecipeName))
	file, err := os.OpenFile(filepath.Join(lf.Dir, fileName), os.O_RDWR|os.O_CREATE|os.O_APPEND, FilePerm)
	if err != nil {
		return lf, err
	}
	lf.File = file
	lf.Writer = bufio.NewWriter(file)
	return lf, nil
}

func (lf *LogFile) CloseLogFile() {
	if lf.Writer != nil {
		lf.Writer.Flush()
	}
	if lf.File != nil {
		lf.File.Close()
	}
}
