package tkt

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"time"
)

var timeType = reflect.TypeOf(time.Time{})
var byteArrayType = reflect.TypeOf([]byte{})

type DatabaseConfig struct {
	DatabaseDriver  *string `json:"databaseDriver"`
	DatasourceName  *string `json:"datasourceName"`
	MaxIdleConns    *int    `json:"maxIdleConns"`
	MaxOpenConns    *int    `json:"maxOpenConns"`
	MaxConnLifetime *int    `json:"maxConnLifetime"`
	SequenceManager *string `json:"sequenceManager"`
}

func (o *DatabaseConfig) Validate() {
	if o.DatabaseDriver == nil {
		panic("Invalid databaseDriver")
	}
	if o.DatasourceName == nil {
		panic("Invalid datasourceName")
	}
}

func CloseDB(db *sql.DB) {
	db.Close()
}

func RollbackOnPanic(tx *sql.Tx) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}

func OpenDB(config DatabaseConfig) *sql.DB {
	db, err := sql.Open(*config.DatabaseDriver, *config.DatasourceName)
	if err != nil {
		panic(err)
	}
	if config.MaxIdleConns != nil {
		db.SetMaxIdleConns(*config.MaxIdleConns)
	}
	if config.MaxOpenConns != nil {
		db.SetMaxOpenConns(*config.MaxOpenConns)
	}
	if config.MaxConnLifetime != nil {
		db.SetConnMaxLifetime(time.Duration(*config.MaxConnLifetime) * time.Second)
	}
	return db
}

func PrepareStmt(tx *sql.Tx, sql string) *sql.Stmt {
	stmt, err := tx.Prepare(sql)
	CheckErr(err)
	return stmt
}

func QuerySingleton(tx *sql.Tx, fields []interface{}, sql string, args ...interface{}) bool {
	stmt, err := tx.Prepare(sql)
	CheckErr(err)
	return QuerySingletonStmt(stmt, fields, args...)
}

func QuerySingletonStmt(stmt *sql.Stmt, fields []interface{}, args ...interface{}) bool {
	r, err := stmt.Query(args...)
	CheckErr(err)
	defer r.Close()
	if r.Next() {
		CheckErr(r.Scan(fields...))
		return true
	} else {
		return false
	}
}

func ExecSql(tx *sql.Tx, sql string, args ...interface{}) *sql.Result {
	stmt, err := tx.Prepare(sql)
	CheckErr(err)
	return ExecStmt(stmt, args...)
}

func ExecStmt(stmt *sql.Stmt, args ...interface{}) *sql.Result {
	r, err := stmt.Exec(args...)
	CheckErr(err)
	return &r
}

func QuerySql(tx *sql.Tx, sql string, args ...interface{}) *sql.Rows {
	stmt, err := tx.Prepare(sql)
	CheckErr(err)
	return QueryStmt(stmt, args...)
}

func QueryStmt(stmt *sql.Stmt, args ...interface{}) *sql.Rows {
	r, err := stmt.Query(args...)
	CheckErr(err)
	return r
}

func Scan(r *sql.Rows, vars ...interface{}) {
	CheckErr(r.Scan(vars...))
}

func FindStruct(tx *sql.Tx, template interface{}, sql string, queryParams ...interface{}) interface{} {
	stmt, err := tx.Prepare(sql)
	CheckErr(err)
	defer stmt.Close()
	return FindStructStmt(stmt, template, queryParams...)
}

func FindStructStmt(stmt *sql.Stmt, template interface{}, queryParams ...interface{}) interface{} {
	result := QueryStructStmt(stmt, template, queryParams...)
	value := reflect.ValueOf(result)
	if value.Len() == 0 {
		objectType := reflect.TypeOf(template)
		return reflect.New(reflect.PtrTo(objectType)).Elem().Interface()
	} else {
		o := value.Index(0)
		return o.Addr().Interface()
	}
}

func QueryStruct(tx *sql.Tx, template interface{}, sql string, queryParams ...interface{}) interface{} {
	stmt, err := tx.Prepare(sql)
	CheckErr(err)
	defer stmt.Close()
	return QueryStructStmt(stmt, template, queryParams...)
}

func QueryStructStmt(stmt *sql.Stmt, template interface{}, queryParams ...interface{}) interface{} {
	objectType := reflect.TypeOf(template)
	fields := listStructFields(objectType, 0)
	r, e := stmt.Query(queryParams...)
	CheckErr(e)
	count := len(fields)
	cols, e := r.Columns()
	CheckErr(e)
	if len(cols) > count {
		panic("Result set column count greater than struct field count")
	}
	buffer := make([]interface{}, len(fields))
	for i := range fields {
		buffer[i] = reflect.New(fields[i].Type).Interface()
	}
	arr := reflect.MakeSlice(reflect.SliceOf(objectType), 0, 0)
	for r.Next() {
		CheckErr(r.Scan(buffer...))
		object := reflect.New(objectType).Elem()
		bufferToFields(object, buffer, 0)
		arr = reflect.Append(arr, object)
	}
	r.Close()
	return arr.Interface()
}

func bufferToFields(object reflect.Value, buffer []interface{}, offset int) int {
	instanceType := object.Type()
	instance := object
	isPtrStruct := object.Kind() == reflect.Ptr
	if isPtrStruct {
		instanceType = object.Type().Elem()
		instance = reflect.Indirect(reflect.New(instanceType))
	}
	n := 0
	created := !isPtrStruct
	for i := 0; i < instanceType.NumField(); i++ {
		of := instance.Field(i)
		if isArrayField(of.Type()) {
			continue
		}
		v := buffer[n+offset]
		indirect := reflect.ValueOf(v).Elem()
		if indirect.Kind() == reflect.Ptr && !indirect.IsNil() && !created {
			object.Set(instance.Addr())
			created = true
		}
		if isStructPtrField(of) {
			n += bufferToFields(of, buffer, n+offset)
		} else if isStructField(of) {
			n += bufferToFields(of, buffer, n+offset)
		} else {
			of.Set(indirect)
			n++
		}
	}
	return n
}

func isArrayField(t reflect.Type) bool {
	kind := t.Kind()
	return (kind == reflect.Slice || kind == reflect.Array) && t != byteArrayType
}

func isStructField(value reflect.Value) bool {
	return value.Kind() == reflect.Struct && value.Type() != timeType
}

func isStructPtrField(value reflect.Value) bool {
	if value.Kind() == reflect.Ptr {
		instanceType := value.Type().Elem()
		return instanceType.Kind() == reflect.Struct && instanceType != timeType
	} else {
		return false
	}

}

func listStructFields(structType reflect.Type, offset int) []reflect.StructField {
	fields := make([]reflect.StructField, 0)
	return addStructFields(structType, fields, offset)
}

func addStructFields(structType reflect.Type, fields []reflect.StructField, offset int) []reflect.StructField {
	if structType.Kind() == reflect.Ptr {
		structType = structType.Elem()
	}
	for i := offset; i < structType.NumField(); i++ {
		f := structType.Field(i)
		structPtr := false
		kind := f.Type.Kind()
		if kind == reflect.Ptr {
			instanceType := f.Type.Elem()
			if instanceType.Kind() == reflect.Struct && instanceType != timeType {
				structPtr = true
				fields = addStructFields(instanceType, fields, 0)
			}
		}
		if !structPtr {
			if kind == reflect.Struct && f.Type != timeType {
				fields = addStructFields(f.Type, fields, 0)
			} else if !isArrayField(f.Type) {
				fields = append(fields, f)
			}
		}
	}
	return fields
}

type FieldInfo struct {
	HolderType  *reflect.Type
	StructField *reflect.StructField
}

func ExecStruct(tx *sql.Tx, sql string, data interface{}) int64 {
	stmt, e := tx.Prepare(sql)
	CheckErr(e)
	defer stmt.Close()
	return ExecStructStmt(stmt, data)
}

func ExecStructStmt(stmt *sql.Stmt, data interface{}) int64 {
	return ExecStructStmtOff(stmt, data, 0)
}

func ExecStructStmtOff(stmt *sql.Stmt, data interface{}, offset int) int64 {
	objectType := reflect.TypeOf(data)
	fields := listStructFields(objectType, offset)
	buffer := make([]interface{}, len(fields))
	value := reflect.ValueOf(data)
	fieldsToBuffer(value, buffer, offset)
	r, err := stmt.Exec(buffer...)
	CheckErr(err)
	lastId, _ := r.LastInsertId()
	return lastId
}

func fieldsToBuffer(value reflect.Value, buffer []interface{}, offset int) {
	fields := buildObjectFields(value)
	for i := 0; i < len(buffer); i++ {
		f := fields[i+offset]
		buffer[i] = f.Interface()
	}
}

func buildObjectFields(value reflect.Value) []reflect.Value {
	fields := make([]reflect.Value, 0)
	for i := 0; i < value.NumField(); i++ {
		f := value.Field(i)
		if f.Kind() == reflect.Struct {
			fields = append(fields, buildObjectFields(f)...)
		} else {
			fields = append(fields, f)
		}
	}
	return fields
}

func ForInsert(template interface{}, offset int) string {
	objectType := reflect.TypeOf(template)
	buffer := bytes.NewBufferString("(")
	buffer.WriteString(forSelect(objectType, nil, offset))
	buffer.WriteString(") values(")
	n := 0
	for i := offset; i < objectType.NumField(); i++ {
		if n > 0 {
			buffer.WriteString(", ")
		}
		n++
		buffer.WriteString(fmt.Sprintf("$%d", n))
	}
	buffer.WriteString(")")
	return buffer.String()
}

func ForUpdate(template interface{}, offset int, firstNum int) string {
	objectType := reflect.TypeOf(template)
	buffer := bytes.NewBufferString("")
	for i := 0; i < objectType.NumField()-offset; i++ {
		if i > 0 {
			buffer.WriteString(", ")
		}
		field := objectType.Field(i + offset)
		if v, ok := field.Tag.Lookup("sql"); ok {
			buffer.WriteString(v)
		} else {
			buffer.WriteString(field.Name)
		}
		buffer.WriteString(fmt.Sprintf(" = $%d", i+firstNum))
	}
	return buffer.String()
}

func ForSelect(template interface{}, offset int, alias ...string) string {
	objectType := reflect.TypeOf(template)
	if len(alias) == 0 {
		return forSelect(objectType, nil, offset)
	} else {
		return forSelect(objectType, &alias[0], offset)
	}
}

func forSelect(objectType reflect.Type, alias *string, offset int) string {
	buffer := bytes.NewBufferString("")
	for i := 0; i < objectType.NumField()-offset; i++ {
		if i > 0 {
			buffer.WriteString(", ")
		}
		if alias != nil {
			buffer.WriteString(*alias)
			buffer.WriteString(".")
		}
		field := objectType.Field(i + offset)
		if v, ok := field.Tag.Lookup("sql"); ok {
			buffer.WriteString(v)
		} else {
			buffer.WriteString(field.Name)
		}
	}
	return buffer.String()
}

func ExecuteDatabase(config DatabaseConfig, delegate func(db *sql.DB)) {
	db := OpenDB(config)
	defer CloseDB(db)
	delegate(db)
}
