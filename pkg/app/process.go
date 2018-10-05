package app

import (
	"time"
	"github.com/reddec/monexec/monexec"
	"github.com/reddec/monexec/pool"
	"fmt"
	"context"
	"errors"
)

type process struct {
	cmd string
	arg []string
	stopGracePeriod time.Duration

	stop context.CancelFunc
	mon *monexec.Config
	ctx context.Context

	waitUntilStop context.Context
	stopListener  *stopListener
}

func (p *process) start() error {
	if p.stop != nil {
		return errors.New("tried to start before stopping")
	}

	if p.mon == nil {
		p.mon = &monexec.Config{
			Services: []pool.Executable{{
				Command: p.cmd,
				Args:    p.arg,
				Restart: -1,
				RestartTimeout: p.stopGracePeriod,
			},},
		}
	}

	fmt.Println("starting app...")

	sv := &pool.Pool{}
	p.stopListener = &stopListener{}
	sv.Watch(p.stopListener)
	p.waitUntilStop = p.stopListener.WaitUntilStop()
	ctx, stop := context.WithCancel(context.Background())

	if err := p.mon.Run(sv, ctx); err != nil {
		return err
	}

	// This is useless as it emits immediately after we call `stop()`
	p.ctx = ctx

	p.stop = stop

	return nil
}

func (p *process) doStop() {
	fmt.Println("stopping app...")
	if p.stop != nil {
		p.stop()
		p.stop = nil

		select {
		case <-p.waitUntilStop.Done():
			fmt.Println("done waiting until app stops")
			p.waitUntilStop = nil
		}
	}
}

func (p *process) restart() error {
	p.doStop()

	if err := p.start(); err != nil {
		return err
	}
	return nil
}
