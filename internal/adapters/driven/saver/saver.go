// Saver adapter.
// Saves data to disk
package saver

import (
	"fmt"
	"os"
	"sync"

	"github.com/vynovikov/postSaver/internal/repo"

	json "github.com/goccy/go-json"
)

type Saver interface {
	FileCreate(repo.Request) (string, error)
	FileWrite(repo.Request) error
	FileClose(repo.Request) error
	TableSave(map[string]repo.NameNumber, string) error
	tableSave(map[string]string, string) error
}

type SaverStruct struct {
	Path string
	F    map[string]*os.File
	l    sync.Mutex
}

func NewSaver(path string) (*SaverStruct, error) {
	f := make(map[string]*os.File)
	_, err := os.Stat(path)

	if err != nil {

		if os.IsNotExist(err) {

			os.Mkdir(path, 0777)

			return &SaverStruct{Path: path, F: f}, nil
		}
		return &SaverStruct{}, err
	}
	return &SaverStruct{Path: path, F: f}, nil
}

// FileCreate returns path of file to save request data.
// Creates file if necessary.
// Tested in saver_test.go
func (sv *SaverStruct) FileCreate(r repo.Request) (string, error) {
	sv.l.Lock()
	defer sv.l.Unlock()

	_, err := os.Stat(sv.Path)

	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(sv.Path, 0777)
			if err != nil {
				return "", fmt.Errorf("in saver.FileCreate unable to create folder %q: %v", sv.Path, err)
			}
		} else {
			return "", fmt.Errorf("in saver.FileCreate error while finding folder %q: %v", sv.Path, err)
		}

	}

	folderPath := sv.Path + "/" + r.TS()

	_, err = os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(folderPath, 0777)
			if err != nil {
				return "", fmt.Errorf("in saver.FileCreate unable to create folder %q: %v", folderPath, err)
			}
		} else {
			return "", fmt.Errorf("in saver.FileCreate error while finding folder %q: %v", folderPath, err)
		}

	}
	filePath := folderPath + "/" + r.FileName()
	f, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("in saver.FileCreate creating file %q failed: %v\n", filePath, err)
	}

	sv.F[r.Name()] = f
	return filePath, nil

}

// Tested in saver_test.go
func (sv *SaverStruct) FileWrite(r repo.Request) error {
	f := sv.F[r.Name()]
	_, err := f.Write(r.GetBody())
	if err != nil {
		return fmt.Errorf("in saver.FileWrite wiriting to file associated with form %q failed: %v\n", r.Name(), err)
	}
	return nil
}

func (sv *SaverStruct) FileClose(r repo.Request) error {
	f := sv.F[r.Name()]
	delete(sv.F, r.Name())
	if len(sv.F) == 0 {
		sv.F = make(map[string]*os.File)
	}
	err := f.Close()
	if err != nil {
		return err
	}
	return nil
}

// TableSave saves table map on disk as .json file.
// Tested in saver_test.go
func (sv *SaverStruct) TableSave(m map[string]repo.NameNumber, ts string) error {

	//logger.L.Infof("saver.TableSave was invoked with m = %v, ts: %q\n", m, ts)

	mSimplified := simplify(m)

	return sv.tableSave(mSimplified, ts)
}

// simplify returns m with all Number fields removed.
// Tested in saver_test.go
func simplify(m map[string]repo.NameNumber) map[string]string {
	result := make(map[string]string)

	for i, v := range m {
		result[i] = v.Name
	}
	return result
}

// tableSave performs saving
func (sv *SaverStruct) tableSave(m map[string]string, ts string) error {
	folderPath := sv.Path + "/" + ts
	_, err := os.Stat(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(folderPath, 0777)
			if err != nil {
				return fmt.Errorf("in saver.saveTable unable to create folder %q: %v", folderPath, err)
			}
		} else {
			return fmt.Errorf("in saver.saveTable error while finding folder %q: %v", folderPath, err)
		}

	}
	filePath := folderPath + "/" + ts + ".json"
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("in saver.saveTable unable to create file %q: %v", filePath, err)
	}
	defer f.Close()

	JSONed, err := json.MarshalIndent(m, "", "   ")
	if err != nil {
		return fmt.Errorf("in saver.saveTable unable to marshal %v: %v", m, err)
	}
	//logger.L.Infof("in saver.tableSave saving %q into %s\n", JSONed, filePath)
	_, err = f.Write(JSONed)
	if err != nil {
		return fmt.Errorf("in saver.saveTable unable write %q to file %q: %v", JSONed, filePath, err)
	}
	return nil
}
