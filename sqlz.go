package sqlz

/************
统计CREATE，SELECT，INSERT，DELETE，UPDATE 等的数量，并且统计操作的表的数量。
************/
import (
	///"github.com/golang/glog"
	"time"

	"sqlz/sqlparser"
)

var (
	z_open       bool
	z_result     StatusResult
	z_stackQuery chan string
)

type SQL_Type string

const (
	ERROR_SQL SQL_Type = "ERROR_SQL"
	UNKNOW    SQL_Type = "UNKNOW_SQL"
	SELECT    SQL_Type = "SELECT"
	INSERT    SQL_Type = "INSERT"
	UPDATE    SQL_Type = "UPDATE"
	DELETE    SQL_Type = "DELETE"
	SHOW      SQL_Type = "SHOW"
	CREATE    SQL_Type = "CREATE"
	RENAME    SQL_Type = "RENAME"
	ALTER     SQL_Type = "ALTER"
	DROP      SQL_Type = "DROP"
)

type Count struct {
	// Table      string
	// Count      int64
	TableCount map[string]int64 //key is table name.
}

type StatusResult struct {
	Opration map[SQL_Type]Count
	Start    time.Time
	End      time.Time
}

// 开启统计功能,可随时开启
func StartZ() {
	z_open = true
	cleanStatus()
	z_stackQuery = make(chan string, 4096)
}

func Z(query string) bool {
	if !z_open {
		return false
	}
	select {
	case z_stackQuery <- query:
		return true
	default:
		// glog.Error("query stack was full or empty.")
		return false
	}
}

func StopZ() {
	z_open = false
}

func Status() StatusResult {
	return StatusResult{}
}

func cleanStatus() {
}

func init() {

}
func getSqlType(query string) SQL_Type {
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return ERROR_SQL
	}

	typ := UNKNOW
	switch sql := stmt.(type) {
	case *sqlparser.Select:
		typ = SELECT
	case *sqlparser.Insert:
		typ = INSERT
	case *sqlparser.Update:
		typ = UPDATE
	case *sqlparser.Delete:
		typ = DELETE
	case *sqlparser.Show:
		typ = SHOW
	case *sqlparser.DDL:
		switch sql.Action {
		case sqlparser.AST_CREATE:
			typ = CREATE
		case sqlparser.AST_RENAME:
			typ = RENAME
		case sqlparser.AST_DROP:
			typ = DROP
		case sqlparser.AST_ALTER:
			typ = ALTER
		default:
			typ = UNKNOW
		}
	default:
		typ = UNKNOW
	}
	return typ
}
