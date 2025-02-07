package agents

import (
	"context"
	"time"
)

type Agent interface {
	Run(ctx context.Context) error
	Name() string
	Initialize(ctx context.Context) error
}

type BaseAgent struct {
	name      string
	startTime time.Time
}

func NewBaseAgent(name string) *BaseAgent {
	return &BaseAgent{
		name:      name,
		startTime: time.Now(),
	}
}

func (b *BaseAgent) Name() string {
	return b.name
}

func (b *BaseAgent) Initialize(ctx context.Context) error {
	b.startTime = time.Now()
	return nil
}

func (b *BaseAgent) Run(ctx context.Context) error {
	return nil
}
