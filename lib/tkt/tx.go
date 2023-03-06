package tkt

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
)

type Future func()

type TxCtx struct {
	databaseConfig DatabaseConfig
	tx             *sql.Tx
	db             *sql.DB
	stmtMap        map[string]*sql.Stmt
	insMap         map[string]*sql.Stmt
	autoIdMap      map[string]*sql.Stmt
	updMap         map[string]*sql.Stmt
	delMap         map[string]*sql.Stmt
	sequences      *Sequences
	future         []Future
}

func (o *TxCtx) Tx() *sql.Tx {
	return o.tx
}

func (o *TxCtx) Db() *sql.DB {
	return o.db
}

func (o *TxCtx) Seq() *Sequences {
	return o.sequences
}

func (o *TxCtx) NextId(name string) *int64 {
	id := o.sequences.Next(name)
	return &id
}

func (o *TxCtx) FindStruct(template interface{}, sql string, queryParams ...interface{}) interface{} {
	stmt := o.resolveStmt(sql)
	return FindStructStmt(stmt, template, queryParams...)
}

func (o *TxCtx) QueryStruct(template interface{}, sql string, queryParams ...interface{}) interface{} {
	stmt := o.resolveStmt(sql)
	return QueryStructStmt(stmt, template, queryParams...)
}

func (o *TxCtx) QueryStructStmt(stmt *sql.Stmt, template interface{}, queryParams ...interface{}) interface{} {
	return QueryStructStmt(stmt, template, queryParams...)
}

func (o *TxCtx) InsertEntity(schema string, data interface{}, autoId bool) int64 {
	offset := 0
	if autoId {
		offset = 1
	}
	name := reflect.TypeOf(data).Name()
	var insMap map[string]*sql.Stmt
	if autoId {
		insMap = o.autoIdMap
	} else {
		insMap = o.insMap
	}
	key := schema + "." + name
	stmt, ok := insMap[key]
	if !ok {
		sentence := "insert into " + schema + "." + name + ForInsert(data, offset)
		var err error
		stmt, err = o.tx.Prepare(sentence)
		CheckErr(err)
		insMap[key] = stmt
	}
	return o.ExecStructStmt(stmt, data, offset)
}

func (o *TxCtx) UpdateEntity(schema string, entity interface{}) int64 {
	objectType := reflect.TypeOf(entity)
	name := objectType.Name()
	key := schema + "." + name
	stmt, ok := o.updMap[key]
	if !ok {
		idName := o.resolveIdName(objectType)
		sentence := "update " + schema + "." + objectType.Name() + " set " + ForUpdate(entity, 1, 2) +
			" where " + idName + " = $1"
		var err error
		stmt, err = o.tx.Prepare(sentence)
		CheckErr(err)
		o.updMap[key] = stmt
	}
	return o.ExecStructStmt(stmt, entity, 0)
}

func (o *TxCtx) resolveIdName(objectType reflect.Type) string {
	idField, ok := objectType.FieldByName("Id")
	if !ok {
		panic(fmt.Sprintf("Id field not found for %s", objectType.Name()))
	}
	idName, ok := idField.Tag.Lookup("sql")
	if ok {
		return idName
	} else {
		return "Id"
	}
}

func (o *TxCtx) DeleteEntity(schema string, entity interface{}) {
	objectType := reflect.TypeOf(entity)
	name := objectType.Name()
	key := schema + "." + name
	stmt, ok := o.delMap[key]
	if !ok {
		idName := o.resolveIdName(objectType)
		sentence := "delete from " + schema + "." + name + " where " + idName + " = $1"
		var err error
		stmt, err = o.tx.Prepare(sentence)
		CheckErr(err)
		o.delMap[key] = stmt
	}
	o.ExecStmt(stmt, reflect.ValueOf(entity).FieldByName("Id").Interface())
}

func (o *TxCtx) ExecStruct(sql string, data interface{}, offset ...int) int64 {
	stmt := o.resolveStmt(sql)
	return o.ExecStructStmt(stmt, data, offset...)
}

func (o *TxCtx) ExecStructStmt(stmt *sql.Stmt, data interface{}, varOffset ...int) int64 {
	if len(varOffset) > 0 {
		return ExecStructStmtOff(stmt, data, varOffset[0])
	} else {
		return ExecStructStmt(stmt, data)
	}
}

func (o *TxCtx) ExecSql(sql string, args ...interface{}) *sql.Result {
	stmt := o.resolveStmt(sql)
	return ExecStmt(stmt, args...)
}

func (o *TxCtx) ExecStmt(stmt *sql.Stmt, args ...interface{}) *sql.Result {
	return ExecStmt(stmt, args...)
}

func (o *TxCtx) QuerySql(sql string, args ...interface{}) *sql.Rows {
	stmt := o.resolveStmt(sql)
	return QueryStmt(stmt, args...)
}

func (o *TxCtx) QuerySingleton(sql string, fields []interface{}, args ...interface{}) bool {
	stmt := o.resolveStmt(sql)
	return QuerySingletonStmt(stmt, fields, args...)
}

func (o *TxCtx) resolveStmt(sql string) *sql.Stmt {
	stmt, ok := o.stmtMap[sql]
	if !ok {
		var err error
		stmt, err = o.tx.Prepare(sql)
		CheckErr(err)
		o.stmtMap[sql] = stmt
	}
	return stmt
}

func (o *TxCtx) AddFuture(f func()) {
	if o.future == nil {
		o.future = make([]Future, 1)
		o.future[0] = f
	} else {
		o.future = append(o.future, f)
	}
}

func NewTxCtx(databaseConfig DatabaseConfig, tx *sql.Tx, db *sql.DB) *TxCtx {
	sequences := NewSequences(databaseConfig, tx)
	txCtx := TxCtx{tx: tx, db: db, stmtMap: make(map[string]*sql.Stmt), insMap: make(map[string]*sql.Stmt),
		autoIdMap: make(map[string]*sql.Stmt), updMap: make(map[string]*sql.Stmt), delMap: make(map[string]*sql.Stmt),
		sequences: sequences, databaseConfig: databaseConfig}
	return &txCtx
}

func InterceptTransactional(databaseConfig *DatabaseConfig, delegate func(txCtx *TxCtx, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := OpenDB(*databaseConfig)
		defer CloseDB(db)
		tx, err := db.Begin()
		if err != nil {
			panic(err)
		}
		defer RollbackOnPanic(tx)
		txCtx := NewTxCtx(*databaseConfig, tx, db)
		ctx := context.WithValue(r.Context(), "txCtx", txCtx)
		delegate(txCtx, w, r.WithContext(ctx))
		CheckErr(tx.Commit())
		if txCtx.future != nil {
			for _, f := range txCtx.future {
				f()
			}
		}
	}
}

func InterceptReadOnlyTransactional(databaseConfig *DatabaseConfig, delegate func(txCtx *TxCtx, w http.ResponseWriter, r *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db := OpenDB(*databaseConfig)
		defer CloseDB(db)
		txOptions := sql.TxOptions{ReadOnly: true}
		tx, err := db.BeginTx(context.Background(), &txOptions)
		if err != nil {
			panic(err)
		}
		defer RollbackOnPanic(tx)
		txCtx := NewTxCtx(*databaseConfig, tx, db)
		ctx := context.WithValue(r.Context(), "txCtx", txCtx)
		delegate(txCtx, w, r.WithContext(ctx))
		CheckErr(tx.Commit())
		if txCtx.future != nil {
			for _, f := range txCtx.future {
				f()
			}
		}
	}
}

func ExecuteTransactional(config DatabaseConfig, callback func(txCtx *TxCtx, args ...interface{}) interface{}, args ...interface{}) interface{} {
	db := OpenDB(config)
	defer CloseDB(db)
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	defer RollbackOnPanic(tx)
	txCtx := NewTxCtx(config, tx, db)
	r := callback(txCtx, args...)
	CheckErr(tx.Commit())
	if txCtx.future != nil {
		for _, f := range txCtx.future {
			f()
		}
	}
	return r
}
