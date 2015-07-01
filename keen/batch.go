package keen

import (
	"github.com/facebookgo/muster"
	"github.com/mgutz/logxi/v1"
)

type batchEvent struct {
	Event      EventType
	Collection string
}

type musterBatch struct {
	Client *Client
	Events []batchEvent
}

func (b *musterBatch) Add(evt interface{}) {
	b.Events = append(b.Events, evt.(batchEvent))
}

func (b *musterBatch) Fire(notifier muster.Notifier) {
	defer notifier.Done()

	data := make(map[string][]EventType)

	for _, e := range b.Events {
		data[e.Collection] = append(data[e.Collection], e.Event)
	}

	var r interface{}
	_, err := b.Client.sling.Post("events").BodyJSON(data).ReceiveSuccess(&r)

	if err != nil {
		log.Warn(err.Error())
	}

}
