package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)
var (
	outFileName = "segment-data-"
	outFileSize int64 = 10000000
)

// db
type Db struct {
	blocks []*block

	dir           string
	segmentName   string
	segmentNumber int
	segmentSize   int64
}

func NewDb(dir string) (*Db, error) {
	db := &Db{
		dir:         dir,
		segmentName: outFileName,
		segmentSize: outFileSize,
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, os.ModePerm)
	}

	f, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	filesNames, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	if len(filesNames) != 0 {
		err := db.recover(filesNames)
		if err != nil {
			return nil, err
		}
	} else {
		err = db.addNewBlockToDB()
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func (db *Db) addNewBlockToDB() error {
	db.segmentNumber++
	b, err := newBlock(db.dir,
		db.segmentName+strconv.Itoa((db.segmentNumber)))
	if err != nil {
		return err
	}
	db.blocks = append(db.blocks, b)
	return nil
}

func (db *Db) recover(filesNames []string) error { // sort by growth
	sort.Strings(filesNames)
	// regexp for checking file names
	r, _ := regexp.Compile(db.segmentName + "[0-9]+")
	for _, fileName := range filesNames {
		match := r.MatchString(fileName)

		if match {
			b, err := newBlock(db.dir, fileName)
			if err != nil {
				return err
			}
			db.blocks = append(db.blocks, b)
			reg, _ := regexp.Compile("[0-9]+")
			db.segmentNumber, err = strconv.Atoi(reg.FindString(fileName))
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("wrongly named file in the working directory: %v. Current file neme pattern: %v + int number", fileName, db.segmentName)
		}
	}
	return nil
}

func (db *Db) Close() error {
	for _, block := range db.blocks {
		block.close()
	}
	return nil
}

func (db *Db) Put(key, value string) error {
	lastBlock := db.blocks[len(db.blocks)-1]
	curSize, err := lastBlock.size()
	if err != nil {
		return err
	}

	if curSize <= db.segmentSize {
		err := lastBlock.put(key, value)
		if err != nil {
			return err
		}
		return nil
	}

	err = db.addNewBlockToDB()
	if err != nil {
		return err
	}

	err = db.blocks[len(db.blocks)-1].put(key, value)
	if err != nil {
		return err
	}

	if len(db.blocks) > 2 {
		err = db.merge()
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *Db) Get(key string) (string, error) {
	for j := len(db.blocks) - 1; j >= 0; j-- {
		val, err := db.blocks[j].get(key)
		if err != nil && err != ErrNotFound {
			return "", err
		}
		if val != "" {
			return val, nil
		}
	}
	return "", ErrNotFound
}

func (db *Db) merge() error {
	tempBlock, err := mergeAll(db.blocks[:len(db.blocks)-1])
	if err != nil {
		return err
	}

	db.blocks = append(db.blocks[:1], db.blocks[:]...)
	db.blocks[0] = tempBlock

	for _, block := range db.blocks[1 : len(db.blocks)-1] {
		err := block.delete()
		if err != nil {
			return err
		}
	}

	db.blocks = append(db.blocks[:1], db.blocks[len(db.blocks)-1])
	err = os.Rename(tempBlock.segment.Name(), filepath.Join(db.dir, db.segmentName+"0"))
	tempBlock.outPath = tempBlock.segment.Name()
	if err != nil {
		return err
	}
	return nil
}