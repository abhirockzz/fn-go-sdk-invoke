package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/oracle/oci-go-sdk/functions"

	"github.com/oracle/oci-go-sdk/common"
)

func invokeFunction(functionID, invokeEndpoint, payload string) {

	fmt.Println("Invoking function endpoint " + invokeEndpoint + " with payload ")

	functionPayload := ioutil.NopCloser(bytes.NewReader([]byte(payload)))
	//functionPayload, _ := os.Open("/home/foo/cat.jpeg")
	//defer functionPayload.Close()
	invokeFunctionReq := functions.InvokeFunctionRequest{FunctionId: common.String(functionID), InvokeFunctionBody: functionPayload}

	//client needs to pointed to the unique (invoke) endpoint specific to a function
	functionsInvokeClient.Host = invokeEndpoint
	invokeFunctionResp, err := functionsInvokeClient.InvokeFunction(context.Background(), invokeFunctionReq)
	if err != nil {
		fmt.Println("Function invocation failed due to", err.Error())
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(invokeFunctionResp.Content)
	functionResponse := buf.String()

	fmt.Println("Function response", functionResponse)
}
