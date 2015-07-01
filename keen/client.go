package keen

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/dghubble/sling"
	"github.com/drinkin/di/env"
	"github.com/facebookgo/muster"
)

const (
	defaultPendingWorkCapacity = 1000
	defaultBatchTimeout        = time.Millisecond * 200
	defaultMaxBatchSize        = 50
)

type Client struct {
	// Capacity of log channel. Defaults to 1000.
	PendingWorkCapacity uint

	// Maximum number of items in a batch. Defaults to 50.
	MaxBatchSize uint

	// Amount of time after which to send a pending batch. Defaults to 10ms.
	BatchTimeout time.Duration

	ProjectId  string
	APIKey     string
	HttpClient *http.Client

	startOnce sync.Once
	startErr  error

	muster muster.Client
	sling  *sling.Sling
}

func New(project_id, api_key string) *Client {
	return &Client{
		ProjectId: project_id,
		APIKey:    api_key,
	}
}

func FromEnv() *Client {
	return New(env.MustGet("KEEN_PROJECT_ID"), env.MustGet("KEEN_API_KEY"))
}

func (c *Client) Track(cn string, data EventType) error {
	return c.TrackWithTimestamp(cn, time.Now(), data)
}

func (c *Client) TrackWithTimestamp(cn string, t time.Time, data EventType) error {
	if err := c.start(); err != nil {
		return err
	}
	data.SetTimestamp(t)

	c.muster.Work <- batchEvent{
		Event:      data,
		Collection: cn,
	}
	return nil
}

// Stop and gracefully wait for the background worker to finish processing
// pending requests.
func (c *Client) Stop() error {
	if err := c.start(); err != nil {
		return err
	}
	return c.muster.Stop()
}

func (c *Client) start() error {
	c.startOnce.Do(func() {
		// Setup sling
		url := fmt.Sprintf("https://api.keen.io/3.0/projects/%s/", c.ProjectId)
		c.sling = sling.New().Client(c.HttpClient).Base(url).Set("Authorization", c.APIKey)

		pendingWorkCapacity := c.PendingWorkCapacity
		if pendingWorkCapacity == 0 {
			pendingWorkCapacity = defaultPendingWorkCapacity
		}
		maxBatchSize := c.MaxBatchSize
		if maxBatchSize == 0 {
			maxBatchSize = defaultMaxBatchSize
		}
		batchTimeout := c.BatchTimeout
		if int64(batchTimeout) == 0 {
			batchTimeout = defaultBatchTimeout
		}

		c.muster.BatchMaker = func() muster.Batch { return &musterBatch{Client: c} }
		c.muster.BatchTimeout = batchTimeout
		c.muster.MaxBatchSize = maxBatchSize
		c.muster.PendingWorkCapacity = pendingWorkCapacity
		c.startErr = c.muster.Start()
	})
	return c.startErr
}
