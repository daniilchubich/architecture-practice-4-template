package datastore

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDb_Put(t *testing.T) {
	dir, err := ioutil.TempDir("", "test-db")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	db, err := NewDb(dir)
	if err != nil {
		t.Fatal(err)
	}

	pairs := [][]string{
		{"key1", "value1"},
		{"key2", "value2"},
		{"key3", "value3"},
	}

	outFile, err := os.Open(filepath.Join(dir, db.segmentName+strconv.Itoa((db.segmentNumber))))
	if err != nil {
		t.Fatal(err)
	}

	t.Run("PUT/GET", func(t *testing.T) {
		for _, pair := range pairs {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("ERROR! Can't put %s: %s", pair[0], err)
			}
			value, err := db.Get(pair[0])
			if err != nil {
				t.Errorf("ERROR! Can't get %s: %s", pair[0], err)
			}
			if value != pair[1] {
				t.Errorf("ERROR! Bad value returned expected %s, got %s", pair[1], value)
			}
		}
	})

	outInfo, err := outFile.Stat()
	if err != nil {
		t.Fatal(err)
	}
	size1 := outInfo.Size()

	t.Run("file growth", func(t *testing.T) {
		for _, pair := range pairs {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("ERROR! Can't put %s: %s", pair[0], err)
			}
		}
		outInfo, err := outFile.Stat()
		if err != nil {
			t.Fatal(err)
		}
		if size1*2 != outInfo.Size() {
			t.Errorf("ERROR! Unexpected size (%d vs %d)", size1, outInfo.Size())
		}
	})

	t.Run("new DB process", func(t *testing.T) {
		db, err = NewDb(dir)
		if err != nil {
			t.Fatal(err)
		}

		for _, pair := range pairs {
			value, err := db.Get(pair[0])
			if err != nil {
				t.Errorf("ERROR! Can't put %s: %s", pairs[0], err)
			}
			if value != pair[1] {
				t.Errorf("ERROR!\nExpected: %s;\nGot: %s", pair[1], value)
			}
		}
	})

	pairs2 := [][]string{
		{"keyA", "val1"},
		{"keyB", "val2"},
		{"keyC", "val3"},
		{"keyD", "val4"},
		{"keyA", "new"},
		{"keyB", "alsoNew"},
	}

	t.Run("create new out file", func(t *testing.T) {
		db.segmentSize = 200
		for _, pair := range pairs2 {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("ERROR! Can't put %s: %s", pair[0], err)
			}
		}

		files, err := os.Open(dir)
		if err != nil {
			t.Fatalf("ERROR! Unexpected error: %v", err)
		}
		defer files.Close()
		filesNames, err := files.Readdirnames(0)
		if err != nil {
			t.Fatalf("ERROR! Unexpected error: %v", err)
		}
		n := len(filesNames)
		if n != 2 {
			t.Errorf("ERROR!\nExpected: 2;\nGot: %v", n)
		}
	})

	t.Run("get with more than one file, ", func(t *testing.T) {
		value, err := db.Get(pairs2[5][0])
		if err != nil {
			t.Errorf("ERROR! Can't get %s: %s", pairs2[5], err)
		}
		if value != pairs2[5][1] {
			t.Errorf("ERROR!\nExpected: %s;\nGot: %s", pairs2[5], value)
		}

		value, err = db.Get(pairs[0][0])
		if err != nil {
			t.Errorf("ERROR! Can't get %s: %s", pairs2[5], err)
		}
		if value != pairs[0][1] {
			t.Errorf("ERROR!\nExpected: %s;\nGot: %s", pairs[0], value)
		}
	})

	t.Run("merge", func(t *testing.T) {
		for _, pair := range pairs2 {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("ERROR! Can't put %s: %s", pair[0], err)
			}
		}
		for _, pair := range pairs2 {
			err := db.Put(pair[0], pair[1])
			if err != nil {
				t.Errorf("ERROR! Can't put %s: %s", pair[0], err)
			}
		}

		files, err := os.Open(dir)
		if err != nil {
			t.Fatalf("ERROR! Unexpected error: %v", err)
		}
		defer files.Close()
		filesNames, err := files.Readdirnames(0)
		if err != nil {
			t.Fatalf("ERROR! Unexpected error: %v", err)
		}
		n := len(filesNames)
		if n != 2 {
			t.Errorf("ERROR!\nExpected: 2;\nGot: %v", n)
		}
	})
}