package main

import "testing"

func TestGetUserAssetAreas(t *testing.T) {
	Init()

	uas := getUserAssetAreas(uid)
	t.Log(uas)
}
