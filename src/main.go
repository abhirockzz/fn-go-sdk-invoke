package main

import (
	"flag"
	"os"

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
	region := "us-phoenix-1" //Functions available only in phoenix during LA
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

	functionsUtil := newFunctionsUtil(tenantOCID, userOCID, region, fingerprint, privateKeyLocation, privateKeyPassphrase)
	compartment, err := functionsUtil.getCompartment(*compartmentName)
	if err != nil {
		panic(err)
	}
	application, err := functionsUtil.getApplication(*appName, *compartment)
	if err != nil {
		panic(err)
	}
	function, err := functionsUtil.getFunction(*funcName, *application)
	if err != nil {
		panic(err)
	}
	functionsUtil.invokeFunction(*function, *invokePayload)
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
