package keen

import (
	"sync"
	"time"

	"github.com/drinkin/shop/src/lg"
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

type Event struct {
	Collection string
	Data       interface{}

	Timestamp time.Time
}

func (c *Client) Track(cn string, data interface{}) error {
	if err := c.start(); err != nil {
		return err
	}

	evt := &Event{
		Collection: cn,
		Data:       data,
		Timestamp:  time.Now(),
	}
	c.muster.Work <- evt
	return nil
}

type EventWithTimestamp struct {
}

type musterBatch struct {
	Client *Client
	Events []*Event
}

func (b *musterBatch) Add(evt interface{}) {
	b.Events = append(b.Events, evt.(*Event))
}

func (b *musterBatch) Fire(notifier muster.Notifier) {
	defer notifier.Done()
	for _, e := range b.Events {
		lg.Pretty(e.Data)
	}
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

// Stop and gracefully wait for the background worker to finish processing
// pending requests.
func (c *Client) Stop() error {
	if err := c.start(); err != nil {
		return err
	}
	return c.muster.Stop()
}
