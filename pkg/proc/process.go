package proc

import (
	"time"
	"github.com/reddec/monexec/monexec"
	"github.com/reddec/monexec/pool"
	"fmt"
	"context"
	"errors"
)

type Process struct {
	Command         string
	Args            []string
	StopGracePeriod time.Duration

	StopAsync context.CancelFunc

	mon *monexec.Config
	ctx context.Context

	waitUntilStop context.Context
	stopListener  *stopListener
}

func (p *Process) Start() error {
	if p.StopAsync != nil {
		return errors.New("tried to start before stopping")
	}

	if p.mon == nil {
		p.mon = &monexec.Config{
			Services: []pool.Executable{{
				Command: p.Command,
				Args:    p.Args,
				Restart: -1,
				RestartTimeout: p.StopGracePeriod,
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

	p.StopAsync = stop

	return nil
}

func (p *Process) Stop() {
	fmt.Println("stopping app...")
	if p.StopAsync != nil {
		p.StopAsync()
		p.StopAsync = nil

		select {
		case <-p.waitUntilStop.Done():
			fmt.Println("done waiting until app stops")
			p.waitUntilStop = nil
		}
	}
}

func (p *Process) Restart() error {
	p.Stop()

	if err := p.Start(); err != nil {
		return err
	}
	return nil
}
