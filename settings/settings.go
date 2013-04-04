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
}

func (settings *Settings) Save() {
	data, err := xml.MarshalIndent(settings, "", "  ")
	if err != nil {
		log.Println(err)
	}
	err = ioutil.WriteFile(SettingsPath, data, os.ModePerm)
	if err != nil {
		log.Println(err)
	}
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

