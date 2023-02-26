package internal

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/robertkrimen/otto"
	"math"
	"math/rand"
	"regexp"
	"strings"
	// https://go.libhunt.com/otto-alternatives    other languages alternatives
	"github.com/saiset-co/saiService"
	"go.mongodb.org/mongo-driver/bson"
)

func (is InternalService) Handlers() saiService.Handler {
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

type TX struct {
	Message         string `json:"message"`
	MessageHash     string `json:"message_hash"`
	Nonce           int    `json:"nonce"`
	SenderAddress   string `json:"sender_address"`
	SenderSignature string `json:"sender_signature"`
	Type            string `json:"type"`
}

type VMrequest struct {
	Block  int    `json:"block"`
	Rnd    int64  `json:"rnd"`
	Tx     TX     `json:"tx"`
	Script string `json:"message"`
}

type VMscript struct {
	Script string `json:"data"`
	Method string `json:"method"`
}

type RegType struct {
	Type        string `json:"type"`
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func updateInProcessingBalance(block int, Distribution []map[string]int64) bool {
	return true
}

func getInProcessingBalance(block int, wallet string) int64 {
	return 0
}

func (is InternalService) execute(data interface{}) interface{} {
	counter++
	var Validators []map[string]bool
	var Distribution []map[string]int64
	var CustomFld []map[string]interface{} // []map[string]interface{}
	var Fee float64
	var request VMrequest
	fmt.Println("REQUEST:::::::", data)
	var CustomTokens []map[string][]map[string]int64 //[ { toeknGroove: [{address:value},{address2:value2]...] },... ]
	var Register []map[string]RegType                // [ {contractaddress:{type:token,symbol:GRV,name:Groove,descr:"yet anotehr token"} }, {'amm_'tokencontractaddress:{type:amm,symbol:ammGRV,name:ammGroove, decr:"can contains some amm rules ot whatever"} ,...]
	fmt.Println("CustomTokens", CustomTokens)
	fmt.Println("Register", Register)
	var theScriptHash string

	dataJSON, _ := json.Marshal(data)
	fmt.Println("REQUEST JSON CONV:::::::", dataJSON)
	err := json.Unmarshal(dataJSON, &request)
	if err != nil {
		fmt.Println("datERROR", err)
		fmt.Println("REQUEST CONV::ERROR", err)
	}
	fmt.Println("REQUEST IS:::::::", request)
	fmt.Println("TX_STR:::::", request.Tx)
	script := request.Script
	decodedScript, err := base64.StdEncoding.DecodeString(request.Script)
	if err == nil {
		script = string(decodedScript)
	}

	var vmScript VMscript
	err = json.Unmarshal([]byte(script), &vmScript)
	if err != nil {
		fmt.Println("datERROR", err)
		fmt.Println("REQUEST CONV::ERROR", err)
	}
	currentContract := ""
	switch vmScript.Method {
	case "execute":
		{
		}
	case "save":
		// Curl example:
		// curl --location --request GET 'http://185.229.119.188:8018' \
		//--header 'Content-Type: text/plain' \
		//--data-raw '{"method":"get-tx","data":"{\"method\": \"execute\", \"data\": \"function hello(name) { return '\'' Hello '\''+ name}; hello('\''world'\'')  ;\"}"}'
		saveScriptData := make(map[string]interface{})
		saveScriptData[request.Tx.MessageHash] = "saved"
		CustomFld = append(CustomFld, saveScriptData)
		return bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"D": Distribution, "V": Validators, "C": CustomFld, "F": Fee}}
	case "run":
		// Curl example:
		// curl --location --request GET 'http://185.229.119.188:8018' \
		//--header 'Content-Type: text/plain' \
		//--data-raw '{"method":"get-tx","data":"{\"method\": \"run\", \"data\": \"f71aec98f3fab3f8ab3af2131465173194fbe45fcb82e07f49b98ef18a10e03b.hello('\''world'\'')  ;\"}"}'
		parts := strings.Split(vmScript.Script, ".")
		Contract := parts[0]
		Function := parts[1]
		currentContract = Contract
		runScript, _ := is.getMessageByHash(Contract)
		logRecord := make(map[string]interface{})
		logRecord[Contract] = "contract was run with " + Function
		CustomFld = append(CustomFld, logRecord)
		// Security check!!!
		//!!! check that Function contains only not private functions declared in the contract and nohing more!!!
		//get declared functions from contract
		functionsList := fetchJSfunctions(runScript)
		//get(form) functions call from contract.cript
		code := Function // "mid(12,'asdweeferferrtgr'); v = 12*getRes(); load(mid(12.getName());"
		myFuncs := functionsList

		var Execute string
		// if _init() exists Execute += "_init();"
		for _, f := range myFuncs {
			re := regexp.MustCompile(f + `\(.*?\)`)
			matches := re.FindAllString(code, -1)
			for _, match := range matches {
				Execute += match + ";"
				fmt.Println(match)
			}
		}
		theScriptHash = currentContract
		vmScript.Script = runScript + " " + Execute
		// ??? if not return vm_result false ???
		//vmScript.Script = runScript + " " + Function
	}

	theSender := request.Tx.SenderAddress // request.Tx.Sender
	if len(theScriptHash) == 0 {
		theScriptHash = request.Tx.MessageHash
	}
	fmt.Println("XXXXX", vmScript.Script)
	if vmScript.Script == "fly me to the moon" && request.Block == 1 {
		thevalidator := make(map[string]bool)
		thevalidator[theSender] = true
		Validators = append(Validators, thevalidator)

		initbalance := make(map[string]int64)
		initbalance[theSender] = int64(1000)
		Distribution = append(Distribution, initbalance)
		updateInProcessingBalance(request.Block, Distribution)
		initbalance = make(map[string]int64)
		initbalance["139uwuYCM1knfLdyVX2yjzwhDDz73Zx7Sj"] = int64(1000)
		Distribution = append(Distribution, initbalance)
		updateInProcessingBalance(request.Block, Distribution)
		initbalance = make(map[string]int64)
		initbalance["1PKVs1mizz4abZ8zvk4gbUNdSvmMXTFfEh"] = int64(1000)
		Distribution = append(Distribution, initbalance)
		updateInProcessingBalance(request.Block, Distribution)

		initsettings := make(map[string]interface{})
		initsettings["FeePerMessageSymbol"] = "0.01"
		CustomFld = append(CustomFld, initsettings)
		initsettings = make(map[string]interface{})
		initsettings["Fee_getBalance"] = "0.05"
		CustomFld = append(CustomFld, initsettings)
		initsettings = make(map[string]interface{})
		initsettings["FeeSaveDataPerSymbol"] = "0.01"
		CustomFld = append(CustomFld, initsettings)

		fmt.Println("RETURN::::::::::::", bson.M{"GENESYS": "GENESYS", "vm_processed": true, "vm_result": true, "vm_response": bson.M{"D": Distribution, "V": Validators, "C": CustomFld}})
		return bson.M{"GENESYS": "GENESYS", "vm_processed": true, "vm_result": true, "vm_response": bson.M{"D": Distribution, "V": Validators, "C": CustomFld}}
	}

	vm := otto.New()
	vm.Set("getRnd", func(call otto.FunctionCall) otto.Value {
		res, _ := vm.ToValue(request.Rnd)
		return res
	})

	vm.Set("getSender", func(call otto.FunctionCall) otto.Value {
		res, _ := vm.ToValue(request.Tx.SenderAddress)
		return res
	})

	vm.Set("getBlock", func(call otto.FunctionCall) otto.Value {
		res, _ := vm.ToValue(request.Block)
		return res
	})

	vm.Set("getMessageHash", func(call otto.FunctionCall) otto.Value {
		res, _ := vm.ToValue(request.Tx.MessageHash)
		return res
	})

	vm.Set("getScriptHash", func(call otto.FunctionCall) otto.Value {
		res, _ := vm.ToValue(theScriptHash)
		return res
	})

	vm.Set("callMessagesPool", func(call otto.FunctionCall) otto.Value {
		Query, _ := call.Argument(0).ToString()
		Options, _ := call.Argument(1).ToString()
		var theQuery interface{}
		var theOptions interface{}
		_ = json.Unmarshal([]byte(Query), &theQuery)
		_ = json.Unmarshal([]byte(Options), &theOptions)
		_, blockchainData := is.Storage.Get("MessagesPool", theQuery, theOptions)
		fmt.Println("blockchainData", string(blockchainData))
		res, _ := vm.ToValue(string(blockchainData))
		return res
	})

	vm.Set("getBalance", func(call otto.FunctionCall) otto.Value {
		Wallet, _ := call.Argument(0).ToString()
		WalletBalance, err := is.getBalance(Wallet)
		if err != nil {
			res, _ := vm.ToValue(false)
			return res
		}
		res, _ := vm.ToValue(WalletBalance)
		return res
	})

	vm.Set("getTokenBalance", func(call otto.FunctionCall) otto.Value {
		token, _ := call.Argument(0).ToString()
		Wallet, _ := call.Argument(1).ToString()
		WalletBalance, err := is.getTokenBalance(token, Wallet)
		if err != nil {
			res, _ := vm.ToValue(false)
			return res
		}
		res, _ := vm.ToValue(WalletBalance)
		return res
	})

	vm.Set("addValidator", func(call otto.FunctionCall) otto.Value {
		fmt.Println("Add validator")
		validatorWallet, _ := call.Argument(0).ToString()
		if is.setValidator(validatorWallet, &Validators) {
			return otto.TrueValue()
		}
		return otto.FalseValue()
	})

	vm.Set("getValidators", func(call otto.FunctionCall) otto.Value {
		// {"collection":"MessagesPool","options":{},"select":{"vm_response.V": {"$ne" : null}  }}
		validatorsList, err := is.getValidators()
		if err != nil {
			return otto.FalseValue()
		}
		res, _ := vm.ToValue(validatorsList)
		return res
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
			updateInProcessingBalance(request.Block, Distribution)
			return otto.TrueValue()
		}
		return otto.FalseValue()
	})

	vm.Set("transfer", func(call otto.FunctionCall) otto.Value {
		to, _ := call.Argument(0).ToString()
		amount, _ := call.Argument(1).ToInteger()
		WalletBalance, _ := is.getBalance(request.Tx.SenderAddress)
		if (WalletBalance - amount) > 0 {
			balance := make(map[string]int64)
			balance[request.Tx.SenderAddress] = int64(0 - amount)
			Distribution = append(Distribution, balance)
			updateInProcessingBalance(request.Block, Distribution)
			balance = make(map[string]int64)
			balance[to] = int64(amount)
			Distribution = append(Distribution, balance)
			updateInProcessingBalance(request.Block, Distribution)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})
	vm.Set("transferToTheContract", func(call otto.FunctionCall) otto.Value {
		if currentContract == "" {
			return otto.FalseValue()
		}
		to := currentContract
		amount, _ := call.Argument(0).ToInteger()
		fmt.Println("transferToTheContract", request.Tx.SenderAddress, ">>>", to, ">>>", amount)
		WalletBalance, _ := is.getBalance(request.Tx.SenderAddress)
		if (WalletBalance - amount) > 0 {
			balance := make(map[string]int64)
			balance[request.Tx.SenderAddress] = int64(0 - amount)
			Distribution = append(Distribution, balance)
			updateInProcessingBalance(request.Block, Distribution)
			balance = make(map[string]int64)
			balance[to] = int64(amount)
			Distribution = append(Distribution, balance)
			updateInProcessingBalance(request.Block, Distribution)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})
	vm.Set("transferFromTheContract", func(call otto.FunctionCall) otto.Value {
		if currentContract == "" {
			return otto.FalseValue()
		}
		to, _ := call.Argument(0).ToString()
		amount, _ := call.Argument(1).ToInteger()
		fmt.Println("transferFromTheContract", currentContract, ">>>", to, ">>>", amount)
		contractBalance, _ := is.getBalance(currentContract)
		if (contractBalance - amount) > 0 {
			balance := make(map[string]int64)
			balance[currentContract] = int64(0 - amount)
			Distribution = append(Distribution, balance)
			updateInProcessingBalance(request.Block, Distribution)
			balance = make(map[string]int64)
			balance[to] = int64(amount)
			Distribution = append(Distribution, balance)
			updateInProcessingBalance(request.Block, Distribution)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})

	vm.Set("currentFee", func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	})

	vm.Set("Register", func(call otto.FunctionCall) otto.Value {
		address, _ := call.Argument(0).ToString()
		item, _ := call.Argument(1).ToString()
		fmt.Println("Register Item:", item)
		var itemObject RegType
		err := json.Unmarshal([]byte(item), &itemObject)
		if err != nil {
			fmt.Println("Reigter unmarshal error::", err)
			return otto.FalseValue()
		}
		if !is.Register(address, itemObject, &Register) {
			return otto.FalseValue()
		}
		return otto.TrueValue()
	})

	vm.Set("getRndSet", func(call otto.FunctionCall) otto.Value {
		distType, _ := call.Argument(0).ToString()
		numbersInSet, _ := call.Argument(1).ToInteger()
		min, _ := call.Argument(2).ToFloat()
		max, _ := call.Argument(3).ToFloat()
		param, _ := call.Argument(4).ToFloat()
		fmt.Println("PARAMS", distType, int(numbersInSet), min, max, param, request.Rnd)
		set := generateRandomNumbers(distType, int(numbersInSet), min, max, param, request.Rnd)
		res, _ := vm.ToValue(set)
		return res
	})

	vm.Set("addCustomFld", func(call otto.FunctionCall) otto.Value {
		field, _ := call.Argument(0).ToString()
		value, _ := call.Argument(1).ToString()
		customData := make(map[string]string)
		customData[field] = value
		record, err := json.Marshal(customData)
		if err != nil {
			return otto.FalseValue()
		}
		fmt.Println(string(record))
		CustomFldElement := make(map[string]interface{})
		CustomFldElement[theScriptHash] = customData //string(record)
		CustomFld = append(CustomFld, CustomFldElement)
		return otto.TrueValue()
	})

	vm.Set("addTokenBalance", func(call otto.FunctionCall) otto.Value {
		token := currentContract
		to, _ := call.Argument(0).ToString()
		amount, _ := call.Argument(1).ToInteger()
		fmt.Println("Add token balance Set", token, "..", to, "..", amount)
		fmt.Println("Add token balance CustomTokens", CustomTokens)
		theamount, err := is.addTokenBalance(token, to, amount, &CustomTokens)
		fmt.Println("Add token balance CustomTokens append", CustomTokens)
		if err != nil {
			fmt.Println("Add token balance Set error", err)
			return otto.FalseValue()
		}
		res, _ := vm.ToValue(theamount)
		return res
	})

	vm.Set("transferToken", func(call otto.FunctionCall) otto.Value {
		token := currentContract
		if !is.isRegistered(currentContract) {
			return otto.FalseValue()
		}
		to, _ := call.Argument(0).ToString()
		amount, _ := call.Argument(1).ToInteger()
		WalletBalance, _ := is.getTokenBalance(token, request.Tx.SenderAddress)
		if (WalletBalance - amount) > 0 {
			fromAmount, err := is.addTokenBalance(token, request.Tx.SenderAddress, int64(0-amount), &CustomTokens)
			if err != nil {
				return otto.FalseValue()
			}
			toAmount, err := is.addTokenBalance(token, to, amount, &CustomTokens)
			if err != nil {
				return otto.FalseValue()
			}
			fmt.Println(fromAmount, ">>>>", toAmount)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})

	vm.Set("transferTokenFromTheContract", func(call otto.FunctionCall) otto.Value {
		token, _ := call.Argument(0).ToString()
		to, _ := call.Argument(1).ToString()
		amount, _ := call.Argument(2).ToInteger()
		if token == "" {
			return otto.FalseValue()
		}
		fmt.Println("transferTokenFromTheContract", currentContract, ">>>", to, ">>>", amount)
		contractTokenBalance, _ := is.getTokenBalance(token, currentContract)
		if (contractTokenBalance - amount) > 0 {
			fromAmount, err := is.addTokenBalance(token, currentContract, int64(0-amount), &CustomTokens)
			if err != nil {
				return otto.FalseValue()
			}
			toAmount, err := is.addTokenBalance(token, to, amount, &CustomTokens)
			if err != nil {
				return otto.FalseValue()
			}
			fmt.Println(fromAmount, ">>>>", toAmount)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})

	//vm.SetTimeout
	/*
		vm.Interrupt = make(chan func(), 1)
		go func() {
			time.Sleep(1000 * time.Millisecond)
			vm.Interrupt <- func() {
				fmt.Println("Script execution timed out.")
			}
		}()
	*/
	// remove the following ======
	vmScript.Script = strings.Replace(vmScript.Script, "semicolon", ";", -1)
	vmScript.Script = strings.Replace(vmScript.Script, "plus", "+", -1)
	vmScript.Script = strings.Replace(vmScript.Script, "~percent~", "%", -1)
	//============================
	result, err := vm.Run(vmScript.Script)
	if err != nil {
		fmt.Println("error", err)
		return bson.M{"vm_processed": true, "vm_result": false}
	}

	fmt.Println(result)
	CustomFldElement := make(map[string]interface{})
	CustomFldElement[theScriptHash], _ = result.ToString()
	CustomFld = append(CustomFld, CustomFldElement)
	fmt.Println("callNumber:", counter)
	fmt.Println("RETURN ::::: ", bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"D": Distribution, "V": Validators, "C": CustomFld, "F": Fee, "T": CustomTokens, "R": Register}})
	return bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"D": Distribution, "V": Validators, "C": CustomFld, "F": Fee, "T": CustomTokens, "R": Register}}
}

func (is InternalService) getBalance(Wallet string) (int64, error) {
	_, blockhainData := is.Storage.Get("MessagesPool", bson.M{"vm_response.D": bson.M{"$elemMatch": bson.M{Wallet: bson.M{"$exists": 1}}}}, bson.M{})
	fmt.Println("blockhainData", string(blockhainData))
	var jsonBlockchainData JSONRESP
	err := json.Unmarshal(blockhainData, &jsonBlockchainData)
	if err != nil {
		fmt.Println("datERROR", err)
		return 0, err
	}

	// Unmarshal C through reflekt
	/*
		CType := reflect.ValueOf(jsonBlockchainData.Result[0].VMResponse.C)
		if CType.Kind() == reflect.Map {
			var CustomFld map[string]string
			json.Unmarshal(jsonBlockchainData.Result[0].VMResponse.C, &CustomFld)
			fmt.Println(CustomFld)
		}
	*/

	fmt.Println("dat", jsonBlockchainData)

	if len(jsonBlockchainData.Result) > 0 {
		fmt.Println("datVMResponse PUPUPUP", jsonBlockchainData.Result[0].VMResponse.D[0][Wallet])
	}

	// {"vm_response.D":{$elemMatch: {"1FTGGrgfHTsgHsw0f8Hff8": {$exists:true} } } }
	var WalletBalance int64
	for _, el := range jsonBlockchainData.Result {
		for _, d := range el.VMResponse.D {
			balance, ok := d[Wallet]
			if ok {
				fmt.Println("-------", d[Wallet])
				WalletBalance += balance
			}
		}
	}
	return WalletBalance, nil
}

func (is InternalService) getTokenBalance(Token, Wallet string) (int64, error) {
	//return 0,nil
	//{"collection":"MessagesPool", "select": { "vm_response.T": { "$elemMatch": { "d1d79e9ed48a3905702143887ba62228eae892117231e1549a80e92f65267b24": { "$elemMatch": { "15UaBLZ7x6czXnFmHxzd3nFQNvXq7DJ3Gp": { "$exists": true } } } } } }, "options": {} }
	_, blockhainData := is.Storage.Get("MessagesPool", bson.M{"vm_response.T": bson.M{"$elemMatch": bson.M{Token: bson.M{"$elemMatch": bson.M{Wallet: bson.M{"$exists": true}}}}}}, bson.M{})
	//_, blockhainData := is.Storage.Get("MessagesPool", bson.M{"vm_response.T": bson.M{"$elemMatch": bson.M{Wallet: bson.M{"$exists": 1}}}}, bson.M{})
	fmt.Println("blockhainData request", bson.M{"vm_response.T": bson.M{"$elemMatch": bson.M{Token: bson.M{"$elemMatch": bson.M{Wallet: bson.M{"$exists": true}}}}}})
	fmt.Println("blockhainData", string(blockhainData))
	var jsonBlockchainData JSONRESP
	err := json.Unmarshal(blockhainData, &jsonBlockchainData)
	if err != nil {
		fmt.Println("getTokenBalance datERROR", err)
		return 0, err
	}
	if len(jsonBlockchainData.Result) > 0 {
		fmt.Println("datVMResponse getTokenBalance", jsonBlockchainData.Result[0].VMResponse.T)
	}
	//????? check if jsonBlockchainData.Result[0].VMResponse.T exists ?????
	var tokenDistr []map[string][]map[string]int64
	err = json.Unmarshal(jsonBlockchainData.Result[0].VMResponse.T, &tokenDistr)
	if err != nil {
		fmt.Println("getTokenBalance tokenDistr", err)
		return 0, err
	}
	var WalletBalance int64
	for _, el := range tokenDistr {
		theTokenBalance, ok := el[Token]
		if ok {
			for _, b := range theTokenBalance {
				balance, bok := b[Wallet]
				if bok {
					WalletBalance += balance
				}
			}
		}
	}
	return WalletBalance, nil
}

func (is InternalService) getMessageByHash(MessageHash string) (string, error) {
	// {"collection":"MessagesPool", "select": {"block_number": {"$exists":true}, "message_hash": "d8e7f63670d4e8ff434d031de226609bc1cb64eeae3ee496553f4cabb22a8c64","vm_processed": true,"vm_result": true}, "options": {} }
	_, blockhainData := is.Storage.Get("MessagesPool", bson.M{"block_number": bson.M{"$exists": true}, "message_hash": MessageHash, "vm_processed": true, "vm_result": true}, bson.M{})
	if len(blockhainData) > 0 {
		fmt.Println("blockhainData", string(blockhainData))
	} else {
		fmt.Println("blockhainData empty", bson.M{"block_number": bson.M{"$exists": true}, "message_hash": MessageHash, "vm_processed": true, "vm_result": true})
	}
	var jsonBlockchainData JSONRESP
	err := json.Unmarshal(blockhainData, &jsonBlockchainData)
	if err != nil {
		fmt.Println("datERROR", err)
		return "", err
	}
	message := jsonBlockchainData.Result[0].Message.Message
	decodedScript, err := base64.StdEncoding.DecodeString(message)
	if err == nil {
		message = string(decodedScript)
	}
	var vmScript VMscript
	err = json.Unmarshal([]byte(message), &vmScript)
	if err != nil {
		fmt.Println("UnmarshalError 439 Message ", message)
		fmt.Println("UnmarshalError 439", err)
		return "", err
	} else {
		return vmScript.Script, nil
	}
}

func getValidatorWalletLicence(wallet string) bool {
	return true
}

func (is InternalService) setValidator(wallet string, Validators *[]map[string]bool) bool {
	thevalidator := make(map[string]bool)
	thevalidator[wallet] = true
	*Validators = append(*Validators, thevalidator)
	return true
}

func (is InternalService) getValidators() ([]string, error) {
	//{"collection":"MessagesPool", "select": { "vm_response.V": {"$ne":null} } , "options": {} }
	_, blockchainData := is.Storage.Get("MessagesPool", bson.M{"vm_response.V": bson.M{"$ne": nil}}, bson.M{})
	var jsonBlockchainData JSONRESP
	err := json.Unmarshal(blockchainData, &jsonBlockchainData)
	if err != nil {
		fmt.Println("datERROR", err)
		return nil, err
	}
	var Validators []string
	for _, el := range jsonBlockchainData.Result {
		for _, d := range el.VMResponse.V {
			for validator, valid := range d {
				if valid {
					Validators = append(Validators, validator)
				}
			}
		}
	}
	return Validators, nil
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

func (is InternalService) Register(address string, item RegType, Register *[]map[string]RegType) bool {
	data := map[string]RegType{
		address: item,
	}
	*Register = append(*Register, data)
	return true
}

func (is InternalService) isRegistered(item string) bool {
	//bson.M{"vm_response.R": bson.M{"$elemMatch": bson.M{"123": bson.M{"$exists": 1}}}}
	_, blockhainData := is.Storage.Get("MessagesPool", bson.M{"vm_response.R": bson.M{"$elemMatch": bson.M{item: bson.M{"$exists": 1}}}, "vm_processed": true, "vm_result": true}, bson.M{})
	var jsonBlockchainData JSONRESP
	err := json.Unmarshal(blockhainData, &jsonBlockchainData)
	if err != nil {
		fmt.Println("datERROR", err)
		return false
	}
	fmt.Println("IS registered", jsonBlockchainData, ">>>", bson.M{"vm_response.R": bson.M{"$elemMatch": bson.M{item: bson.M{"$exists": 1}}}, "vm_processed": true, "vm_result": true})
	if len(jsonBlockchainData.Result) > 0 {
		return true
	}
	return false
}

func (is InternalService) addTokenBalance(token, address string, amount int64, CustomTokens *[]map[string][]map[string]int64) (int64, error) {
	if !is.isRegistered(token) {
		return 0, errors.New("not registered")
	}
	data := map[string][]map[string]int64{
		token: []map[string]int64{
			{address: amount},
		},
	}
	fmt.Println("Add token balance FN", CustomTokens)
	*CustomTokens = append(*CustomTokens, data)
	fmt.Println("Add token balance FN append", CustomTokens)
	return amount, nil
}

func (is InternalService) ammAddTokenBalance(token string, amount int64, CustomTokens *[]map[string][]map[string]int64) (int64, error) {
	if !is.isRegistered("amm_" + token) {
		return 0, nil
	}
	res, err := is.addTokenBalance(token, "amm_"+token, amount, CustomTokens)
	if err != nil {
		return 0, err
	}
	return res, nil
}

func (is InternalService) ammExchange(wallet, token string, amount int64, CustomTokens *[]map[string][]map[string]int64, Distribution *[]map[string]int64) bool {
	exchRate, ammTokenBalance, ammBalance := is.ammGetExchangeRate(token)
	if exchRate == 0 {
		return false
	}
	theWalletTokenBalance, err := is.getTokenBalance(token, wallet)
	if err != nil {
		return false
	}
	if theWalletTokenBalance-amount < 0 {
		return false
	}
	if amount < 0 && ammBalance-amount*int64(exchRate) > 0 {
		ammTokenBalance := 0 - amount              // plus
		ammBalance := 0 - amount*int64(exchRate)   //minus
		walletTokenBalance := amount               // minus
		walletBalance := -amount * int64(exchRate) // plus

		_, err := is.addTokenBalance(token, wallet, walletTokenBalance, CustomTokens)
		if err != nil {
			return false
		}
		_, err = is.addTokenBalance(token, "amm_"+token, ammTokenBalance, CustomTokens)
		if err != nil {
			return false
		}
		balance := make(map[string]int64)
		balance["amm_"+token] = ammBalance
		*Distribution = append(*Distribution, balance)
		//updateInProcessingBalance(request.Block,*Distribution)
		balance = make(map[string]int64)
		balance[wallet] = walletBalance
		*Distribution = append(*Distribution, balance)
		//updateInProcessingBalance(request.Block,*Distribution)
	}
	if amount > 0 && ammTokenBalance-amount*int64(exchRate) > 0 {
		ammTokenBalance := -amount                 // minus
		ammBalance := amount * int64(exchRate)     // plus
		walletTokenBalance := amount               // plus
		walletBalance := -amount * int64(exchRate) // minus

		_, err := is.addTokenBalance(token, wallet, walletTokenBalance, CustomTokens)
		if err != nil {
			return false
		}
		_, err = is.addTokenBalance(token, "amm_"+token, ammTokenBalance, CustomTokens)
		if err != nil {
			return false
		}
		balance := make(map[string]int64)
		balance["amm_"+token] = ammBalance
		*Distribution = append(*Distribution, balance)
		//updateInProcessingBalance(request.Block,*Distribution)
		balance = make(map[string]int64)
		balance[wallet] = walletBalance
		*Distribution = append(*Distribution, balance)
		//updateInProcessingBalance(request.Block,*Distribution)
	}
	return true
}

func (is InternalService) ammGetExchangeRate(token string) (float64, int64, int64) {
	balanceCoins, err := is.getBalance("amm_" + token)
	if err != nil {
		balanceCoins = 0
	}
	balanceTokens, err := is.getTokenBalance(token, "amm_"+token)
	if err != nil {
		balanceTokens = 0
	}
	var exchRate float64
	if balanceCoins == 0 || balanceTokens == 0 {
		exchRate = 0
	} else {
		exchRate = float64(balanceTokens / balanceCoins)
	}
	return exchRate, balanceTokens, balanceCoins
}

func fetchJSfunctions(code string) []string {
	//code := "function greet() {\n console.log('Hello, World!'); \n}\nfunction bye() {\n console.log('Goodbye!'); \n}"
	functionDeclarationRegex := regexp.MustCompile(`function\s+(\w+)\s*\(`)
	functions := functionDeclarationRegex.FindAllStringSubmatch(code, -1)
	var functionsLit []string
	for _, match := range functions {
		functionsLit = append(functionsLit, match[1])
	}
	return functionsLit
}

type Result struct {
	ID           string `json:"_id"`
	BlockHash    string `json:"block_hash"`
	BlockNumber  int    `json:"block_number"`
	ExecutedHash string `json:"executed_hash"`
	Message      struct {
		Message         string `json:"message"`
		MessageHash     string `json:"message_hash"`
		Nonce           int    `json:"nonce"`
		SenderAddress   string `json:"sender_address"`
		SenderSignature string `json:"sender_signature"`
		Type            string `json:"type"`
	} `json:"message"`
	MessageHash string `json:"message_hash"`
	VMProcessed bool   `json:"vm_processed"`
	VMResponse  struct {
		C json.RawMessage    `json:"C"`
		D []map[string]int64 `json:"D"`
		V []map[string]bool  `json:"V"`
		R json.RawMessage    `json:"R"`
		T json.RawMessage    `json:"T"`
	} `json:"vm_response"`
	VMResult bool  `json:"vm_result"`
	Votes    []int `json:"votes"`
}

type JSONRESP struct {
	Result []Result `json:"result"`
}

func generateRandomNumbers(distType string, numbersInSet int, min, max, param float64, baseRand int64) []float64 {

	//distType := "pareto"
	//param := 1.5
	//generateRandomNumbers(distType, param)
	//rand.Seed(time.Now().UnixNano())
	fmt.Println("BGIN rndset")
	var set []float64
	rand.Seed(baseRand)
	for i := 0; i < numbersInSet; i++ {
		var num float64
		switch distType {
		case "uniform":
			num = rand.Float64()*float64(max-min) + min
			fmt.Println("NUM unifirm:", num)
		case "exponential":
			num = rand.ExpFloat64()*float64(max-min) + min
		case "normal":
			mean := 50.0
			stdDev := 15.0
			num = math.Round(rand.NormFloat64()*stdDev + mean)
			if num < min {
				num = min
			} else if num > max {
				num = max
			}
		//https://go-recipes.dev/generating-random-numbers-with-go-616d30ccc926
		case "poisson":
			for {
				x := rand.ExpFloat64() / param
				if x >= 1.0 {
					continue
				}
				p := 1.0
				for k := 0; k <= int(num); k++ {
					p *= x * param / float64(k+1)
				}
				if rand.Float64() <= p {
					break
				}
				num++
			}
			num++
		//case "lognormal":
		//	num = rand.LogNormal(param, param)
		//case "pareto":
		//	num = rand.Pareto(param)
		//case "beta":
		//	num = rand.Beta(param, param)
		default:
			return nil
		}
		set = append(set, num)
	}
	fmt.Println("THE set ", set)
	return set
}
