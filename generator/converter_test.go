package generator

import (
	"fmt"
	"os"
	"testing"
)

func TestDecode(t *testing.T) {
	mf, err := os.Open("../models/entity/blog.yml")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		_ = mf.Close()
	}()

	scheme, err := Decode(mf)
	if err != nil {
		t.Error(err)
		return
	}

	modelBytes, err := scheme.GenModel()
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Printf("%s", modelBytes)
}

func TestScheme_GenTable(t *testing.T) {
	mf, err := os.Open("../models/entity/user.yml")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		_ = mf.Close()
	}()

	scheme, err := Decode(mf)
	if err != nil {
		t.Error(err)
		return
	}

	tableStr, err := scheme.GenTable()
	if err != nil {
		t.Error(err)
	}

	fmt.Println(tableStr)
}

func TestScheme_GenApi(t *testing.T) {
	mf, err := os.Open("../models/entity/user.yml")
	if err != nil {
		t.Error(err)
		return
	}

	defer func() {
		_ = mf.Close()
	}()

	scheme, err := Decode(mf)
	if err != nil {
		t.Error(err)
		return
	}

	str := scheme.GenApi()
	fmt.Println(str)
}
