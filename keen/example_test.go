package keen_test

import (
	"testing"

	"github.com/drinkin/keen-go/keen"
	"github.com/stretchr/testify/require"
)

type ExampleEvent struct {
	Name string `json:"name"`
}

func TestClient(t *testing.T) {
	assert := require.New(t)
	assert.Equal(1, 1)

	client := &keen.Client{}

	err := client.Track("test", &ExampleEvent{"hi"})

	assert.NoError(err)

	assert.NoError(client.Stop())
}
