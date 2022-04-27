package rest

import (
	"encoding/json"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Err struct {
	*errors.StatusError
}

func (e *Err) Error() string {
	return fmt.Sprintf("code: %v,reason: %s,message: %s",
		e.ErrStatus.Code, e.ErrStatus.Reason, e.ErrStatus.Message)
}

func ErrorConvert(result []byte, err error) error {
	if err == nil {
		return nil
	}

	if statusErr, ok := err.(*errors.StatusError); !ok {
		return err
	} else {
		status := metav1.Status{}
		_ = json.Unmarshal(result, &status)
		status.Reason = statusErr.Status().Reason
		return &Err{StatusError: &errors.StatusError{ErrStatus: status}}
	}
}
