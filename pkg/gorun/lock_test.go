package gorun

import (
	"testing"
	"time"
)

func TestTryLock(t *testing.T) {
	mutex := NewMutex()

	mutex.TryLock()

	if mutex.TryLock() {
		t.Fatal()
	}

	mutex.Unlock()

	if !mutex.TryLock() {
		t.Fatal()
	}
}

func TestMutex_TryLockWithTimeout(t *testing.T) {
	mutex := NewMutex()

	mutex.Lock()

	//参数获取锁，等待1秒
	if mutex.TryLockWithTimeout(time.Second) {
		t.Fatal()
	}

	mutex.Unlock()

	if !mutex.TryLock() {
		t.Fatal()
	}
}
