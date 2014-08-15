package sqlite_test

import (
	"database/sql"
	"testing"

	. "github.com/mix3/go-tenco/storage/sqlite"
	"github.com/stretchr/testify/assert"
)

func TestSqlite(t *testing.T) {
	s, err := New(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	func() {
		err = s.Set("hoge", "http://localhost:5001")
		if err != nil {
			t.Fatal(err)
		}
		err = s.Set("fuga", "http://localhost:5002")
		if err != nil {
			t.Fatal(err)
		}
	}()
	func() {
		m, err := s.Map()
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, map[string]string{
			"hoge": "http://localhost:5001",
			"fuga": "http://localhost:5002",
		}, m)
	}()
	func() {
		d, err := s.Get("hoge")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "http://localhost:5001", d)
	}()
	func() {
		d, err := s.Get("fuga")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, "http://localhost:5002", d)
	}()
	func() {
		s.Delete("hoge")
		_, err := s.Get("hoge")
		assert.Equal(t, sql.ErrNoRows, err)
	}()
	func() {
		s.DeleteAll()
		_, err := s.Get("fuga")
		assert.Equal(t, sql.ErrNoRows, err)
	}()
}
