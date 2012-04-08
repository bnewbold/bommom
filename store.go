package main

var bomstore BomStore

// TODO: who owns returned BOMs? Caller? need "free" methods?
type BomStore interface {
	GetStub(user, name ShortName) (*BomStub, error)
	GetHead(user, name ShortName) (*Bom, error)
	GetBom(user, name, version ShortName) (*Bom, error)
	Persist(bom *Bom) error
}

/*
// Dummy BomStore backed by hashtable in memory, for testing and demo purposes
type MemoryBomStore map[string] Bom
*/

// Basic BomStore backend using a directory structure of JSON files saved to
// disk.
type JSONFileBomStore struct {
	rootPath string
}
