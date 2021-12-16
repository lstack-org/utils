package gorun

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestUntil_Cancel(t *testing.T) {
	//每秒加1，5秒后超时
	i := 0
	_, err := UntilWithTimeout(func(until Until) {
		i = i + 1
		if i > 3 {
			//条件符合，成功退出
			until.Cancel()
		}
	}, time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestUntil_Timout(t *testing.T) {
	//每秒加1，5秒后超时
	i := 0
	_, err := UntilWithTimeout(func(until Until) {
		i = i + 1
		if i > 7 {
			until.Cancel()
		}
	}, time.Second, 5*time.Second)
	if err == nil {
		t.Fatal("timeout err == nil")
	}
	if err.Error() != "timeout" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestUntil_ErrorBreak(t *testing.T) {
	//每秒加1，5秒后超时
	i := 0
	_, err := UntilWithTimeout(func(until Until) {
		i = i + 1
		switch i {
		case 3:
			until.ErrorBreak(errors.New("break error"))
		}
	}, time.Second, 5*time.Second)
	if err == nil {
		t.Fatal("break err == nil")
	}
	if err.Error() != "break error" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestUntil_ItemSave(t *testing.T) {
	item, err := UntilWithTimeout(func(until Until) {
		until.ItemSave(1)
		until.Cancel()
	}, time.Second, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if item == nil {
		t.Fatal("item == nil")
	}
	if item != 1 {
		t.Fatal("item != 1")
	}
}

func TestUntil_ErrorSave(t *testing.T) {
	_, err := UntilWithTimeout(func(until Until) {
		//只有超时时，保存的错误才能生效，否则会被忽略
		until.ErrorSave(errors.New("saved error"))
	}, time.Second, 5*time.Second)
	if err == nil {
		t.Fatal("timout err == nil")
	}
	if err.Error() != "timeout: saved error" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestUntil_Retry(t *testing.T) {
	i := 0
	_, err := UntilWithTimeout(func(until Until) {
		i = i + 1
		until.Retry(fmt.Errorf("err retry: %v", i), 3)
	}, time.Second, 5*time.Second)
	if err == nil {
		t.Fatal("err == nil")
	}
	if err.Error() != "err retry: 3" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestUntilWithCancel(t *testing.T) {
	i := 0
	_, err := UntilWithCancel(func(until Until) {
		i = i + 1
		if i > 6 {
			until.Cancel()
		}
	}, time.Second)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}
