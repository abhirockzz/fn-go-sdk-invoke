package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/oracle/oci-go-sdk/identity"
)

//Returns info for a function given its name and the application it belongs to
func getFunction(compartmentName, appName, functionName, tenantOCID string) (functions.FunctionSummary, error) {
	fmt.Println("Finding info for function " + functionName + " in app " + appName + " belonging to compartment " + compartmentName + " within tenancy " + tenantOCID)
	appOCID, err := getApplicationOCID(appName, compartmentName, tenantOCID)
	if err != nil {
		//fmt.Println("Could not list applications due to", err.Error())
		return functions.FunctionSummary{}, err
	}
	listFunctionsRequest := functions.ListFunctionsRequest{ApplicationId: common.String(appOCID), DisplayName: common.String(functionName)}
	listFunctionsResp, err := functionsMgtClient.ListFunctions(context.Background(), listFunctionsRequest)
	if err != nil {
		//fmt.Println("Could not list functions due to", err.Error())
		return functions.FunctionSummary{}, err
	}

	if len(listFunctionsResp.Items) == 0 {
		return functions.FunctionSummary{}, errors.New("Could not find function " + functionName + " in app " + appName)
	}

	return listFunctionsResp.Items[0], nil
}

//Returns compartment OCID given compartment name and tenancy OCID
func getCompartmentOCID(compartmentName string, tenantOCID string) (string, error) {
	fmt.Println("Finding OCID for compartment " + compartmentName + " in tenancy " + tenantOCID)

	//To get a full list of all compartments and subcompartments in the tenancy (root compartment), set the parameter `compartmentIdInSubtree` to true and `accessLevel` to ANY.
	//details - https://godoc.org/github.com/oracle/oci-go-sdk/identity#IdentityClient.ListCompartments
	listCompartmentsRequest := identity.ListCompartmentsRequest{CompartmentId: common.String(tenantOCID), CompartmentIdInSubtree: common.Bool(true), AccessLevel: identity.ListCompartmentsAccessLevelAny}
	listCompartmentsResp, err := identityClient.ListCompartments(context.Background(), listCompartmentsRequest)

	if err != nil {
		return "", err
	}
	for _, compartment := range listCompartmentsResp.Items {
		if *compartment.Name == compartmentName {
			fmt.Println("OCID for compartment "+compartmentName, *compartment.Id)
			return *compartment.Id, nil
		}
	}

	return "", errors.New("Could not find OCID for compartment " + compartmentName + " in tenancy " + tenantOCID)
}

//Returns app OCID, provided the name of the app and its specific compartment and tenancy
func getApplicationOCID(appName, compartmentName, tenantOCID string) (string, error) {
	fmt.Println("Finding OCID for application " + appName + " in compartmentName " + compartmentName + " within tenancy " + tenantOCID)
	compartmentOCID, err := getCompartmentOCID(compartmentName, tenantOCID)
	if err != nil {
		//fmt.Println("Could not list compartments due to", err.Error())
		return "", err
	}
	listAppsRequest := functions.ListApplicationsRequest{CompartmentId: common.String(compartmentOCID), DisplayName: common.String(appName)}
	listAppsResp, err := functionsMgtClient.ListApplications(context.Background(), listAppsRequest)
	if err != nil {
		return "", err
	}

	if len(listAppsResp.Items) == 0 {
		return "", errors.New("Could not find OCID for application " + appName + " in compartmentName " + compartmentName + " within tenancy " + tenantOCID)
	}

	appOCID := listAppsResp.Items[0].Id
	fmt.Println("OCID for application "+appName, *appOCID)
	return *appOCID, nil
}
