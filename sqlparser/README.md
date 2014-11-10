sqlparser
===
this folder code was from [mix](https://github.com/siddontang/mixer) which base on [vitess](https://github.com/youtube/vitess)

I was changed something.

Usage
---
	import (
		"sqlparser"		
	)
	
	stmt, err = sqlparser.Parse("SELECT * FROM t1")

	if err != nil{
        println("Parse success.")
    }

	switch sql := stmt.(type) {
	case *sqlparser.Select:
		typ = SELECT
	case *sqlparser.Insert:
		typ = INSERT
	case *sqlparser.DDL:
		switch sql.Action {
		case sqlparser.AST_CREATE:
		case sqlparser.AST_RENAME:
		default: println("do what you want with DDL SQL.")
		}
	....// case SQL other type.
    ....
    default:
     // do what you want.
    }