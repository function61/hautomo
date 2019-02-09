package main

import (
	"fmt"
	"time"
)

type booleanStorage struct {
	values           map[string]bool
	changeTimestamps map[string]time.Time
}

func NewBooleanStorage(keys ...string) *booleanStorage {
	values := map[string]bool{}
	changeTimestamps := map[string]time.Time{}

	for _, key := range keys {
		values[key] = false
		changeTimestamps[key] = time.Time{} // zero
	}

	return &booleanStorage{values, changeTimestamps}
}

func (b *booleanStorage) GetLastChangeTime(key string) (time.Time, error) {
	changeTimestamp, exists := b.changeTimestamps[key]
	if !exists {
		return changeTimestamp, fmt.Errorf("boolean %s does not exist", key)
	}

	return changeTimestamp, nil
}

func (b *booleanStorage) Get(key string) (bool, error) {
	value, exists := b.values[key]
	if !exists {
		return false, fmt.Errorf("boolean %s does not exist", key)
	}

	return value, nil
}

func (b *booleanStorage) Set(key string, to bool) (bool, error) {
	previousValue, exists := b.values[key]
	if !exists {
		return false, fmt.Errorf("boolean %s does not exist", key)
	}

	if previousValue == to {
		return false, nil // no change in value
	}

	b.values[key] = to
	b.changeTimestamps[key] = time.Now()

	return true, nil // value changed
}
