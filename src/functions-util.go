package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/oracle/oci-go-sdk/identity"
)

//Contains helper methods for invoking a function
type functionsUtil struct {
	tenantOCID            string
	functionsInvokeClient functions.FunctionsInvokeClient
	functionsMgtClient    functions.FunctionsManagementClient
	identityClient        identity.IdentityClient
}

//Initializes required clients for Functions and Identity
func newFunctionsUtil(tenantOCID, userOCID, region, fingerprint, privateKeyLocation, privateKeyPassphrase string) functionsUtil {
	var err error

	fmt.Println("Reading private key", privateKeyLocation)
	privateKey, err := ioutil.ReadFile(privateKeyLocation)
	if err != nil {
		panic("Unable to read private key file contents from " + privateKeyLocation + " due to " + err.Error())
	}

	//instantiate required clients
	configProvider := common.NewRawConfigurationProvider(tenantOCID, userOCID, region, fingerprint, string(privateKey), common.String(privateKeyPassphrase))
	identityClient, err = identity.NewIdentityClientWithConfigurationProvider(configProvider)
	if err != nil {
		panic("Could not instantiate Identity client - " + err.Error())
	}

	functionsInvokeClient, err = functions.NewFunctionsInvokeClientWithConfigurationProvider(configProvider)
	if err != nil {
		panic("Could not instantiate Functions Invoke client - " + err.Error())
	}

	functionsMgtClient, err = functions.NewFunctionsManagementClientWithConfigurationProvider(configProvider)
	if err != nil {
		panic("Could not instantiate Functions Management client - " + err.Error())
	}
	functionsMgtClient.SetRegion("us-phoenix-1") //Functions available only in phoenix during LA
	return functionsUtil{tenantOCID, functionsInvokeClient, functionsMgtClient, identityClient}
}

//Invokes a function
func (util functionsUtil) invokeFunction(function functions.FunctionSummary, payload string) {
	//client needs to pointed to the unique (invoke) endpoint specific to a function
	functionsInvokeClient.Host = *function.InvokeEndpoint

	fmt.Println("Invoking function endpoint " + functionsInvokeClient.Host + " with payload " + payload)

	functionPayload := ioutil.NopCloser(bytes.NewReader([]byte(payload)))
	//functionPayload, _ := os.Open("/home/foo/cat.jpeg")
	//defer functionPayload.Close()
	invokeFunctionReq := functions.InvokeFunctionRequest{FunctionId: function.Id, InvokeFunctionBody: functionPayload}

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

//Returns details for a function given its name, the details of the application it belongs to. This is an expensive operation and the results should be cached.
func (util functionsUtil) getFunction(functionName string, application functions.ApplicationSummary) (*functions.FunctionSummary, error) {
	fmt.Println("Finding details for function " + functionName + " in application " + *application.DisplayName)

	listFunctionsRequest := functions.ListFunctionsRequest{ApplicationId: application.Id, DisplayName: common.String(functionName)}
	listFunctionsResp, err := functionsMgtClient.ListFunctions(context.Background(), listFunctionsRequest)
	if err != nil {
		//fmt.Println("Could not list functions due to", err.Error())
		return nil, err
	}

	if len(listFunctionsResp.Items) == 0 {
		return nil, errors.New("Could not find function " + functionName + " in application " + *application.DisplayName)
	}
	fmt.Println("Found details for function", functionName)
	return &listFunctionsResp.Items[0], nil
}

//Returns Application details, given its name and Compartment details
func (util functionsUtil) getApplication(appName string, compartment identity.Compartment) (*functions.ApplicationSummary, error) {
	fmt.Println("Finding details for application " + appName + " in compartment " + *compartment.Name)

	listAppsRequest := functions.ListApplicationsRequest{CompartmentId: compartment.Id, DisplayName: common.String(appName)}
	listAppsResp, err := functionsMgtClient.ListApplications(context.Background(), listAppsRequest)
	if err != nil {
		return nil, err
	}

	if len(listAppsResp.Items) == 0 {
		return nil, errors.New("Could not find application " + appName + " in compartment " + *compartment.Name + " within tenancy " + util.tenantOCID)
	}

	app := listAppsResp.Items[0]
	fmt.Println("Found details for application", appName)
	return &app, nil
}

//Returns Compartment details given its name. Also uses the tenancy OCID info.
func (util functionsUtil) getCompartment(compartmentName string) (*identity.Compartment, error) {
	fmt.Println("Finding details for compartment " + compartmentName + " in tenancy " + util.tenantOCID)

	//To get a full list of all compartments and subcompartments in the tenancy (root compartment), set the parameter `compartmentIdInSubtree` to true and `accessLevel` to ANY.
	//details - https://godoc.org/github.com/oracle/oci-go-sdk/identity#IdentityClient.ListCompartments
	listCompartmentsRequest := identity.ListCompartmentsRequest{CompartmentId: common.String(util.tenantOCID), CompartmentIdInSubtree: common.Bool(true), AccessLevel: identity.ListCompartmentsAccessLevelAny}
	listCompartmentsResp, err := identityClient.ListCompartments(context.Background(), listCompartmentsRequest)

	if err != nil {
		return nil, err
	}
	for _, compartment := range listCompartmentsResp.Items {
		if *compartment.Name == compartmentName {
			fmt.Println("Found details for compartment", compartmentName)
			return &compartment, nil
		}
	}

	return nil, errors.New("Could not find details for compartment " + compartmentName + " in tenancy " + util.tenantOCID)
}
