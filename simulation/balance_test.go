package simulation

import (
	"fmt"
	"testing"
)

func TestInvoke(t *testing.T) {
	// bt := "{\"Target\": \"8867-3-0000002\", \"Source\": \"\", \"Amount\": 200}"
	// args := []string{bt}
	// byteRes,_ := Invoke("deposit", args)

	args := []string{"8867-2-0000001"}
	byteRes, err := Invoke("query", args)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(string(byteRes))
}
