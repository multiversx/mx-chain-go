package config

import (
	"os"
	"testing"
)

func TestOpenReadConfigFile(t *testing.T) {
	filename := "configTest.json"

	_, err := os.Stat(filename)

	if !os.IsNotExist(err) {
		//file exists, try to delete
		var err = os.Remove(filename)
		if err != nil {
			t.Fatalf("Filename %s can not be removed! [%s]", filename, err.Error())
		}
	}

	file, err := os.Create("configTest.json")

	if err != nil {
		t.Fatal("Unexpected file error [writing]: " + err.Error())
	}

	file.WriteString(`
{
  "Port": 8080,
  "Peers": ["a", "b"]
}
`)

	err = file.Close()

	if err != nil {
		t.Fatal("Unexpected file error [closing]: " + err.Error())
	}

	//json test file is generated, try to read and interpret data...

	var config JsonConfig

	err = config.ReadFromFile(filename)

	_, err = os.Stat(filename)

	if !os.IsNotExist(err) {
		//file exists, try to delete
		var err = os.Remove(filename)
		if err != nil {
			t.Fatalf("Filename %s can not be removed! [%s]", filename, err.Error())
		}
	}

	if err != nil {
		t.Fatal("Unexpected file error [reading]: " + err.Error())
	}

	if config.Port != 8080 {
		t.Fatal("Invalid port read!")
	}

	if len(config.Peers) != 2 && config.Peers[0] != "a" && config.Peers[1] != "b" {
		t.Fatal("Invalid peers!")
	}
}
