package gogen

import (
	"context"
	"fmt"
	"os"
)

type Inport[REQUEST, RESPONSE any] interface {
	Execute(ctx context.Context, req REQUEST) (*RESPONSE, error)
}

func GetInport[Req, Res any](usecase any) Inport[Req, Res] {
	inport, ok := usecase.(Inport[Req, Res])
	if !ok {
		fmt.Printf("some usecase is not registered yet in application\n")
		os.Exit(0)
	}
	return inport
}
