package main

import (
    "net/http"
    "net/url"
    "encoding/json"
    "bytes"
    "log"
    //"io/ioutil"
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
    infoCache map[string]interface{}
}

func NewOctopartClient(apikey string) *OctopartClient {
    oc := &OctopartClient{ApiKey: apikey,
                          RemoteHost: "https://octopart.com"}
    oc.client = &http.Client{}
    oc.infoCache = make(map[string]interface{})
    return oc
}

func openPricingSource() {
    // TODO: pass through octopart API key here
    pricingSource = NewOctopartClient("")
}

func (oc *OctopartClient) apiCall(method string, params map[string]string) (map[string]interface{}, error) {
    paramString := "?apikey=" + oc.ApiKey
    // TODO: assert clean-ness of params
    // TODO: use url.Values instead...
    for key := range params {
        paramString += "&" + url.QueryEscape(key) + "=" + url.QueryEscape(params[key])
    }
    paramStringUnescaped, _ := url.QueryUnescape(paramString) // TODO: err
    log.Println("Fetching: " + oc.RemoteHost + "/api/v2/" + method + paramStringUnescaped)
    resp, err := oc.client.Get(oc.RemoteHost + "/api/v2/" + method + paramString)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode != 200 {
        return nil, Error("Octopart API call error: " + resp.Status)
    }
    result := make(map[string]interface{})
    defer resp.Body.Close()
    //body, err := ioutil.ReadAll(resp.Body)
    //if err != nil {
    //    return nil, err
    //}
    //body = append(body, '\n')
    //dec := json.NewDecoder(bytes.NewReader(body))
    dec := json.NewDecoder(resp.Body)
    if err = dec.Decode(&result); err != nil {
        return nil, err
    }
    return result, nil
}

// this method doesn't check internal cache, but it does append to it
func (oc *OctopartClient) bomApiCall(manufacturers, mpns []string) ([]map[string]interface{}, error) {
    // TODO: check len(mpns) == len(manufacturers) 
    queryList := make([]map[string]string, len(mpns))
    listItem := make(map[string]string)
    for i, _ := range mpns {
        listItem = make(map[string]string)
        listItem["mpn_or_sku"] = mpns[i]
        listItem["manufacturer"] = manufacturers[i]
        listItem["limit"] = "1"
        listItem["reference"] = manufacturers[i] + "|" + mpns[i]
        queryList[i] = listItem
    }

    linesBuffer := new(bytes.Buffer)
    enc := json.NewEncoder(linesBuffer)
    if err := enc.Encode(queryList); err != nil {
        return nil, err
    }

    response, err := oc.apiCall("bom/match", map[string]string{"lines": linesBuffer.String()})
    if err != nil {
        return nil, err
    }
    // TODO: just grabbing first result for now; user can make better specification later
    ret := make([]map[string]interface{}, len(mpns))
    for i, rawresult := range response["results"].([]interface{}) {
        result := rawresult.(map[string]interface{})
        hits := int(result["hits"].(float64))
        reference := result["reference"].(string)
        if hits == 0 {
            ret[i] = nil
            oc.infoCache[reference] = nil
        } else {
            ret[i] = result["items"].([]interface{})[0].(map[string]interface{})
            oc.infoCache[reference] = ret[i]
        }
    }
    return ret, nil
}

func (oc *OctopartClient) GetMarketInfo(manufacturers, mpns []string) ([]interface{}, error) {
    if len(mpns) < 1 {
        return nil, Error("no mpns strings passed in")
    }
    if len(mpns) != len(manufacturers) {
        return nil, Error("number of mpns doesn't match number of manufacturers")
    }
    if len(mpns) > 100 {
        return nil, Error("can't handle more than 100 queries at a time (yet)")
    }
    mpnToQuery := make([]string, 0)
    manufacturersToQuery := make([]string, 0)
    queryHash := ""
    // check for queryHashes in internal cache
    for i, _ := range mpns {
        queryHash = manufacturers[i] + "|" + mpns[i]
        if _, hasKey := oc.infoCache[queryHash]; hasKey != true {
            manufacturersToQuery = append(manufacturersToQuery, manufacturers[i])
            mpnToQuery = append(mpnToQuery, mpns[i])
        }
    }
    // if necessary, fetch missing queryHashes to internal cache
    if len(mpnToQuery) > 0 {
        if _, err := oc.bomApiCall(manufacturersToQuery, mpnToQuery); err != nil {
            return nil, err
        }
    }
    // construct list of return info
    result := make([]interface{}, len(mpns))
    for i, _ := range mpns {
        queryHash = manufacturers[i] + "|" + mpns[i]
        value, hasKey := oc.infoCache[queryHash]
        if hasKey != true {
            return nil, Error("key should be in cache, but isn't: " + queryHash)
        }
        result[i] = value
    }
    return result, nil
}

func (oc *OctopartClient) GetReducedPricing(mpn, manufacturer string) (interface{}, error) {
    marketInfo, err := oc.GetMarketInfo([]string{mpn}, []string{manufacturer})
    if err != nil {
        return nil, err
    }
    // reduce marketInfo to pricing
    return marketInfo[0], nil
}

