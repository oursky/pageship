package watch

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
)

type File[T any] struct {
	mu     sync.RWMutex
	logger *zap.Logger
	path   string
	load   func(path string) (T, error)

	watcher *fsnotify.Watcher
	stop    func()
	loaded  chan struct{}
	dirty   chan struct{}
	value   T
}

func NewFile[T any](logger *zap.Logger, path string, load func(path string) (T, error)) (*File[T], error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	file := &File[T]{
		logger:  logger,
		path:    path,
		load:    load,
		watcher: watcher,
		stop:    cancel,
		loaded:  make(chan struct{}),
		dirty:   make(chan struct{}, 1),
	}

	go file.watch(ctx)
	go file.reload(ctx)

	return file, nil
}

func (f *File[T]) Close() {
	f.stop()
	f.watcher.Close()
}

func (f *File[T]) Get(ctx context.Context) (value T, err error) {
	select {
	case <-f.loaded:
		break
	case <-ctx.Done():
		err = ctx.Err()
		return
	}

	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.value, nil
}

func (f *File[T]) watch(ctx context.Context) {
	log := f.logger.Named("watch")
	defer close(f.dirty)

	dir := filepath.Dir(f.path)

	done := false
	for !done {
		err := f.watcher.Add(dir)
		for err != nil {
			log.Debug("failed to watch", zap.Error(err))

			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}

			err = f.watcher.Add(dir)
		}

		realpath, err := filepath.EvalSymlinks(f.path)
		if err != nil {
			realpath = "" // Let next fs event restart the watch.
		}

		log.Info("watching file", zap.String("dir", dir), zap.String("path", f.path))
		done = f.doWatch(log, realpath)
	}
}

func (f *File[T]) doWatch(log *zap.Logger, realpath string) bool {
	for {
		select {
		case event, ok := <-f.watcher.Events:
			if !ok {
				return true
			}

			if event.Name != f.path {
				nowRealpath, err := filepath.EvalSymlinks(f.path)
				if err != nil {
					continue
				} else if nowRealpath != realpath {
					log.Info("restarting watch...", zap.String("path", f.path))
					f.markDirty()
					return false
				}
				continue
			}

			if event.Op.Has(fsnotify.Remove) {
				log.Info("restarting watch...", zap.String("path", f.path))
				f.markDirty()
				return false
			} else if event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Write) {
				log.Info("file changed", zap.String("path", f.path))
				f.markDirty()
			}

		case err, ok := <-f.watcher.Errors:
			if !ok {
				return true
			}
			log.Error("watch failed", zap.Error(err))
		}
	}
}

func (f *File[T]) markDirty() {
	select {
	case f.dirty <- struct{}{}:
	default:
	}
}

func (f *File[T]) reload(ctx context.Context) {
	log := f.logger.Named("load")

	load := make(chan struct{}, 1)
	timer := time.AfterFunc(0, func() {
		select {
		case load <- struct{}{}:
		default:
		}
	})
	defer timer.Stop()

	loaded := false
	for {
		select {
		case <-load:
			f.doLoad(log, loaded)
			if !loaded {
				close(f.loaded)
				loaded = true
			}

		case _, ok := <-f.dirty:
			if !ok {
				return
			}
			timer.Reset(time.Second)
		}
	}
}

func (f *File[T]) doLoad(log *zap.Logger, loaded bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	log.Info("reloading...", zap.String("path", f.path))
	value, err := f.load(f.path)
	if err != nil {
		log.Error("failed to reload", zap.Error(err))
		return
	}
	f.value = value
}
