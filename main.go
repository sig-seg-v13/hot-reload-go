package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

func main() {
	path := flag.String("path", "", "path to a directory or a file to watch")
	write := flag.Bool("write", false, "watch for write events")
	create := flag.Bool("create", false, "watch for create events")
	chmod := flag.Bool("chmod", false, "watch for mode change events")
	remove := flag.Bool("remove", false, "watch for removal events")
	rename := flag.Bool("rename", false, "watch for renaming events")
	flag.Parse()

	if path != nil && *path == "" {
		log.Fatal("path is required")
	}

	operations := map[fsnotify.Op]bool{
		fsnotify.Chmod:  *chmod,
		fsnotify.Create: *create,
		fsnotify.Remove: *remove,
		fsnotify.Write:  *write,
		fsnotify.Rename: *rename,
	}

	var oneTrue bool
	for _, op := range operations {
		if op {
			oneTrue = true
			break
		}
	}

	if !oneTrue {
		log.Fatal("the tool should be watching at least one event")
	}

	_, err := os.Stat(*path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatal("path does not exist")
		}
		log.Fatalf("error accessing path: %q\n", *path)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("error creating watcher")
	}
	defer watcher.Close()

	err = watcher.Add(*path)
	if err != nil {
		log.Fatalf("error watching path %q: %v\n", *path, err)
	}

	eventTime := time.Now()

	for {
		select {
		case event := <-watcher.Events:
			for op, ok := range operations {
				if ok && event.Has(op) && time.Since(eventTime).Seconds() > 1 {
					fmt.Println("event fired:", event.Name)
					eventTime = time.Now()
				}
			}
		case err = <-watcher.Errors:
			log.Fatal("err:", err)
		}
	}
}
