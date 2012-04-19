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
	Persist(bm *BomMeta, b *Bom, version ShortName) error
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
	bm := BomMeta{}
	if err := readJsonBomMeta(fpath, &bm); err != nil {
		return nil, err
	}
	return &bm, nil
}

func (jfbs *JSONFileBomStore) GetHead(user, name ShortName) (*BomMeta, *Bom, error) {
	bm, err := jfbs.GetBomMeta(user, name)
	if err != nil {
		return nil, nil, err
	}
	version := bm.HeadVersion
	if version == "" {
		log.Fatal("Tried to read undefined HEAD for " + string(user) + "/" + string(name))
	}
	b, err := jfbs.GetBom(user, name, ShortName(version))
	return bm, b, err
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
	bmList := []BomMeta{}
	dirInfo, err := rootDir.Readdir(0)
	for _, node := range dirInfo {
		if !node.IsDir() || !isShortName(node.Name()) {
			continue
		}
		uList, err := jfbs.listBomsForUser(ShortName(node.Name()))
		if err != nil {
			log.Fatal(err)
		}
		bmList = append(bmList, uList...)
	}
	return bmList, nil
}

func (jfbs *JSONFileBomStore) listBomsForUser(user ShortName) ([]BomMeta, error) {
	bmList := []BomMeta{}
	uDirPath := jfbs.Rootfpath + "/" + string(user)
	uDir, err := os.Open(uDirPath)
	if err != nil {
		if e, ok := err.(*os.PathError); ok && e.Err.Error() == "no such file or directory" {
			// XXX: should probably check for a specific syscall error? same below
			return bmList, nil
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
		bm := BomMeta{}
		if err := readJsonBomMeta(fpath, &bm); err != nil {
			if e, ok := err.(*os.PathError); ok && e.Err.Error() == "no such file or directory" {
				// no _meta.json in there
				continue
			}
			return nil, err
		}
		bmList = append(bmList, bm)
	}
	return bmList, nil
}

func (jfbs *JSONFileBomStore) Persist(bm *BomMeta, b *Bom, version ShortName) error {

	if err := bm.Validate(); err != nil {
		return err
	}
	if err := b.Validate(); err != nil {
		return err
	}

	b_fpath := jfbs.Rootfpath + "/" + string(bm.Owner) + "/" + string(bm.Name) + "/" + string(version) + ".json"
	bm_fpath := jfbs.Rootfpath + "/" + string(bm.Owner) + "/" + string(bm.Name) + "/_meta.json"

	if f, err := os.Open(b_fpath); err == nil {
		f.Close()
		return Error("bom with same owner, name, and version already exists")
	}
	if err := writeJsonBomMeta(bm_fpath, bm); err != nil {
		log.Fatal(err)
	}
	if err := writeJsonBom(b_fpath, b); err != nil {
		log.Fatal(err)
	}
	return nil
}

func readJsonBomMeta(fpath string, bm *BomMeta) error {
	f, err := os.Open(path.Clean(fpath))
	if err != nil {
		return err
	}
	defer f.Close()
	dec := json.NewDecoder(f)
	if err = dec.Decode(&bm); err != nil {
		return err
	}
	return nil
}

func writeJsonBomMeta(fpath string, bm *BomMeta) error {
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
	if err = enc.Encode(&bm); err != nil {
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
