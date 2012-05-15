package main

import (
    "net/http"
)

/*
Routines to make (cached) API calls to Octopart and merge results into BOM
LineItems.
*/

var pricingSource *OctopartClient

type OctopartClient struct {
    ApiKey string
    RemoteHost string
    client *http.Client
}

func NewOctopartClient(apikey string) *OctopartClient {
    oc := &OctopartClient{ApiKey: apikey,
                          RemoteHost: "https://www.octopart.com"}
    oc.client = &http.Client{}
    return oc
}

func openPricingSource() {
    pricingSource = NewOctopartClient("")
}

func (*oc OctopartClient) apiCall(method string, params map[string]string) (map[string]interface, error) {
    paramString := "?apikey=" + oc.ApiKey
    for key := range params {
        paramString += "&" + key + "=" + params[key]
    }
    resp, err := oc.client.Get(oc.RemoteHost + "/api/v2/" + method)

    // resp as json, or interpret as error
    return 
}

func (*oc OctopartClient) GetMarketInfo(mpn, manufacturer string) (map[string]interface, error) {

}

func (*oc OctopartClient) GetPricing(method string, params map[string]string) (map[string]interface, error) {

}
