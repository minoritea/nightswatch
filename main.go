package main

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v2"
	"log"
	"os"
	"path/filepath"
	"time"
)

var helpText = `
nightswatch - A configurable file watcher with fsnotify API

Usage:
   nightswatch [options]

   --config, -c   read a specific config file (default: ./nightswatch.toml)
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
`

func run(c *cli.Context) error {
	if c.Bool("help") {
		fmt.Println(helpText)
		return nil
	}
	conf, err := config(c)
	if err != nil {
		return err
	}

	watcher, err := NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	return watcher.Run(conf)
}

type Watcher struct {
	*fsnotify.Watcher
	interval time.Duration
}

func NewWatcher() (*Watcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &Watcher{w, 3 * time.Second}, nil
}

func (w *Watcher) Watch() error {
	tick := time.Tick(w.interval)
	for {
		log.Println("loop")
		changed := false
		select {
		case e := <-w.Events:
			log.Println(e)
			changed = true
		case err := <-w.Errors:
			return err
		case <-tick:
			if changed {
				w.Build()
				w.Reload()
				changed = false
			}
		}
	}
	return nil
}

func (w *Watcher) Build() error {
	log.Println("Build")
	return nil
}

func (w *Watcher) Reload() error {
	log.Println("Reload")
	return nil
}

func (w *Watcher) Load(c *Config) error {
	if c.Path == "" {
		errors.New("Paths must not be empty.")
	}
	return filepath.Walk(c.Path, w.Visit)
}

func (w *Watcher) Run(c *Config) error {
	if err := w.Load(c); err != nil {
		return err
	}
	return w.Watch()
}

func (w *Watcher) Visit(path string, _ os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	return w.Add(path)
}

type Config struct {
	Path     string `toml:"path"`
	Match    string `toml:"match"`
	Build    string `toml:"build"`
	Reload   string `toml:"reload"`
	Interval string `toml:"interval"`
}

func config(c *cli.Context) (*Config, error) {
	path := "./nightswatch.toml"
	if fv := c.String("config"); fv != "" {
		path = fv
	}
	tree, err := toml.LoadFile(path)
	if err != nil {
		return nil, err
	}
	conf := new(Config)
	if err := tree.Unmarshal(conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func main() {
	app := new(cli.App)
	app.Name = "nightswatch"
	app.HideHelp = true
	app.Flags = append(app.Flags, cli.HelpFlag)
	app.Action = run
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
