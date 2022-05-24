package encipherment

import (
	"fmt"
	"testing"
)

func TestAes(t *testing.T) {
	orig := "EV8mXRBtnPLVxtIhoxEA3vbTGDqUeBDlUUvu4Kds"
	key := "launcherSoNBPlus"
	fmt.Println("原文：", orig)

	encryptCode , _ := AesEncrypt(orig, key)
	fmt.Println("密文：", encryptCode)

	decryptCode := AesDecrypt(encryptCode, key)
	fmt.Println("解密结果：", decryptCode)
}
