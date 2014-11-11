sqlz
====

用于go 中 统计sql情况：统计select,Insert,delete,update,次数，对应操作表的次数。

statistics the SQL count of the: select, Insert, delete, update... and the opration table.

Usage
---


	improt (
		"encoding/json"
		"testing"
		"time"

		"sqlz"
	)

	func TestSQLZ(t *testing.T) {
		tcases := []struct {
			desc       string
			query      string
		}{
			{"create table sql.", "create table t1(id int, val int)"},
			{"select sql.", "select * from t1 where id = 2"},
			{"select sql.", "select * from t1,t2 where id = 2"T},
			{"insert sql.", "inSERT INTO t1(id, val) values(123,456),(678,90)"},
			{"delete sql.", "delete from t1 where id=3"},
			{"update sql.", "update t1 set val=333 where id=1"},
			{"drop table sql.", "drop table t1"},
			{"show tables", "show tables"},
			{"error sql", "error sql"},
			{"unknow sql", "set @@a=1"},
		}
	
		sqlz.StartZ()
		for _, tcase := range tcases {
			sqlz.Z(tcase.query)
		}
	
		time.Sleep(time.Second)
	
		st := sqlz.Status()
		byt, err := json.Marshal(st)
		if err != nil {
			t.Error(err.Error())
		} else {
			println(string(byt))
		}
	
		sqlz.StopZ()
	}

Ouput
---
	{"End": "2014-11-11T17:41:39.0728559+08:00",
	 "Error": 1,
	 "Opration": {"CREATE": {"TableCount": {"t1": 1}},
	              "DELETE": {"TableCount": {"t1": 1}},
	              "DROP": {"TableCount": {"t1": 1}},
	              "INSERT": {"TableCount": {"t1": 1}},
	              "SELECT": {"TableCount": {"t1": 2, "t2": 1}},
	              "SHOW": {"TableCount": {"tables": 1}},
	              "UNKNOW_SQL": {"TableCount": {"nil": 1}},
	              "UPDATE": {"TableCount": {"t1": 1}}},
	 "Start": "2014-11-11T17:41:38.0727987+08:00",
	 "Success": 9,
	 "Waiting": 0}

Status:
---
Stable

but nut support **drop tables tb1,tb2 ...** SQL. Just statistics the tb1, not include tb2.