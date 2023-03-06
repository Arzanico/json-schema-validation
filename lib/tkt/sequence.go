package tkt

import (
	"database/sql"
	_ "github.com/lib/pq"
	"sync"
)

type Sequence struct {
	Name   string
	LastId int64
}

var inMemoryManagerMap = make(map[string]*InMemorySequenceManager)
var mux = sync.Mutex{}

type SequenceManager interface {
	next(string) int64
}

type InMemorySequenceManager struct {
	SequenceManager
	sequenceMap    map[string]*Sequence
	databaseConfig DatabaseConfig
	mux            sync.Mutex
}

func (o *InMemorySequenceManager) next(name string) int64 {
	o.mux.Lock()
	defer o.mux.Unlock()
	seq, ok := o.sequenceMap[name]
	if !ok {
		seq = &Sequence{name, 0}
		o.initSequence(seq)
		o.sequenceMap[name] = seq
	}
	seq.LastId = seq.LastId + 1
	return seq.LastId
}

func (o *InMemorySequenceManager) initSequence(sequence *Sequence) {
	db := OpenDB(o.databaseConfig)
	defer db.Close()
	tx, err := db.Begin()
	CheckErr(err)
	defer tx.Commit()
	result, err := tx.Query("select max(id) from " + sequence.Name)
	CheckErr(err)
	defer result.Close()
	result.Next()
	var r sql.NullInt64
	result.Scan(&r)
	var id = int64(0)
	if r.Valid {
		id = r.Int64
	}
	sequence.LastId = id
}

func NewInMemorySequenceManager(config DatabaseConfig) *InMemorySequenceManager {
	mux.Lock()
	defer mux.Unlock()
	instance, ok := inMemoryManagerMap[*config.DatasourceName]
	if !ok {
		instance = &InMemorySequenceManager{sequenceMap: make(map[string]*Sequence, 0), databaseConfig: config, mux: sync.Mutex{}}
		inMemoryManagerMap[*config.DatasourceName] = instance
	}
	return instance
}

type PgSequenceManager struct {
	SequenceManager
	tx      *sql.Tx
	stmtMap map[string]*sql.Stmt
}

func (o *PgSequenceManager) next(name string) int64 {
	stmt := o.stmtMap[name]
	if stmt == nil {
		var err error
		stmt, err = o.tx.Prepare("select nextval('" + name + "seq')")
		CheckErr(err)
		o.stmtMap[name] = stmt
	}
	r := QueryStmt(stmt)
	defer r.Close()
	var id int64
	r.Next()
	CheckErr(r.Scan(&id))
	return id
}

func NewPgSequenceManager(tx *sql.Tx) *PgSequenceManager {
	seqs := PgSequenceManager{tx: tx, stmtMap: make(map[string]*sql.Stmt)}
	return &seqs
}

type Sequences struct {
	databaseConfig DatabaseConfig
	manager        SequenceManager
}

func (o *Sequences) Next(name string) int64 {
	return o.manager.next(name)
}

func NewSequences(config DatabaseConfig, tx *sql.Tx) *Sequences {
	manager := resolveSequenceManager(config, tx)
	return &Sequences{databaseConfig: config, manager: manager}
}

func resolveSequenceManager(config DatabaseConfig, tx *sql.Tx) SequenceManager {
	if config.SequenceManager == nil || *config.SequenceManager == "pgSequence" {
		return NewPgSequenceManager(tx)
	} else if *config.SequenceManager == "inMemory" {
		return NewInMemorySequenceManager(config)
	} else {
		panic("Unknown sequence manager: " + *config.SequenceManager)
	}
}
