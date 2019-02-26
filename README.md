# Invoke Oracle Functions using the OCI Go SDK

This example demonstrates how to invoke a function on Oracle Functions using (preview version of) the Oracle Cloud Infrastructure Go SDK. 

## Introduction

To be specifc, it shows how you can invoke a function by its name given that you also provide the application name (to which the function belongs), the OCI compartment (for which your Oracle Functions service is configured) and the OCID of your tenancy

OCI SDK exposes two endpoints for Oracle Functions

- `functions.FunctionsManagementClient` - for CRUD operations e.g. creating applications, listing functions etc.
- `functions.FunctionsInvokeClient` - only required for invoking a function

The `invokeFunction` function in `functions.FunctionsInvokeClient` requires the function OCID and the function invoke endpoint which needs to be extracted using the following - function name, application name, the compartment name and the tenant OCID. This involves multiple API calls

- The first step extracts the Compartment OCID from the name using `ListCompartments` function in `identity.IdentityClient` - it looks for compartments in the tenancy and matches the one with the provided name
- The compartment OCID is then used to find the Application OCID from the name using `ListApplications` function exposed by `functions.FunctionsManagementClient`
- Once we have the application OCID, the function information (in the form of a `functions.FunctionSummary` object) is extracted using `ListFunctions` in `functions.FunctionsManagementClient` - this allows us to get both the function OCID as well as its invoke endpoint via a single API call

The key thing to note here is that the function ID and its invoke endpoint will not change unless you delete the function (or the application it's a part of). As a result you do not need to repeat the above mentioned flow of API calls - the funtion ID and its invoke endpoint can be derived once and then **cached** in-memory (e.g. in a `map`) or an external data store

Now that we have the function OCID and invoke enpoint at our disposal

- we build `functions.InvokeFunctionRequest` with the function OCID and the (optional) payload which we want to send to our function, and,
- point the `functions.FunctionsInvokeClient` towards the invoke endpoint 
- finally, we call `InvokeFunction` and extract the response from `invokeFunctionResp.InvokeFunctionResponse`

### Authentication

The client program needs to authenticate to OCI before being able to make service calls. The standard OCI authenitcation is used, which accepts the following inputs (details below) - tenant OCID, user OCID, fingerprint, private key and passphrase (optional). These details are required to instantiate a `common.ConfigurationProvider` using `common.NewRawConfigurationProvider` method and subsequently used by the service client objects (`functions.FunctionsInvokeClient`, `functions.FunctionsManagementClient`, `identity.IdentityClient`)

This example does not assume the presence of an OCI config file on the machine from where this is being executed. However, if you have one present as per the standard OCI practices i.e. a config file in your home directory, you can use `common.DefaultConfigProvider` for convenience

## Pre-requisites

1. Install latest Fn CLI - 

`curl -LSs https://raw.githubusercontent.com/fnproject/cli/master/install | sh`

2. Create a function to invoke

Create a function using [Go Hello World Function](https://github.com/abhirockzz/oracle-functions-hello-worlds/blob/master/golang-hello-world.md)

### Install preview OCI Go SDK

You need to add the OCI Go SDK to your `GOPATH`

`export GOPATH=<your GOPATH>` e.g. `export GOPATH=/Users/foobar/go`

1. Back up existing installation of OCI Go SDK and remove it from your `GOPATH` - `rm -rf $GOPATH/src/github.com/oracle/oci-go-sdk`

2. Download and unzip the preview version of the OCI Go SDK

`unzip <to-be-filled>.zip`

3. Copy the preview SDK to your `GOPATH`

`cp -R <to-be-filled>/ $GOPATH/src/github.com/oracle/oci-go-sdk`

## You can now invoke your function!

Clone this repository - 

`git clone https://github.com/abhirockzz/fn-go-sdk-invoke`

Change to the correct directory where you cloned the example: 

`cd fn-go-sdk-invoke`

Set environment variables

	export TENANT_OCID=<OCID of your tenancy>
	export USER_OCID=<OCID of the OCI user>
	export PUBLIC_KEY_FINGERPRINT=<public key fingerprint>
	export PRIVATE_KEY_LOCATION=<location of the private key on your machine>

> please note that `PASSPHRASE` is optional i.e. only required if your private key has one

	export PASSPHRASE=<private key passphrase>

e.g. 

	export TENANT_OCID=ocid1.tenancy.oc1..aaaaaaaaydrjd77otncda2xn7qrv7l3hqnd3zxn2u4siwdhniibwfv4wwhtz
	export USER_OCID=ocid1.user.oc1..aaaaaaaavz5efd7jwjjipbvm536plgylg7rfr53obvtghpi2vbg3qyrnrtfa
	export PUBLIC_KEY_FINGERPRINT=42:42:5f:42:ca:a1:2e:58:d2:63:6a:af:42:d5:3d:42
	export PRIVATE_KEY_LOCATION=/Users/foobar/oci_api_key.pem
	
> and only if your private key has a passphrase:

	export PASSPHRASE=4242

`go run src/*.go --compartmentName=<compartmentName> --appName=<appName> --funcName=<funcName> --invokePayload=<(optional) invokePayload>`

> Payload is optional. If your function doesn't expect any input parameters, you can omit `--invokePayload` argument

e.g. with payload:

`go run src/*.go --compartmentName=mycompartment --appName=helloworld-app --funcName=helloworld-go --invokePayload={\"name\":\"foobar\"}`

e.g. without payload:

`go run src/*.go --compartmentName=testcomp --appName=helloworld-app --funcName=helloworld-go`

## What if my function needs input in binary form ?

This example demonstrates how to invoke a boilerplate function which accepts (an optional) string payload (JSON data). But, it is possible to send binary payload as well.

You can use this Tensorflow based function as an example to explore the possibility of invoking a function using binary content - https://github.com/abhirockzz/fn-hello-tensorflow. This function expects the image data (in binary form) as an input and returns what object that image resembles along with the percentage accuracy

If you were to deploy the above function and invoke it using Fn CLI, the command would something like this - `cat /home/foo/cat.jpeg | fn invoke fn-tensorflow-app classify`. In this case, the `cat.jpeg` image is being passed as an input to the function. The programmatic (using Go SDK) equivalent of this would look something like the below snippet, where the function invocation request (`functions.InvokeFunctionRequest`) is being built along with the binary input (image file content)

    functionPayload, _ := os.Open("/home/foo/cat.jpeg") //error handling ignored for brevity
	defer functionPayload.Close()
	invokeFunctionReq := functions.InvokeFunctionRequest{FunctionId: functionID, InvokeFunctionBody: functionPayload}

Pay attention to the following line `functionPayload, _ := os.Open("/home/foo/cat.jpeg")`. The `InvokeFunctionBody` accepts an `io.ReadCloser` - as a result, it is possible to use `os.File` (or any other compatible type) as a function payload

## Troubleshooting

### If you fail to set the required environment variables like `TENANT_OCID` etc.

You will see the following error - `panic: Please set the environment variable TENANT_OCID`

### If you do not provide required arguments i.e. function name etc.

You will see the following error - `panic: Please set mandatory flag compartmentName`

### If you provide an invalid value for function name etc.

You will see something similar to - `Could not find OCID for application incorrect-app-name in compartmentName mycompartment within tenancy ocid1.tenancy.oc1..aaaaaaaaycrjm42etncqa2xn7qtv2l3hqnd3zxn2u4siwdhniibwfv4rrhta`

### If you provide an incorrect `TENANT_OCID` or `USER_OCID` or `PUBLIC_KEY_FINGERPRINT`

You will get this error - `Service error:NotAuthenticated. The required information to complete authentication was not provided or was incorrect.. http status code: 401. Opc request id: 4d7d76463a4f437361a1eb47815015d9/450D86C860C0592949AFF344C6DB17F0/84F3CCC7542175C08646CD4E213A2CEF`

### If your key has a passphrase but you fail to set the environment variable PASSPHRASE or provid an incorrect value

You will get this error - `panic: Could not instantiate Identity client - can not create client, bad configuration: x509: decryption password incorrect`

