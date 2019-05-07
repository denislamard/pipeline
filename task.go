package pipeline

import (
	"context"
	"github.com/google/uuid"
)

type Executor interface {
	ID() string
	Name() string
	Context() context.Context
	Init(pipeline *Pipeline, name string)
	Run()
}

type Task struct {
	id       uuid.UUID
	name     string
	pipeline *Pipeline
}

func (t *Task) makeID() {
	t.id = uuid.New()
}

func (t *Task) ID() string {
	return t.id.String()
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) Context() context.Context {
	return t.pipeline.Context()
}

func (t *Task) Init(pipeline *Pipeline, name string) {
	t.pipeline = pipeline
	t.name = name
	t.makeID()
}

func (t *Task) Done() <-chan struct{} {
	if t.pipeline != nil {
		return t.pipeline.Done()
	}
	return nil
}

func (t *Task) TearDown() {
	if t.pipeline != nil {
		t.pipeline.TearDown(t.name, t.id)
	}
}

