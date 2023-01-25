package internal

import (
	"encoding/json"
	"fmt"
	"github.com/robertkrimen/otto"
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

func (is InternalService) execute(data interface{}) interface{} {
	counter++
	var Validators []map[string]bool
	var Distribution []map[string]int64
	var CustomFld []map[string]string
	var Fee float64
	var request VMrequest
	fmt.Println("REQUEST:::::::", data)

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
		saveScriptData := make(map[string]string)
		saveScriptData[request.Tx.MessageHash] = "saved"
		CustomFld = append(CustomFld, saveScriptData)
		return bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"callNumber": counter, "D": Distribution, "V": Validators, "C": CustomFld, "F": Fee}}
	case "run":
		// Curl example:
		// curl --location --request GET 'http://185.229.119.188:8018' \
		//--header 'Content-Type: text/plain' \
		//--data-raw '{"method":"get-tx","data":"{\"method\": \"run\", \"data\": \"f71aec98f3fab3f8ab3af2131465173194fbe45fcb82e07f49b98ef18a10e03b.hello('\''world'\'')  ;\"}"}'
		parts := strings.Split(vmScript.Script, ".")
		Contract := parts[0]
		Function := parts[1]
		currentContract = Contract
		runScript, _ := getMessageByHash(Contract, is)
		logRecord := make(map[string]string)
		logRecord[Contract] = "contract was run with " + Function
		CustomFld = append(CustomFld, logRecord)
		vmScript.Script = runScript + " " + Function
	}

	theSender := request.Tx.SenderAddress // request.Tx.Sender
	theScriptHash := request.Tx.MessageHash
	fmt.Println("XXXXX", vmScript.Script)
	if vmScript.Script == "fly me to the moon" && request.Block == 1 {
		fmt.Println("L51")
		thevalidator := make(map[string]bool)
		thevalidator[theSender] = true
		Validators = append(Validators, thevalidator)

		fmt.Println("L56")
		initbalance := make(map[string]int64)
		initbalance[theSender] = int64(1000)
		Distribution = append(Distribution, initbalance)
		initbalance = make(map[string]int64)
		initbalance["139uwuYCM1knfLdyVX2yjzwhDDz73Zx7Sj"] = int64(1000)
		Distribution = append(Distribution, initbalance)
		initbalance = make(map[string]int64)
		initbalance["1PKVs1mizz4abZ8zvk4gbUNdSvmMXTFfEh"] = int64(1000)
		Distribution = append(Distribution, initbalance)

		fmt.Println("L67")
		initsettings := make(map[string]string)
		initsettings["FeePerMessageSymbol"] = "0.01"
		CustomFld = append(CustomFld, initsettings)
		initsettings = make(map[string]string)
		initsettings["Fee_getBalance"] = "0.05"
		CustomFld = append(CustomFld, initsettings)
		initsettings = make(map[string]string)
		initsettings["FeeSaveDataPerSymbol"] = "0.01"
		CustomFld = append(CustomFld, initsettings)

		fmt.Println("L78")
		fmt.Println("RETURN::::::::::::", bson.M{"GENESYS": "GENESYS", "vm_processed": true, "vm_result": true, "vm_response": bson.M{"callNumber": counter, "D": Distribution, "V": Validators, "C": CustomFld}})
		return bson.M{"GENESYS": "GENESYS", "vm_processed": true, "vm_result": true, "vm_response": bson.M{"callNumber": counter, "D": Distribution, "V": Validators, "C": CustomFld}}
	}

	vm := otto.New()
	vm.Set("getRnd", func(call otto.FunctionCall) otto.Value {
		res, _ := vm.ToValue(request.Rnd)
		return res
	})

	vm.Set("callMessagesPool", func(call otto.FunctionCall) otto.Value {
		Query, _ := call.Argument(0).ToString()
		Options, _ := call.Argument(1).ToString()
		var theQuery interface{}
		var theOptions interface{}
		_ = json.Unmarshal([]byte(Query), &theQuery)
		_ = json.Unmarshal([]byte(Options), &theOptions)
		_, blockhainData := is.Storage.Get("MessagesPool", theQuery, theOptions)
		fmt.Println("blockhainData", string(blockhainData))
		res, _ := vm.ToValue(string(blockhainData))
		return res
	})

	vm.Set("getValidators", func(call otto.FunctionCall) otto.Value {
		// {"collection":"MessagesPool","options":{},"select":{"vm_response.V": {"$ne" : null}  }}
		res, _ := vm.ToValue("Validators")
		return res
	})

	vm.Set("getBalance", func(call otto.FunctionCall) otto.Value {
		Wallet, _ := call.Argument(0).ToString()
		WalletBalance, err := is.getBalance(Wallet)
		if err != nil {
			res, _ := vm.ToValue(false)
			return res
		} else {
			res, _ := vm.ToValue(WalletBalance)
			return res
		}
	})

	vm.Set("addValidator", func(call otto.FunctionCall) otto.Value {
		fmt.Println("Add validator")
		validatorWallet, _ := call.Argument(0).ToString()
		//fmt.Println("Hello, addBalance world!", fnWrapper)
		if getValidatorWalletLicence(validatorWallet) {
			thevalidator := make(map[string]bool)
			thevalidator[validatorWallet] = true
			Validators = append(Validators, thevalidator)
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

	vm.Set("transfer", func(call otto.FunctionCall) otto.Value {
		to, _ := call.Argument(0).ToString()
		amount, _ := call.Argument(1).ToInteger()
		WalletBalance, _ := is.getBalance(request.Tx.SenderAddress)
		if (WalletBalance - amount) > 0 {
			balance := make(map[string]int64)
			balance[request.Tx.SenderAddress] = int64(0 - amount)
			Distribution = append(Distribution, balance)
			balance = make(map[string]int64)
			balance[to] = int64(amount)
			Distribution = append(Distribution, balance)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})
	vm.Set("transferFromContract", func(call otto.FunctionCall) otto.Value {
		if currentContract == "" {
			return otto.FalseValue()
		}
		to, _ := call.Argument(0).ToString()
		amount, _ := call.Argument(1).ToInteger()
		contractBalance, _ := is.getBalance(currentContract)
		if (contractBalance - amount) > 0 {
			balance := make(map[string]int64)
			balance[request.Tx.SenderAddress] = int64(0 - amount)
			Distribution = append(Distribution, balance)
			balance = make(map[string]int64)
			balance[to] = int64(amount)
			Distribution = append(Distribution, balance)
			return otto.TrueValue()
		} else {
			return otto.FalseValue()
		}
	})

	vm.Set("currentFee", func(call otto.FunctionCall) otto.Value {
		return otto.Value{}
	})
	//vm.SetTimeout(time.Second)
	result, err := vm.Run(vmScript.Script)
	if err != nil {
		fmt.Println("error", err)
		return bson.M{"vm_processed": true, "vm_result": false}
	}

	fmt.Println(result)
	CustomFldElement := make(map[string]string)
	CustomFldElement[theScriptHash], _ = result.ToString()
	CustomFld = append(CustomFld, CustomFldElement)
	fmt.Println("RETURN ::::: ", bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"callNumber": counter, "D": Distribution, "V": Validators, "C": CustomFld}})
	return bson.M{"vm_processed": true, "vm_result": true, "vm_response": bson.M{"callNumber": counter, "D": Distribution, "V": Validators, "C": CustomFld, "F": Fee}}
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

func getMessageByHash(MessageHash string, is InternalService) (string, error) {
	// {"collection":"MessagesPool", "select": {"block_number": {"$exists":true}, "message_hash": "d8e7f63670d4e8ff434d031de226609bc1cb64eeae3ee496553f4cabb22a8c64","vm_processed": true,"vm_result": true}, "options": {} }
	_, blockhainData := is.Storage.Get("MessagesPool", bson.M{"block_number": bson.M{"$exists": true}, "message_hash": MessageHash, "vm_processed": true, "vm_result": true}, bson.M{})
	fmt.Println("blockhainData", string(blockhainData))
	var jsonBlockchainData JSONRESP
	err := json.Unmarshal(blockhainData, &jsonBlockchainData)
	if err != nil {
		fmt.Println("datERROR", err)
		return "", err
	}
	message := jsonBlockchainData.Result[0].Message.Message
	var vmScript VMscript
	err = json.Unmarshal([]byte(message), &vmScript)
	if err != nil {
		return "", err
	} else {
		return vmScript.Script, nil
	}
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
		C          json.RawMessage    `json:"C"`
		D          []map[string]int64 `json:"D"`
		V          []map[string]bool  `json:"V"`
		CallNumber int                `json:"callNumber"`
	} `json:"vm_response"`
	VMResult bool  `json:"vm_result"`
	Votes    []int `json:"votes"`
}

type JSONRESP struct {
	Result []Result `json:"result"`
}
