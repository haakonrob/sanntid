package backup

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
)

func Restart() {
	backup := exec.Command("gnome-terminal", "-x", "sh", "-c", "go run coordinator.go ./backupdata")
	backup.Run()
	os.Exit(0)
}

func Write(state interface{}) {
	data, _ := json.Marshal(state)
	_ = ioutil.WriteFile("./backupdata", data, 0644)
}

func Load(filePath string) (interface{}, bool) {
	data, _ := ioutil.ReadFile(filePath)
	var temp interface{}
	err := json.Unmarshal(data, &temp)
	if err == nil {
		return temp, true
	}
	return temp, false
}
