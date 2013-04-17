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
	WriteLogToScreen bool
	LoopOnEmpty bool
	DoNotBootStrap bool
	CheckSumMaxSize int64
	SandwichDirName string
	LocalServerPort string
	DontOpenBrowserOnStart bool
}

func (settings *Settings) Clone() *Settings {
	retVal := new(Settings)
	retVal.PingUntilFoundOnStart = settings.PingUntilFoundOnStart
	retVal.WriteLogToScreen = settings.WriteLogToScreen
	retVal.LoopOnEmpty = settings.LoopOnEmpty
	retVal.DoNotBootStrap = settings.DoNotBootStrap
	retVal.CheckSumMaxSize = settings.CheckSumMaxSize
	retVal.SandwichDirName = settings.SandwichDirName
	retVal.LocalServerPort = settings.LocalServerPort
	retVal.DontOpenBrowserOnStart = settings.DontOpenBrowserOnStart
	return retVal
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

