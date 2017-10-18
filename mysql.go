package main

import (
	"fmt"
	"strings"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jhunt/vcaptive"

	"github.com/starkandwayne/shield-cf-demo/internal/rand"
)

type MySQLSystem struct {
	mysql *sql.DB
}

func (s *MySQLSystem) Configure(services vcaptive.Services) (bool, error) {
	svc, found := services.Tagged("mysql")
	if !found {
		return false, nil
	}

	uri, ok := svc.GetString("uri")
	if !ok {
		return true, fmt.Errorf("VCAP_SERVICES: '%s' service has no 'uri' credential", svc.Label)
	}

	uri = strings.TrimPrefix(uri, "mysql://")
	db, err := sql.Open("mysql", uri)
	if err != nil {
		return true, err
	}

	if err := db.Ping(); err != nil {
		return true, err
	}

	s.mysql = db
	return true, nil
}

func (s *MySQLSystem) Setup() error {
	if _, err := s.mysql.Exec(`CREATE TABLE IF NOT EXISTS dat (v INT(11) NOT NULL)`); err != nil {
		return err
	}

	if _, err := s.mysql.Exec(`CREATE TABLE IF NOT EXISTS verify (s VARCHAR(30) NOT NULL)`); err != nil {
		return err
	}

	r, err := s.mysql.Query(`SELECT * FROM verify LIMIT 1`)
	if err != nil {
		return err
	}

	if !r.Next() {
		if _, err := s.mysql.Exec(`INSERT INTO verify (s) VALUES (?)`, rand.String(16)); err != nil {
			return err
		}
	}

	for i := 0; i < rand.Bound(100000, 20000); i++ {
		if _, err := s.mysql.Exec(`INSERT INTO dat (v) VALUES (?)`, i); err != nil {
			return err
		}
	}

	return nil
}

func (s *MySQLSystem) Teardown() error {
	if _, err := s.mysql.Exec(`DROP TABLE dat`); err != nil {
		return err
	}

	if _, err := s.mysql.Exec(`DROP TABLE verify`); err != nil {
		return err
	}

	return nil
}

func (s *MySQLSystem) Summarize() Data {
	dat := Data{
		System:       "MySQL",
		Summary:      "              *no data found*\n",
		Verification: "UNKNOWN",
		OK:           false,
	}

	r, err := s.mysql.Query(`SELECT COUNT(v) FROM dat`)
	if err != nil {
		dat.Summary = fmt.Sprintf("ERROR:        *%s*\n", err)
		return dat
	}
	if !r.Next() {
		return dat
	}

	var n int
	if err := r.Scan(&n); err != nil {
		return dat
	}

	r, err = s.mysql.Query(`SELECT s FROM verify`)
	if err != nil {
		dat.Summary = fmt.Sprintf("ERROR:        *%s*\n", err)
		return dat
	}
	if !r.Next() {
		return dat
	}

	var vfy string
	if err := r.Scan(&vfy); err != nil {
		dat.Summary = fmt.Sprintf("ERROR:        *%s*\n", err)
		return dat
	}

	dat.Summary = fmt.Sprintf("Records:      *%d*\n", n)
	dat.Verification = vfy
	return dat
}

func init() {
	if Systems == nil {
		Systems = make(map[string]System)
	}
	Systems["mysql"] = &MySQLSystem{}
}
