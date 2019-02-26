package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/oracle/oci-go-sdk/identity"
)

var identityClient identity.IdentityClient
var functionsInvokeClient functions.FunctionsInvokeClient
var functionsMgtClient functions.FunctionsManagementClient

func main() {

	//read environment variables
	tenantOCID := getEnvVarValue("TENANT_OCID")
	userOCID := getEnvVarValue("USER_OCID")
	region := "us-phoenix-1"
	fingerprint := getEnvVarValue("PUBLIC_KEY_FINGERPRINT")
	privateKeyLocation := getEnvVarValue("PRIVATE_KEY_LOCATION")
	privateKeyPassphrase := os.Getenv("PASSPHRASE")

	//read flags
	compartmentName := flag.String("compartmentName", "", "Name of the compartment for Oracle Functions service")
	appName := flag.String("appName", "", "Oracle Functions application ame")
	funcName := flag.String("funcName", "", "Oracle Functions function name")
	invokePayload := flag.String("invokePayload", "", "(optional) Invocation payload for your function")

	flag.Parse()
	checkMandatoryFlags()

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

	function, err := getFunction(*compartmentName, *appName, *funcName, tenantOCID)

	if err != nil {
		panic(err)
	}
	functionID := function.Id
	invokeEndpoint := function.InvokeEndpoint
	//invoke the function
	invokeFunction(*functionID, *invokeEndpoint, *invokePayload)
}

func getEnvVarValue(varName string) string {
	value := os.Getenv(varName)
	if value == "" {
		panic("Please set the environment variable " + varName)
	}

	return value
}

func checkMandatoryFlags() {
	//invokePayload flag is NOT mandatory
	mandatoryFlags := []string{"compartmentName", "appName", "funcName"}
	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	for _, mandatoryFlag := range mandatoryFlags {
		if !setFlags[mandatoryFlag] {
			panic("Please set mandatory flag " + mandatoryFlag)
		}
	}
}
