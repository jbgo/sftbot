package db

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	require.Nil(t, err)

	assert.Equal(t, bucketName, store.BucketName)
	assert.Equal(t, dbFile, store.DBFile)

	bolts := []*lightningBolt{
		&lightningBolt{"high", "near"},
		&lightningBolt{"medium", "near"},
		&lightningBolt{"low", "far"},
	}

	err = store.Write("foo", bolts)
	require.Nil(t, err)

	bolts = make([]*lightningBolt, 0)
	err = store.Read("foo", &bolts)
	require.Nil(t, err)

	assert.Equal(t, 3, len(bolts))

	summary := ""
	for _, b := range bolts {
		summary += fmt.Sprintf("%s/%s ", b.Intensity, b.Distance)
	}

	assert.Equal(t, "high/near medium/near low/far ", summary)

	err, ok := store.HasData("foo")
	require.Nil(t, err)
	assert.True(t, ok)

	err = store.Delete("foo")
	require.Nil(t, err)

	err, ok = store.HasData("foo")
	require.Nil(t, err)
	assert.False(t, ok)

	bolts = make([]*lightningBolt, 0)
	err = store.Read("foo", &bolts)
	require.NotNil(t, err)
}
