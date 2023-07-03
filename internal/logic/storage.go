package logic

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/xypwn/southpark-downloader-ui/pkg/data"
)

type StorageItem[T any] struct {
	*data.Binding[T]
}

func NewStorageItem[T any](s *Storage, id string, getDefault func() T, onError func(error)) (*StorageItem[T], error) {
	s.mtx.RLock()
	if _, ok := s.items[id]; ok {
		panic("logic: NewStorageItem: attempted to create item with same ID twice (\"" + id + "\")")
	}
	s.mtx.RUnlock()

	filePath := path.Join(s.pathBase, id+".json")

	save := func(v T) {
		fmt.Println("STORAGE: Saving", id)
		s.mtx.Lock()
		defer s.mtx.Unlock()

		s.items[id] = struct{}{}

		data, err := json.Marshal(v)
		if err != nil {
			onError(err)
			return
		}
		if err := os.WriteFile(
			filePath,
			data,
			0644,
		); err != nil {
			onError(err)
			return
		}
	}

	binding := data.NewBinding[T]()
	client := binding.NewClient()
	client.AddListener(save)

	data, err := os.ReadFile(filePath)
	if err == nil {
		var v T
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		client.Change(func(T) T {
			return v
		})
	} else {
		if errors.Is(err, os.ErrNotExist) {
			defaultValue := getDefault()
			save(defaultValue)
			client.Change(func(T) T {
				return defaultValue
			})
		} else {
			return nil, err
		}
	}

	return &StorageItem[T]{
		Binding: binding,
	}, nil
}

type Storage struct {
	pathBase string
	items    map[string]struct{}
	mtx      sync.RWMutex
}

func NewStorage(pathBase string) (*Storage, error) {
	res := &Storage{
		pathBase: pathBase,
		items:    make(map[string]struct{}),
	}
	if err := os.MkdirAll(pathBase, os.ModePerm); err != nil {
		return nil, err
	}
	return res, nil
}
