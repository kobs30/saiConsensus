package internal

import (
	"github.com/saiset-co/saiService"
)

type InternalService struct {
	Context *saiService.Context
}

var counter = 0
var Validators []string
var Distribution []map[string]int64

func (is InternalService) Init() {

}

func (is InternalService) Process() {

}
