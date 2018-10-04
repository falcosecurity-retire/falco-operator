package app

import (
	"github.com/reddec/monexec/monexec"
	"github.com/radovskyb/watcher"
	"github.com/tidwall/sjson"
	"github.com/spf13/afero"
	"time"
	"strings"
	"fmt"
	"log"
	"os"
	"os/signal"
	"errors"
	"github.com/reddec/monexec/pool"
	"context"
	"syscall"
	"path/filepath"
	"github.com/ghodss/yaml"
	"text/template"
	"bytes"
)

type SuperviseOpts struct {
	RestarGracePeriod time.Duration
	WatchInterval     time.Duration
	StopGracePeriod   time.Duration
	Set               []string
	Watch             []string
	Restart           bool
}

type Set struct {
	Path string
	Key string
	Value string
}

type Supervisor struct {
	SuperviseOpts
	fs afero.Fs
}

type templateData struct {
	Path string
	Content string
}

func (o *Supervisor) DoSet(mem map[string]string) error {
	for _, s := range o.Set {
		var err error

		set, err := parseSet(s)
		if err != nil {
			return err
		}

		var bytesData []byte

		bytesData, err = afero.ReadFile(o.fs, fmt.Sprintf("%s.orig", set.Path))
		if err != nil {
			bytesData, err = afero.ReadFile(o.fs, set.Path)
			if err != nil {
				return err
			}
			if err := afero.WriteFile(o.fs, fmt.Sprintf("%s.orig", set.Path), bytesData, 0644); err != nil {
				return err
			}
		}

		ext := filepath.Ext(set.Path)
		if ext == ".json" {

		} else if ext == ".yaml" {
			bytesData, err = yaml.YAMLToJSON(bytesData)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unexpected ext: %s", ext)
		}

		doc := string(bytesData)

		if !strings.HasSuffix(set.Key, "-1") && len(mem) > 1 {
			return fmt.Errorf("the key suffix must be \"-`\" in order to append multiple items")
		}

		for path, content := range mem {
			tpl, err := template.New("valuetemplate").Parse(set.Value)
			if err != nil {
				return err
			}

			buf := new(bytes.Buffer)
			if err := tpl.Execute(buf, templateData{Path: path, Content: content,}); err != nil {
				return err
			}

			doc, err = sjson.Set(doc, set.Key, buf.String())
			if err != nil {
				return err
			}
		}

		result := []byte(doc)
		if ext == ".yaml" {
			result, err = yaml.JSONToYAML(result)
		}

		err = afero.WriteFile(o.fs, set.Path, append(result, []byte("\n")...), 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseSet(s string) (*Set, error) {
	parsed := strings.Split(s, "=")
	if len(parsed) != 3 {
		return nil, fmt.Errorf("unexpected 3 items, but got %d items: %v", len(parsed), parsed)
	}
	return &Set{
		parsed[0], parsed[1], parsed[2],
	}, nil
}

type monitoredProc struct {
	cmd string
	arg []string
	stopGracePeriod time.Duration

	stop context.CancelFunc
	mon *monexec.Config
	ctx context.Context

	waitUntilStop context.Context
	stopListener  *stopListener
}

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

func (p *monitoredProc) start() error {
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

func (p *monitoredProc) doStop() {
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

func (p *monitoredProc) restart() error {
	p.doStop()

	if err := p.start(); err != nil {
		return err
	}
	return nil
}

func Supervise(args []string, opts SuperviseOpts) error {
	if len(args) == 0 {
		return errors.New("missing args")
	}
	if opts.RestarGracePeriod <= opts.WatchInterval {
		return errors.New("watch interval should be less than restart grace period")
	}
	cmd := args[0]
	arg := args[1:]

	proc := &monitoredProc{cmd: cmd, arg: arg, stopGracePeriod: opts.StopGracePeriod,}

	err := proc.start()
	if err != nil {
		return err
	}

	sp := &Supervisor{opts, afero.NewOsFs()}

	w := watcher.New()
	w.IgnoreHiddenFiles(true)

	mem := map[string]string{}

	done := make(chan struct{})
	go func() {
		var restartGrace <-chan time.Time
		var t *time.Timer

		defer close(done)

		for {
			select {
			case event := <-w.Event:
				fmt.Println(event)

				paths := []string{}

				if event.IsDir() {
					if event.Op == watcher.Remove {
						fmt.Printf("dir %s has been removed\n", event.Path)
						for k, _ := range mem {
							if strings.HasPrefix(k, event.Path) {
								fmt.Printf("marking %s as removed\n", k)
								delete(mem, k)
							}
						}
					} else {
						fmt.Printf("%s has been updated\n", event.Path)
						paths = append(paths, event.Path)
					}
				} else {
					if event.Op == watcher.Remove {
						fmt.Printf("%s has been removed\n", event.Path)
						delete(mem, event.Path)
					} else {
						fmt.Printf("%s has been updated\n", event.Path)
						paths = append(paths, event.Path)
					}
				}

				for _, path := range paths {
					content, err := afero.ReadFile(sp.fs, path)
					if err != nil {
						panic(err)
					}
					mem[path] = string(content)
				}

				fmt.Printf("state updated: %v", mem)
				if t == nil {
					fmt.Println("setting timer")
					t = time.NewTimer(opts.RestarGracePeriod)
				} else {
					fmt.Println("resetting timer")
					// See https://golang.org/pkg/time/#Timer.Reset
					if !t.Stop() {
						//<-t.C
						restartGrace = nil
					}
					t.Reset(opts.RestarGracePeriod)
				}
				restartGrace = t.C
			case <-w.Closed:
				return
			case <-restartGrace:
				fmt.Println("restart grace period elapsed. restarting...")
				if err := sp.DoSet(mem); err != nil {
					panic(err)
				}
				if opts.Restart {
					if err := proc.restart(); err != nil {
						panic(err)
					}
				}

				// reset timer
				t.Stop()
				t = nil
			case err := <-w.Error:
				if err == watcher.ErrWatchedFileDeleted {
					fmt.Println(err)
					continue
				}
				log.Fatalln(err)
			}
		}
	}()

	if len(opts.Watch) == 0 {
		return errors.New("no files to watch specified. specify one or more --watch flags")
	}

	for _, f := range opts.Watch {
		if err := w.AddRecursive(f); err != nil {
			return err
		}
	}

	for path, f := range w.WatchedFiles() {
		fmt.Printf("%s: %s\n", path, f.Name())
	}
	fmt.Println()

	fmt.Printf("Watching %d files\n", len(w.WatchedFiles()))

	if err := w.Start(opts.WatchInterval); err != nil {
		log.Fatalln(err)
	}

	{
		cleanup := make(chan struct{})

		c := make(chan os.Signal, 2)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		go func() {
			for range c {
				// Stop the targeted application process
				proc.stop()

				fmt.Println("application stopped")

				// Stop the file watcher
				w.Close()
				<-done

				close(cleanup)

				fmt.Println("watcher stopped")

				break
			}
		}()

		<-cleanup
	}

	return nil
}
