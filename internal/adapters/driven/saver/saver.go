// Saver adapter.
// Saves data to disk
package saver

import (
	"fmt"
	"os"
	"sync"

	json "github.com/goccy/go-json"
	"github.com/vynovikov/highLoadSaver/internal/adapters/driver/rpc/pb"
	"github.com/vynovikov/highLoadSaver/internal/repo"
)

type Saver interface {
	Save(*pb.Message) error
}

type SaverStruct struct {
	Path string
	F    map[string]*repo.FileInfo
	T    map[string]string
	l    sync.Mutex
}

func NewSaver(path string) (*SaverStruct, error) {
	f := make(map[string]*repo.FileInfo)
	t := make(map[string]string)
	_, err := os.Stat(path)

	if err != nil {

		if os.IsNotExist(err) {

			os.Mkdir(path, 0777)

			return &SaverStruct{Path: path, F: f, T: t}, nil
		}
		return &SaverStruct{}, err
	}
	return &SaverStruct{Path: path, F: f, T: t}, nil
}

func (s *SaverStruct) Save(m *pb.Message) error {
	if len(s.T) == 0 {
		s.createFolder(m.Ts)
	}
	//logger.L.Infof("in saver.Save receiving m %v, s.T became %v\n", m, s.T)
	if len(m.FileName) > 0 {
		filePath, err := s.saveToFile(m)
		if err != nil {
			return err
		}
		if _, ok := s.T[m.FormName]; !ok {
			s.T[m.FormName] = filePath
		}
	} else { // fix empty formName issue
		s.T[m.FormName] = string(m.FieldValue)
	}

	//logger.L.Infof("in saver.Save after receiving m %v, s.T became %v\n", m, s.T)

	if m.Last {
		err := s.saveToTable(m)
		if err != nil {
			return err
		}
		s.closeFiles()
		s.reset()
	}
	return nil
}
func (s *SaverStruct) createFolder(ts string) error {
	folderName := "results" + "/" + ts
	err := os.Mkdir(folderName, 0777)
	if err != nil {
		return fmt.Errorf("in saver.createFolder unable to create folder %q: %v", folderName, err)
	}
	return nil
}

func (s *SaverStruct) getFileForMessageSaving(m *pb.Message) (*repo.FileInfo, error) {
	var (
		f        *os.File
		err      error
		fileName string
	)
	folderName := "results" + "/" + m.Ts
	fileName = folderName + "/" + m.FileName

	if FI, ok := s.F[m.FormName]; ok {
		return FI, nil
	}
	f, err = os.Create(fileName)
	if err != nil {
		return &repo.FileInfo{}, fmt.Errorf("in saver.ToTable unable to create file %q: %v", fileName, err)
	}

	return repo.NewFileInfo(f, 0), nil
}

func (s *SaverStruct) getFileForTableSaving(ts string) (*repo.FileInfo, error) {
	var (
		f        *os.File
		err      error
		fileName string
	)
	folderName := "results" + "/" + ts
	fileName = folderName + "/" + ts + ".json"

	f, err = os.Create(fileName)
	if err != nil {
		return &repo.FileInfo{}, fmt.Errorf("in saver.ToTable unable to create file %q: %v", fileName, err)
	}

	return repo.NewFileInfo(f, 0), nil
}

func (s *SaverStruct) saveToFile(m *pb.Message) (string, error) {
	FI, err := s.getFileForMessageSaving(m)
	if err != nil {
		return "", err
	}
	//logger.L.Infof("in saver.SaveToFile m.FieldValue = %q, FI = %v", m.FieldValue, FI)
	n, err := FI.F.WriteAt(m.FieldValue, FI.O)
	if err != nil {
		return "", err
	}
	k := int64(n)
	FI.AddOffset(k)

	if _, ok := s.F[m.FormName]; !ok {
		s.F[m.FormName] = FI
	}
	return m.Ts + "/" + m.FileName, nil
}

func (s *SaverStruct) saveToTable(m *pb.Message) error {
	FI, err := s.getFileForTableSaving(m.Ts)
	if err != nil {
		return err
	}
	fileName := m.Ts + "/" + m.Ts + ".json"

	JSONed, err := json.MarshalIndent(s.T, "", "  ")
	if err != nil {
		return fmt.Errorf("in saver.ToTable unable unmarshal map %v: %v", s.T, err)
	}
	_, err = FI.F.Write(JSONed)
	if err != nil {
		return fmt.Errorf("in saver.ToTable unable to write to file %q: %v", fileName, err)
	}
	return nil
}

func (s *SaverStruct) closeFiles() []error {
	errs := make([]error, 0, 15)
	for _, v := range s.F {
		//logger.L.Infof("in saver.closeFiles closing file corresponding to %s\n", i)
		err := v.F.Close()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (s *SaverStruct) reset() {
	s.T = make(map[string]string)
	s.F = make(map[string]*repo.FileInfo)
}
