package dao

import (
	"context"
	"testing"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
)

func TestMain(m *testing.M) {
	db, err := sqlx.Connect("mysql", "root:abcd1234@tcp(localhost:8080)/titan_explorer?charset=utf8mb4&parseTime=true&loc=Local")
	if err != nil {
		panic(err)
	}

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

func TestDeleteUserGroupAsset(t *testing.T) {
	ctx := context.Background()
	userID := "0x5e48ee53a85343b7b57014a1eb20e21fff92d4a4"
	gids := []int64{}

	err := DeleteUserGroupAsset(ctx, userID, gids)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	userID := "0x5e48ee53a85343b7b57014a1eb20e21fff92d4a4"
	query, args, err := squirrel.Update(tableNameUser).Set("used_storage_size", squirrel.Expr("GREATEST(used_storage_size - ?,0)", 1000)).Where("username = ?", userID).ToSql()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(query, args)
}
