package db

import "testing"

func TestDBNameFromDSN(t *testing.T) {
	got := dbNameFromDSN("host=127.0.0.1 port=5432 user=root password=root dbname=integration sslmode=disable")
	if got != "integration" {
		t.Fatalf("got %q", got)
	}
	got = dbNameFromDSN("postgres://root:root@127.0.0.1:5432/integration?sslmode=disable")
	if got != "integration" {
		t.Fatalf("url got %q", got)
	}
}

func TestWithDBName(t *testing.T) {
	got := withDBName("host=127.0.0.1 dbname=integration sslmode=disable", "postgres")
	if got != "host=127.0.0.1 dbname=postgres sslmode=disable" {
		t.Fatalf("got %q", got)
	}
}
