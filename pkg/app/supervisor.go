package app

import (
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
	"syscall"
	"path/filepath"
	"github.com/ghodss/yaml"
	"text/template"
	"bytes"
)

type SupervisorOpts struct {
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
	SupervisorOpts

	cmd string
	args []string

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

		// To avoid `panic: open /etc/falco/falco.yaml.orig: read-only file system` errors
		origFile := fmt.Sprintf("/var/falco-operator/%s.orig", filepath.Base(set.Path))
		bytesData, err = afero.ReadFile(o.fs, origFile)
		if err != nil {
			bytesData, err = afero.ReadFile(o.fs, set.Path)
			if err != nil {
				return err
			}
			if err := afero.WriteFile(o.fs, origFile, bytesData, 0644); err != nil {
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

		writePath := fmt.Sprintf("/var/falco-operator/%s", filepath.Base(set.Path))
		err = afero.WriteFile(o.fs, writePath, append(result, []byte("\n")...), 0644)
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
func NewSupervisor(args []string, opts SupervisorOpts) (*Supervisor, error) {
	if len(args) == 0 {
		return nil, errors.New("missing args")
	}

	if opts.RestarGracePeriod <= opts.WatchInterval {
		return nil, errors.New("watch interval should be less than restart grace period")
	}

	cmd := args[0]
	arg := args[1:]

	return &Supervisor{SupervisorOpts: opts, fs: afero.NewOsFs(), cmd: cmd, args: arg,}, nil
}

func Supervise(args []string, opts SupervisorOpts) error {
	sp, err := NewSupervisor(args, opts)
	if err != nil {
		return err
	}
	return sp.Supervise()
}

func (sp *Supervisor) Supervise() error {
	opts := sp.SupervisorOpts
	cmd := sp.cmd
	arg := sp.args

	proc := &process{cmd: cmd, arg: arg, stopGracePeriod: opts.StopGracePeriod,}

	falcoYaml, err := afero.ReadFile(sp.fs, "/etc/falco/falco.yaml")
	if err != nil {
		return err
	}
	if err := afero.WriteFile(sp.fs, "/var/falco-operator/falco.yaml", falcoYaml, 0644); err != nil {
		return err
	}

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
						if err := afero.Walk(sp.fs, event.Path, func(path string, info os.FileInfo, err error) error {
							if info.IsDir() {
								return nil
							}
							if strings.HasPrefix(filepath.Base(filepath.Dir(path)), ".") {
								fmt.Printf("ignoring dir of %s\n", path)
								return nil
							}
							if strings.HasPrefix(info.Name(), ".") {
								fmt.Printf("ignoring %s\n", path)
								return nil
							}
							fmt.Printf("%s has been updated\n", path)
							paths = append(paths, path)
							return nil
						}); err != nil {
							panic(err)
						}
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

	paths := []string{}
	for path, fileOrDir := range w.WatchedFiles() {
		if fileOrDir.IsDir() {
			if err := afero.Walk(sp.fs, path, func(path string, info os.FileInfo, err error) error {
				if info.IsDir() {
					return nil
				}
				// Ignore files under ..$DIR like:
				// - /var/falco-operator/rules/..2018_10_04_13_42_27.933690714/default
				// -/var/falco-operator/rules/..2018_10_04_13_42_27.933690714/test1
				if strings.HasPrefix(filepath.Base(filepath.Dir(path)), ".") {
					fmt.Printf("ignoring dir of %s\n", path)
					return nil
				}
				if strings.HasPrefix(info.Name(), ".") {
					fmt.Printf("ignoring %s\n", path)
					return nil
				}
				fmt.Printf("%s has been updated\n", path)
				paths = append(paths, path)
				return nil
			}); err != nil {
				return err
			}
		} else {
			paths = append(paths, path)
		}
	}
	for _, path := range paths {
		content, err := afero.ReadFile(sp.fs, path)
		if err != nil {
			panic(err)
		}
		mem[path] = string(content)
	}
	if err := sp.DoSet(mem); err != nil {
		panic(err)
	}

	err = proc.start()
	if err != nil {
		return err
	}

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
