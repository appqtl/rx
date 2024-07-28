package rx

type Command uint8

const (
	DONE Command = iota
	PULL
	CANCEL
)

type EventType uint8

const (
	PUSH EventType = iota
	ERROR
	COMPLETE
)

type Event struct {
	Data     any
	Error    error
	Complete bool
}

func (e Event) Type() EventType {
	if e.Error != nil {
		return ERROR
	}
	if e.Complete {
		return COMPLETE
	}
	return PUSH
}

type Outlet interface {
	Push(any)
	Error(error)
	Complete()
}

type Inlet interface {
	Pull()
	Cancel()
}

type Pipe interface {
	Inlet
	Events() <-chan Event
	Close()
}

type pipe struct {
	commands chan Command
	events   chan Event
}

func newPipe() pipe {
	return pipe{
		commands: make(chan Command),  // TODO maybe we need a capacity
		events:   make(chan Event), // TODO maybe we need a capacity
	}
}

func (p pipe) Pull() {
	p.commands <- PULL
}

func (p pipe) Cancel() {
	p.commands <- CANCEL
}

func (p pipe) Events() <-chan Event {
	return p.events
}

func (p pipe) Close() {
	close(p.commands)
	// TODO do we need to send a CANCEL?
}

func (p pipe) Push(t any) {
	p.events <- Event{Data: t}
}

func (p pipe) Error(err error) {
	p.events <- Event{Error: err}
}

func (p pipe) Complete() {
	defer close(p.events)
	p.events <- Event{Complete: true}
}
