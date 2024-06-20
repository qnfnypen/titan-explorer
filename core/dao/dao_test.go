package dao

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
)

func TestMain(m *testing.M) {
	db, _ := sqlx.Connect("mysql", "root:123456@tcp(localhost:3306)/titan-explorer?charset=utf8mb4&parseTime=true&loc=Local")
	DB = db

	m.Run()
}

func TestMoveBackDeletedDevice(t *testing.T) {
	err := MoveBackDeletedDevice(context.Background(), []string{"1"}, "1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetNodeNums(t *testing.T) {
	on, ab, off, del, err := GetNodeNums(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(on, ab, off, del)
}
