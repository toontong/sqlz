package sqlz

/************
统计CREATE，SELECT，INSERT，DELETE，UPDATE 等的数量，并且统计操作的表的数量。
************/
import (
	///"github.com/golang/glog"
	"sync"
	"time"

	"github.com/toontong/sqlz/sqlparser"
)

var (
	z_open       bool
	z_running    chan bool
	z_result     *StatusResult
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

func NewCount() *Count {
	c := &Count{
		TableCount: make(map[string]int64),
	}
	return c
}

func (count *Count) add(tableName string) int64 {
	if cnt, ok := count.TableCount[tableName]; ok {
		count.TableCount[tableName]++
		return cnt + 1
	} else {
		count.TableCount[tableName] = 1
		return 1
	}
}

type StatusResult struct {
	Opration map[SQL_Type]*Count
	Error    int
	Success  int
	Waiting  int
	Start    time.Time
	End      time.Time
	lock     *sync.RWMutex
}

func newStatusResult() *StatusResult {
	status := &StatusResult{
		Opration: make(map[SQL_Type]*Count),
		Start:    time.Now(),
		Error:    0,
		Success:  0,
		Waiting:  0,
		lock:     new(sync.RWMutex),
	}
	return status
}

func (res *StatusResult) addOpration(typ SQL_Type, tableName string) {
	z_result.lock.Lock()
	defer z_result.lock.Unlock()

	if count, ok := z_result.Opration[typ]; ok {
		count.add(tableName)
	} else {
		count = NewCount()
		count.add(tableName)
		z_result.Opration[typ] = count
	}
}

// -----------------------------
// 开启统计功能,可随时开启
func StartZ() {
	if z_open {
		return
	}
	z_open = true

	cleanStatus()
	z_stackQuery = make(chan string, 4096)
	go statistics()
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
	if !z_open {
		return
	}
	z_open = false

	// wait statistics thread exit
	<-z_running

	z_stackQuery = nil
}

func Status() StatusResult {
	if z_result == nil {
		cleanStatus()
	}
	z_result.lock.Lock()
	defer z_result.lock.Unlock()
	z_result.Waiting = len(z_stackQuery)
	z_result.End = time.Now()
	return *z_result
}

// -----------------------------
func cleanStatus() {
	z_result = newStatusResult()
}

func init() {
	cleanStatus()
	z_running = make(chan bool, 1)
}

func statistics() {
	var query string
	for z_open {

		select {
		case query = <-z_stackQuery:
		default:
			continue
		}

		typ, table := ParseQuery(query)
		if !z_open {
			break
		}

		if typ == ERROR_SQL {
			z_result.Error++
			continue
		}
		z_result.addOpration(typ, table)
		z_result.Success++
	}
	z_running <- false
}

func ParseQuery(query string) (action SQL_Type, table string) {
	// stmt sqlparser.Statement
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return ERROR_SQL, string(UNKNOW)
	}

	switch sql := stmt.(type) {
	case *sqlparser.Select:
		action = SELECT
		for _, tableExpr := range sql.From {
			node, ok := tableExpr.(*sqlparser.AliasedTableExpr)
			if !ok {
				table = string(UNKNOW)
			} else {
				table = sqlparser.GetTableName(node.Expr)
			}
		}
	case *sqlparser.Insert:
		action, table = INSERT, sqlparser.GetTableName(sql.Table)
	case *sqlparser.Update:
		action, table = UPDATE, sqlparser.GetTableName(sql.Table)
	case *sqlparser.Delete:
		action, table = DELETE, sqlparser.GetTableName(sql.Table)
	case *sqlparser.Show:
		action, table = SHOW, sql.Section
	case *sqlparser.DDL:

		var tableName []byte
		switch sql.Action {
		case sqlparser.AST_CREATE:
			action = CREATE
			tableName = sql.NewName
		case sqlparser.AST_RENAME:
			action = RENAME
			tableName = sql.Table
		case sqlparser.AST_DROP:
			action = DROP
			tableName = sql.Table
		case sqlparser.AST_ALTER:
			action = ALTER
			tableName = sql.Table
		default:
			action = UNKNOW
		}
		table = string(tableName)
	case nil:
		action, table = ERROR_SQL, "nil"
	default:
		action, table = UNKNOW, "nil"
	}
	return action, table
}
