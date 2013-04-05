package settings

import(
	"encoding/xml"
	"os"
	"io/ioutil"
	"log"
)

var SettingsPath string

type Settings struct {
	PingUntilFoundOnStart bool
	ListenLocal bool
	DisableInterface bool
	WriteLogToScreen bool
	LoopOnEmpty bool
	DoNotBootStrap bool
	CheckSumMaxSize int64
	SandwichDirName string
}

func (settings *Settings) Save() error {
	data, err := xml.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Println(err)
		return err
	}
	err = ioutil.WriteFile(SettingsPath, data, os.ModePerm)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func Load() (*Settings, error) {
	data, err := ioutil.ReadFile(SettingsPath)
	if err != nil {
		return nil, err
	}
	settings := &Settings{}
	err = xml.Unmarshal(data, settings)
	return settings, err
}

