package keen_test

import (
	"testing"

	"github.com/drinkin/keen-go/keen"
	"github.com/stretchr/testify/require"
)

type ExampleEvent struct {
	Name string `json:"name"`
	keen.Event
}

func TestClient(t *testing.T) {
	assert := require.New(t)

	client := keen.FromEnv()

	err := client.Track("test", &ExampleEvent{
		Name: "hi",
	})
	assert.NoError(err)
	err = client.Track("test", &ExampleEvent{
		Name: "hi2",
	})

	assert.NoError(err)

	assert.NoError(client.Stop())
}
