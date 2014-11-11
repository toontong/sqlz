package sqlz

/************
统计CREATE，SELECT，INSERT，DELETE，UPDATE 等的数量，并且统计操作的表的数量。
************/
import (
	///"github.com/golang/glog"
	"sync"
	"time"

	"sqlz/sqlparser"
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

func newCount() *Count {
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
		count = newCount()
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

		stmt, err := sqlparser.Parse(query)
		if err != nil {
			z_result.Error++
			continue
		}
		if !z_open {
			break
		}
		z_result.analyze(stmt)
		z_result.Success++
	}
	z_running <- false
}

func (res *StatusResult) analyze(stmt sqlparser.Statement) {

	switch sql := stmt.(type) {
	case *sqlparser.Select:
		for _, tableExpr := range sql.From {
			node, ok := tableExpr.(*sqlparser.AliasedTableExpr)
			if !ok {
				res.addOpration(SELECT, "")
			} else {
				res.addOpration(SELECT, sqlparser.GetTableName(node.Expr))
			}
		}
	case *sqlparser.Insert:
		res.addOpration(INSERT, sqlparser.GetTableName(sql.Table))
	case *sqlparser.Update:
		res.addOpration(UPDATE, sqlparser.GetTableName(sql.Table))
	case *sqlparser.Delete:
		res.addOpration(DELETE, sqlparser.GetTableName(sql.Table))
	case *sqlparser.Show:
		res.addOpration(SHOW, sql.Section)
	case *sqlparser.DDL:
		var typ SQL_Type
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
		res.addOpration(typ, string(sql.Table))
	case nil:
		res.addOpration(ERROR_SQL, "nil")
	default:
		res.addOpration(UNKNOW, "nil")
	}
}
