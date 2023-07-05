package util

import (
	"github.com/golang/glog"

	"encoding/json"
	"os"
)

type Persister interface {
	Write(state interface{})
	Read(state interface{})
}

type FilePersister struct {
	Fn string
}

type InMemoryPersister struct {
}

func NewPersister(fname string) Persister {
	if fname == "" {
		return &InMemoryPersister{}
	}
	return &FilePersister{fname}
}

func (i *InMemoryPersister) Read(state interface{}) {
	return
}

func (i *InMemoryPersister) Write(state interface{}) {
	return
}

func (fp *FilePersister) Write(state interface{}) {
	glog.Infof("Writing to %v", fp.Fn)
	f, err := os.Create(fp.Fn)
	if err != nil {
		glog.Fatalln("Error opening %v: %v", fp.Fn, err)
	}

	writer := json.NewEncoder(f)
	err = writer.Encode(state)

	if err != nil {
		glog.Fatalf("Error encoding %v", state)
	}
	return
}

func (fp *FilePersister) Read(state interface{}) {
	glog.Infof("Reading from %v", fp.Fn)
	f, err := os.Open(fp.Fn)
	if err != nil {
		if os.IsNotExist(err) {
			glog.Infof("Nothing to recover...")
			return
		} else {
			glog.Fatalln(err)
		}
	}

	reader := json.NewDecoder(f)
	err = reader.Decode(state)
	if err != nil {
		glog.Fatalln(err)
	}
	return
}
