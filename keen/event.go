package keen

import "time"

type EventType interface {
	SetTimestamp(time.Time)
}

type keenProps struct {
	Timestamp string `json:"timestamp"`
}

type Event struct {
	Keen *keenProps `json:"keen"`
}

func (evt *Event) SetTimestamp(t time.Time) {
	if evt.Keen == nil {
		evt.Keen = &keenProps{}
	}
	evt.Keen.Timestamp = t.UTC().Format("2006-01-02T15:04:05.000Z")
}
