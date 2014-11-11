package sqlz

import (
	"encoding/json"
	"testing"
	"time"

	// "sqlz/sqlparser"
)

func TestSQLZ(t *testing.T) {
	tcases := []struct {
		desc       string
		query      string
		expectType SQL_Type
	}{
		{"create table sql.", "create table t1(id int, val int)", CREATE},
		{"select sql.", "select * from t1 where id = 2", SELECT},
		{"select sql.", "select * from t1,t2 where id = 2", SELECT},
		{"insert sql.", "inSERT INTO t1(id, val) values(123,456),(678,90)", INSERT},
		{"delete sql.", "delete from t1 where id=3", DELETE},
		{"delete sql.", "delete from T1 where id=1", DELETE},
		{"delete sql.", "delete from `t222` where id=1", DELETE},
		{"update sql.", "update t1 set val=333 where id=1", UPDATE},
		{"drop table sql.", "drop table t1", DROP},
		{"show tables", "show tables", SHOW},
		{"error sql", "error sql", ERROR_SQL},
		{"unknow sql", "set @@a=1", UNKNOW},
	}

	StartZ()
	for _, tcase := range tcases {
		Z(tcase.query)
	}

	time.Sleep(time.Second)

	st := Status()
	byt, err := json.Marshal(st)
	if err != nil {
		t.Error(err.Error())
	} else {
		println(string(byt))
	}

	StopZ()
	StopZ()
}
