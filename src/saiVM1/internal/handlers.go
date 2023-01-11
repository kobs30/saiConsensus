package internal

import (
	"fmt"
	"github.com/robertkrimen/otto"
	// https://go.libhunt.com/otto-alternatives    other languages alternatives
	"github.com/saiset-co/saiService"
	"go.mongodb.org/mongo-driver/bson"
)

func (is InternalService) NewHandler() saiService.Handler {
	return saiService.Handler{
		"execute": saiService.HandlerElement{
			Name:        "execute",
			Description: "Execute smart-contract",
			Function: func(data interface{}) (interface{}, error) {
				return is.execute(data), nil
			},
		},
	}
}

func (is InternalService) execute(data interface{}) interface{} {
	counter++
	var Validators []string
	var Distribution []map[string]int64
	script := data.(string)
	fmt.Println(script)
	vm := otto.New()
	vm.Set("getBalanceChanges", func(call otto.FunctionCall) otto.Value {
		Wallet, _ := call.Argument(0).ToString()
		// {"vm_response.D":{$elemMatch: {"1FTGGrgfHTsgHsw0f8Hff8": {$exists:true} } } }
		res, _ := vm.ToValue(Wallet)
		return res
	})

	vm.Set("addValidator", func(call otto.FunctionCall) otto.Value {
		fmt.Println("Add validator")
		validatorWallet, _ := call.Argument(0).ToString()
		//fmt.Println("Hello, addBalance world!", fnWrapper)
		if getValidatorWalletLicence(validatorWallet) {
			Validators = append(Validators, validatorWallet)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
		return otto.Value{}
	})

	vm.Set("addBalance", func(call otto.FunctionCall) otto.Value {
		fmt.Println("Hello, addBalance world!")
		thewallet, _ := call.Argument(0).ToString()
		thebalancetoadd, _ := call.Argument(1).ToInteger()
		//fmt.Println("Hello, addBalance world!", fnWrapper)
		if thebalancetoadd > -100 {
			balance := make(map[string]int64)
			balance[thewallet] = int64(thebalancetoadd)
			Distribution = append(Distribution, balance)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
		return otto.Value{}
	})
	//vm.SetTimeout(time.Second)
	result, err := vm.Run(script)
	if err != nil {
		fmt.Println("error", err)
	}
	fmt.Println(result)
	return bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"callNumber": counter, "D": Distribution, "V": Validators, "C": result}}
}

func getValidatorWalletLicence(wallet string) bool {
	return true
}

func addBalance(thebalance, balance int64) otto.Value {
	thebalance += balance
	fmt.Println("adding balance")
	if thebalance > 10 {
		return otto.TrueValue()
	} else {
		return otto.FalseValue()
	}
}
