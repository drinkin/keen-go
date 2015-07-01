package keen

import (
	"sync"
	"time"

	"github.com/facebookgo/muster"
)

const (
	defaultPendingWorkCapacity = 1000
	defaultBatchTimeout        = time.Millisecond * 10
	defaultMaxBatchSize        = 50
)

type Client struct {
	// Capacity of log channel. Defaults to 1000.
	PendingWorkCapacity uint

	// Maximum number of items in a batch. Defaults to 50.
	MaxBatchSize uint

	// Amount of time after which to send a pending batch. Defaults to 10ms.
	BatchTimeout time.Duration

	startOnce sync.Once
	startErr  error
	muster    muster.Client
}

func (c *Client) Track(cn string, data EventType) error {
	return c.TrackWithTimestamp(cn, time.Now(), data)
}

func (c *Client) TrackWithTimestamp(cn string, t time.Time, data EventType) error {
	if err := c.start(); err != nil {
		return err
	}
	data.SetTimestamp(t)

	c.muster.Work <- &batchEvent{
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
