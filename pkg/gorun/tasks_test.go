package gorun

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/lstack-org/utils/pkg/rest"
	"k8s.io/klog/v2"
)

func TestBatchTasks_Await(t *testing.T) {
	_, err := Tasks(func(ctx BatchContext) {
		time.Sleep(time.Second)
		ctx.AddError(fmt.Errorf("test err"))
	}, func(ctx BatchContext) {
		time.Sleep(2 * time.Second)
	}).Await(context.Background())
	if err == nil {
		t.Fatal("test err == nil")
	}
	if err.Error() != "test err" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBatchTasks_AwaitWithTimeout(t *testing.T) {
	_, err := Tasks(func(ctx BatchContext) {
		time.Sleep(time.Second)
	}, func(ctx BatchContext) {
		time.Sleep(3 * time.Second)
	}).AwaitWithTimeout(2 * time.Second)
	if err == nil {
		t.Fatal("timeout == nil")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBatchTasks_GetMergedError(t *testing.T) {
	_, err := Tasks(func(ctx BatchContext) {
		ctx.AddError(fmt.Errorf("errhaha1"))
	}, func(ctx BatchContext) {
		time.Sleep(time.Second)
		ctx.AddError(fmt.Errorf("errhaha2"))
	}).Await(context.Background())
	if err == nil {
		t.Fatal("MergedError == nil")
	}
	if err.Error() != "[errhaha1, errhaha2]" {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestBatchTasks_AwaitWithCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	//2秒后退出
	_, _ = Tasks(func(ctx BatchContext) {
		time.Sleep(10 * time.Second)
	}, func(ctx BatchContext) {
		time.Sleep(11 * time.Second)
	}).Await(ctx)
}

func TestBatchTasks_Item(t *testing.T) {
	res, err := Tasks(func(ctx BatchContext) {
		ctx.AddItem(1)
	}, func(ctx BatchContext) {
		ctx.AddError(nil)
	}, func(ctx BatchContext) {
		time.Sleep(time.Second)
	}).AwaitWithTimeout(5 * time.Second)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if res.GetItem() != 1 {
		t.Fatal("item != 1")
	}
}

func TestRequest(t *testing.T) {
	go func() {
		setupHttp()
	}()
	client, _ := rest.NewRESTClientEasy("test", "http://127.0.0.1", nil)

	_, err := Tasks(func(ctx BatchContext) {
		_, err := client.Get().AbsPath("/hello").DoRaw(ctx)
		ctx.AddError(err)
	}, func(ctx BatchContext) {
		time.Sleep(2 * time.Second)
	}).AwaitWithTimeout(3 * time.Second)
	t.Log(err)
}

func hello(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	fmt.Println("server: hello handler started")
	defer fmt.Println("server: hello handler ended")

	select {
	case <-time.After(10 * time.Second):
		fmt.Fprintf(w, "hello\n")
	case <-ctx.Done():
		err := ctx.Err()
		klog.Errorf("server: %v", err)
		internalError := http.StatusInternalServerError
		http.Error(w, err.Error(), internalError)
	}
}

func setupHttp() {
	http.HandleFunc("/hello", hello)
	_ = http.ListenAndServe(":80", nil)
}

func TestPanic(t *testing.T) {
	defer func() {
		if panicI := recover(); panicI == nil {
			t.Fatalf("recover failed")
		} else {
			if "woxiaole" != panicI {
				t.Fatalf("unexpected panicI: %v", panicI)
			}
		}
	}()
	_, _ = Tasks(func(ctx BatchContext) {
		panic("woxiaole")
	}, func(ctx BatchContext) {
		time.Sleep(time.Second)
	}).Await(context.TODO())
}
