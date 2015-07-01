package keen

import (
	"github.com/drinkin/shop/src/lg"
	"github.com/facebookgo/muster"
)

type batchEvent struct {
	Event      EventType
	Collection string
}

type musterBatch struct {
	Client *Client
	Events []*batchEvent
}

func (b *musterBatch) Add(evt interface{}) {
	b.Events = append(b.Events, evt.(*batchEvent))
}

func (b *musterBatch) Fire(notifier muster.Notifier) {
	defer notifier.Done()

	data := make(map[string][]EventType)

	for _, e := range b.Events {
		data[e.Collection] = append(data[e.Collection], e.Event)
	}
	lg.Pretty(data)
}
