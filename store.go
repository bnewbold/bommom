package main

import (
	"encoding/json"
	"log"
	"os"
	"path"
)

// TODO: who owns returned BOMs? Caller? need "free" methods?
type BomStore interface {
	GetBomMeta(user, name ShortName) (*BomMeta, error)
	GetHead(user, name ShortName) (*BomMeta, *Bom, error)
	GetBom(user, name, version ShortName) (*Bom, error)
	Persist(bs *BomMeta, b *Bom, version ShortName) error
	ListBoms(user ShortName) ([]BomMeta, error)
}

// Basic BomStore backend using a directory structure of JSON files saved to
// disk.
type JSONFileBomStore struct {
	Rootfpath string
}

func NewJSONFileBomStore(fpath string) error {
	err := os.MkdirAll(fpath, os.ModePerm|os.ModeDir)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func OpenJSONFileBomStore(fpath string) (*JSONFileBomStore, error) {
	_, err := os.Open(fpath)
	if err != nil && !os.IsExist(err) {
		return nil, err
	}
	return &JSONFileBomStore{Rootfpath: fpath}, nil
}

func (jfbs *JSONFileBomStore) GetBomMeta(user, name ShortName) (*BomMeta, error) {
	fpath := jfbs.Rootfpath + "/" + string(user) + "/" + string(name) + "/_meta.json"
	bs := BomMeta{}
	if err := readJsonBomMeta(fpath, &bs); err != nil {
		return nil, err
	}
	return &bs, nil
}

func (jfbs *JSONFileBomStore) GetHead(user, name ShortName) (*BomMeta, *Bom, error) {
	bs, err := jfbs.GetBomMeta(user, name)
	if err != nil {
		return nil, nil, err
	}
	version := bs.HeadVersion
	if version == "" {
		log.Fatal("Tried to read undefined HEAD for " + string(user) + "/" + string(name))
	}
	b, err := jfbs.GetBom(user, name, ShortName(version))
	return bs, b, err
}

func (jfbs *JSONFileBomStore) GetBom(user, name, version ShortName) (*Bom, error) {
	fpath := jfbs.Rootfpath + "/" + string(user) + "/" + string(name) + "/" + string(version) + ".json"
	b := Bom{}
	if err := readJsonBom(fpath, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (jfbs *JSONFileBomStore) ListBoms(user ShortName) ([]BomMeta, error) {
	if user != "" {
		return jfbs.listBomsForUser(user)
	}
	// else iterator over all users...
	rootDir, err := os.Open(jfbs.Rootfpath)
	if err != nil {
		log.Fatal(err)
	}
	defer rootDir.Close()
	bsList := []BomMeta{}
	dirInfo, err := rootDir.Readdir(0)
	for _, node := range dirInfo {
		if !node.IsDir() || !isShortName(node.Name()) {
			continue
		}
		uList, err := jfbs.listBomsForUser(ShortName(node.Name()))
		if err != nil {
			log.Fatal(err)
		}
		bsList = append(bsList, uList...)
	}
	return bsList, nil
}

func (jfbs *JSONFileBomStore) listBomsForUser(user ShortName) ([]BomMeta, error) {
	bsList := []BomMeta{}
	uDirPath := jfbs.Rootfpath + "/" + string(user)
	uDir, err := os.Open(uDirPath)
	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err.Error() == "no such file or directory" {
			// XXX: should probably check for a specific syscall error? same below
			return bsList, nil
		}
		return nil, err
	}
	defer uDir.Close()
	dirContents, err := uDir.Readdir(0)
	if err != nil {
		return nil, err
	}
	for _, node := range dirContents {
		if !node.IsDir() || !isShortName(node.Name()) {
			continue
		}
		fpath := jfbs.Rootfpath + "/" + string(user) + "/" + node.Name() + "/_meta.json"
		bs := BomMeta{}
		if err := readJsonBomMeta(fpath, &bs); err != nil {
			if e, ok := err.(*os.PathError); ok && e.Err.Error() == "no such file or directory" {
				// no _meta.json in there
				continue
			}
			return nil, err
		}
		bsList = append(bsList, bs)
	}
	return bsList, nil
}

func (jfbs *JSONFileBomStore) Persist(bs *BomMeta, b *Bom, version ShortName) error {

	if err := bs.Validate(); err != nil {
		return err
	}
	if err := b.Validate(); err != nil {
		return err
	}

	b_fpath := jfbs.Rootfpath + "/" + string(bs.Owner) + "/" + string(bs.Name) + "/" + string(version) + ".json"
	bs_fpath := jfbs.Rootfpath + "/" + string(bs.Owner) + "/" + string(bs.Name) + "/_meta.json"

	if f, err := os.Open(b_fpath); err == nil {
		f.Close()
		return Error("bom with same owner, name, and version already exists")
	}
	if err := writeJsonBomMeta(bs_fpath, bs); err != nil {
		log.Fatal(err)
	}
	if err := writeJsonBom(b_fpath, b); err != nil {
		log.Fatal(err)
	}
	return nil
}

func readJsonBomMeta(fpath string, bs *BomMeta) error {
	f, err := os.Open(path.Clean(fpath))
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

func writeJsonBomMeta(fpath string, bs *BomMeta) error {
	err := os.MkdirAll(path.Dir(fpath), os.ModePerm|os.ModeDir)
	if err != nil && !os.IsExist(err) {
		return err
	}
	f, err := os.Create(fpath)
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

func readJsonBom(fpath string, b *Bom) error {
	f, err := os.Open(path.Clean(fpath))
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

// Need to write the BomMeta before writing the Bom
func writeJsonBom(fpath string, b *Bom) error {
	f, err := os.Create(fpath)
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
