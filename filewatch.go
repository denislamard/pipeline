package pipeline

import (
	"context"
	"fmt"
	"github.com/rjeczalik/notify"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	MaxFilesCapacity = 100
)

type FileWatch struct {
	channelInfoEvent chan notify.EventInfo
}

func NewFileWatch() *FileWatch {
	return new(FileWatch).init()
}

func (fw *FileWatch) init() *FileWatch {
	fw.channelInfoEvent = make(chan notify.EventInfo)
	return fw
}

func (fw *FileWatch) Watching(ctx context.Context) <-chan string {

	fileCreated := make(chan string, MaxFilesCapacity)

	go func() {
		defer func() {
			log.Info("stopping watching paths")
			close(fileCreated)
		}()
		for {
			select {
			case event := <-fw.channelInfoEvent:
				fileCreated <- event.Path()
			case <-ctx.Done():
				return
			}
		}
	}()
	return fileCreated
}

func (fw *FileWatch) Stop() {
	log.Info("stopping file watch component")
	notify.Stop(fw.channelInfoEvent)
	close(fw.channelInfoEvent)
}

func (fw *FileWatch) AddPath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return fmt.Errorf("%s does not exist", path)
	}
	if err := notify.Watch(path, fw.channelInfoEvent, notify.Create); err != nil {
		return err
	}
	return nil
}
