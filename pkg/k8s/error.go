package k8s

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Err struct {
	metav1.Status
}

func (e *Err) Error() string {
	return fmt.Sprintf("code: %v,reason: %s,message: %s",
		e.Code, e.Reason, e.Message)
}
