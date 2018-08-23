package files

import (
	"testing"
	"github.com/iwind/TeaGo/Tea"
)

func TestReader_ReadByte(t *testing.T) {
	reader, err := NewReader(Tea.TmpFile("test.txt"))
	if err != nil {
		t.Fatal(err)
	}

	//reader.Seek(1)

	for {
		data := reader.ReadByte()
		if len(data) > 0 {
			t.Log(string(data))
		} else {
			t.Log("EOF")
			break
		}
	}

	reader.Close()
}

func TestReader_Read(t *testing.T) {
	file := NewFile(Tea.TmpFile("test.txt"))
	reader, err := file.Reader()
	if err != nil {
		t.Fatal(err)
	}

	//reader.Seek(1)

	for {
		data := reader.Read(10)
		if len(data) > 0 {
			t.Log(string(data))
		} else {
			t.Log("EOF")
			break
		}
	}

	reader.Close()
}

func TestReader_ReadLine(t *testing.T) {
	file := NewFile(Tea.TmpFile("test.txt"))
	reader, err := file.Reader()
	if err != nil {
		t.Fatal(err)
	}

	//reader.Seek(1)

	for {
		data := reader.ReadLine()
		if len(data) > 0 {
			t.Log(string(data))
		} else {
			t.Log("EOF")
			break
		}
	}

	reader.Close()
}

func TestReader_ReadAll(t *testing.T) {
	file := NewFile(Tea.TmpFile("test.txt"))
	reader, err := file.Reader()
	if err != nil {
		t.Fatal(err)
	}

	//reader.Seek(1)

	data := reader.ReadAll()
	t.Log(string(data))

	reader.Close()
}

func TestReader_ReadJSON(t *testing.T) {
	file := NewFile(Tea.TmpFile("test.json"))
	reader, err := file.Reader()
	if err != nil {
		t.Fatal(err)
	}

	defer reader.Close()

	dataMap := map[string]interface{}{}
	err = reader.ReadJSON(&dataMap)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(dataMap)
}

func TestReader_ReadYAML(t *testing.T) {
	reader, err := NewReader(Tea.ConfigFile("server.conf"))
	if err != nil {
		t.Fatal(err)
	}

	defer reader.Close()

	dataMap := map[string]interface{}{}
	err = reader.ReadYAML(&dataMap)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(dataMap)
}
