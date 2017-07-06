package db

import (
	"fmt"
	"testing"
)

type lightningBolt struct {
	Intensity string
	Distance  string
}

func TestBoltStore(t *testing.T) {
	bucketName := "test-bolt-store"
	dbFile := bucketName + ".db"

	store, err := NewBoltStore(bucketName, dbFile)
	if err != nil {
		t.Fatal(err)
	}

	if store.BucketName != bucketName {
		t.Errorf("expected store.BucketName to be '%s', got '%s'", bucketName, store.BucketName)
	}

	if store.DBFile != dbFile {
		t.Errorf("expected store.DBFile to be '%s', got '%s'", dbFile, store.DBFile)
	}

	bolts := []*lightningBolt{
		&lightningBolt{"high", "near"},
		&lightningBolt{"medium", "near"},
		&lightningBolt{"low", "far"},
	}

	err = store.Write("foo", bolts)
	if err != nil {
		t.Fatal(err)
	}

	bolts = make([]*lightningBolt, 0)
	err = store.Read("foo", &bolts)
	if err != nil {
		t.Fatal(err)
	}

	if len(bolts) != 3 {
		t.Errorf("expecting %d bolts, got %d", 3, len(bolts))
	}

	summary := ""
	for _, b := range bolts {
		summary += fmt.Sprintf("%s/%s ", b.Intensity, b.Distance)
	}

	expectedSummary := "high/near medium/near low/far "
	if summary != expectedSummary {
		t.Errorf("expecting '%s' to equal '%s'", summary, expectedSummary)
	}

	err, ok := store.HasData("foo")
	if err != nil {
		t.Fatal(err)
	}
	if ok != true {
		t.Error("expect HasData to be true")
	}

	err = store.Delete("foo")
	if err != nil {
		t.Fatal(err)
	}

	err, ok = store.HasData("foo")
	if err != nil {
		t.Fatal(err)
	}
	if ok != false {
		t.Error("expect HasData to be false")
	}

	bolts = make([]*lightningBolt, 0)
	err = store.Read("foo", &bolts)
	if err == nil {
		t.Fatal("expecting error, got none")
	}
}
