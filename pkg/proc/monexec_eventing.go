package proc

import (
	"github.com/reddec/monexec/pool"
	"context"
)

type noopListener struct {
}

func (l noopListener) OnSpawned(ctx context.Context, in pool.Instance) {

}

func (l noopListener) OnStarted(ctx context.Context, in pool.Instance) {

}

func (l noopListener) OnStopped(ctx context.Context, in pool.Instance, err error) {

}

func (l noopListener) OnFinished(ctx context.Context, in pool.Instance) {

}

type stopListener struct {
	ctx context.Context
	cancel context.CancelFunc

	noopListener
}

func (l *stopListener) OnStopped(ctx context.Context, in pool.Instance, err error) {
	_, cancel := l.currentCtx()
	cancel()
}

func (l *stopListener) currentCtx() (context.Context, context.CancelFunc) {
	if l.ctx == nil {
		chld, cancel := context.WithCancel(context.Background())
		l.ctx = chld
		l.cancel = cancel
	}
	return l.ctx, l.cancel
}

func (l *stopListener) WaitUntilStop() context.Context {
	curCtx, _ := l.currentCtx()
	chldCtx, _ := context.WithCancel(curCtx)
	return chldCtx
}
