package k8s

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
)

type Err struct {
	*errors.StatusError
}

func (e *Err) Error() string {
	return fmt.Sprintf("code: %v,reason: %s,message: %s",
		e.ErrStatus.Code, e.ErrStatus.Reason, e.ErrStatus.Message)
}
