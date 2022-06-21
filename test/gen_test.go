package test

import (
	"github.com/ozone2021/ozone/ozone-lib/config/config_utils"
	"github.com/ozone2021/ozone/ozone-lib/config/config_variable"
	"log"
	"strings"
	"testing"
)

type String struct {
	Value string
}

func (s String) String() string {
	return s.Value
}

func TestGen(t *testing.T) {
	varsMap := make(map[string]interface{})
	genVarNeedRendered := config_variable.NewGenVariable[string]("its {{new}} !!!", 1)
	genVarNew := config_variable.NewGenVariable[string]("working", 1)
	varsMap["new"] = genVarNew

	//value, _ := varsMap["new"].(*config_variable.GenVariable[string])

	output, err := genVarNeedRendered.Render(varsMap)
	if err != nil {
		t.Fatal("Should not be error.")
	}

	if strings.Compare(output.GetValue(), "its working !!!") != 0 {
		t.Fatal("Should be equal")
	}

	if err != nil {
		log.Println(err)
	}

	log.Println(output)
}

//func TestCopyCopy(t *testing.T) {
//	genVarString := config_variable.NewGenVariable[string]("its !!!", 1)
//	genVarSlice := config_variable.NewGenVariable[[]string]([]string{"working", "ish"}, 1)
//	config_utils.CopyCopy()
//}

func TestGetOrdinal(t *testing.T) {
	varsMap := make(map[string]interface{})

	genVarString := config_variable.NewGenVariable[string]("its !!!", 1)
	genVarSlice := config_variable.NewGenVariable[[]string]([]string{"working", "ish"}, 2)

	varsMap["new"] = genVarString

	stringOrd, err := config_utils.GetOrdinal(genVarString)
	if err != nil {
		log.Println(err)
	}
	sliceOrd, err2 := config_utils.GetOrdinal(genVarSlice)
	if err2 != nil {
		log.Println(err2)
	}

	log.Println(stringOrd)
	log.Println(sliceOrd)
	log.Println(interOrd)

}
