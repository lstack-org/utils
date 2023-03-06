package gt

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"unsafe"
)

// ErrUnknownType 未知错误类型
var ErrUnknownType = errors.New("unknown type")

// Decoder is the decoding interface
type Decoder interface {
	Decode(v interface{}) error
	Value() interface{}
}

// BodyDecode body decoder structure
type BodyDecode struct {
	r   io.Reader
	obj interface{}
}

// NewBodyDecode create a new body decoder
func NewBodyDecode(r io.Reader) Decoder {
	if r == nil {
		return nil
	}

	return &BodyDecode{r: r}
}

// Decode body decoder
func (b *BodyDecode) Decode(v interface{}) error {
	return Body(b.r, v)
}

// Value ...
func (b *BodyDecode) Value() interface{} {
	return b.obj
}

// Body body decoder
func Body(r io.Reader, obj interface{}) error {
	if w, ok := obj.(io.Writer); ok {
		_, err := io.Copy(w, r)
		return err
	}

	all, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	value := LoopElem(reflect.ValueOf(obj))

	if value.Kind() == reflect.String {
		value.SetString(BytesToString(all))
		return nil
	}

	if _, ok := value.Interface().([]byte); ok {
		value.SetBytes(all)
		return nil
	}
	return fmt.Errorf("type (%T) %s", value, ErrUnknownType)
}

// LoopElem 不停地对指针解引用
func LoopElem(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return v
		}
		v = v.Elem()
	}
	return v
}

// BytesToString 没有内存开销的转换
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
