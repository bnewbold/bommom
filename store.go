package main

import (
	"encoding/json"
	"log"
	"os"
)

var bomstore BomStore

// TODO: who owns returned BOMs? Caller? need "free" methods?
type BomStore interface {
	GetStub(user, name ShortName) (*BomStub, error)
	GetHead(user, name ShortName) (*Bom, error)
	GetBom(user, name, version ShortName) (*Bom, error)
	Persist(bs *BomStub, b *Bom, version ShortName) error
}

// Basic BomStore backend using a directory structure of JSON files saved to
// disk.
type JSONFileBomStore struct {
	RootPath string
}

func NewJSONFileBomStore(path string) *JSONFileBomStore {
	err := os.MkdirAll(path, os.ModePerm|os.ModeDir)
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}
	return &JSONFileBomStore{RootPath: path}
}

func (jfbs *JSONFileBomStore) GetStub(user, name ShortName) (*BomStub, error) {
	path := jfbs.RootPath + "/" + string(user) + "/" + string(name) + "/meta.json"
	bs := BomStub{}
	if err := readJsonBomStub(path, &bs); err != nil {
		return nil, err
	}
	return &bs, nil
}

func (jfbs *JSONFileBomStore) GetHead(user, name ShortName) (*Bom, error) {
	bs, err := jfbs.GetStub(user, name)
	if err != nil {
		return nil, err
	}
	version := bs.HeadVersion
	if version == "" {
		log.Fatal("Tried to read undefined HEAD for " + string(user) + "/" + string(name))
	}
	return jfbs.GetBom(user, name, ShortName(version))
}

func (jfbs *JSONFileBomStore) GetBom(user, name, version ShortName) (*Bom, error) {
	path := jfbs.RootPath + "/" + string(user) + "/" + string(name) + "/" + string(version) + ".json"
	b := Bom{}
	if err := readJsonBom(path, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (jfbs *JSONFileBomStore) Persist(bs *BomStub, b *Bom, version ShortName) error {
	return nil
}

func readJsonBomStub(path string, bs *BomStub) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err = dec.Decode(&bs); err != nil {
		return err
	}
	return nil
}

func writeJsonBomStub(path string, bs *BomStub) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(&bs); err != nil {
		return err
	}
	return nil
}

func readJsonBom(path string, b *Bom) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err = dec.Decode(&b); err != nil {
		return err
	}
	return nil
}

func writeJsonBom(path string, b *Bom) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	if err = enc.Encode(&b); err != nil {
		return err
	}
	return nil
}
