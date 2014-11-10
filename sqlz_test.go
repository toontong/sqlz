package sqlz

import (
	"log"
	"os"
	"testing"
)

var (
	user      string
	pass      string
	prot      string
	addr      string
	dbname    string
	dsn       string
	netAddr   string
	available bool
)

var logger = log.New(os.Stdout, "sqlz", log.Lshortfile|log.LstdFlags)

// See https://github.com/go-sql-driver/mysql/wiki/Testing
func init() {
	// get environment variables
	env := func(key, defaultValue string) string {
		if value := os.Getenv(key); value != "" {
			return value
		}
		return defaultValue
	}
	user = env("MYSQL_TEST_USER", "qing")
	pass = env("MYSQL_TEST_PASS", "admin")
	prot = env("MYSQL_TEST_PROT", "tcp")
	addr = env("MYSQL_TEST_ADDR", "10.20.187.81:3306")
	dbname = env("MYSQL_TEST_DBNAME", "gotest")
	netAddr = fmt.Sprintf("%s(%s)", prot, addr)
	dsn = fmt.Sprintf("%s:%s@%s/%s?timeout=30s&strict=true", user, pass, netAddr, dbname)
	logger.Println(dsn)
	c, err := net.Dial(prot, addr)
	if err == nil {
		available = true
		c.Close()
		logger.Fatal("error on connect:%s", err.Error())
	}
}
