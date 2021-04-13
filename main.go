package main

import (
	"encoding/json"
	"fmt"
	"github.com/fasthttp/router"
	"github.com/go-playground/validator/v10"
	"github.com/heckdevice/moxy/core"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"os"
	"strings"
)

const (
	jsontype = "application/json"
)

var (
	apiPathParams = [...]string{
		"serviceID",
		"apiID",
	}
)

// String - return string value of path param enum
func (pp PathParams) String() string { return apiPathParams[pp] }

// PathParams - Type to enumise API Path param strings
type PathParams int

const (
	// SERVICEID - serviceID path param
	SERVICEID PathParams = iota
	// APIID - apiID path param
	APIID
)

var (
	info     = map[string]interface{}{"ver": "1.0", "name": "moxy", "description": "Reverse Proxy with inbuilt mocking feature"}
	r        = router.New()
	validate = validator.New()
)

func initLogger() {
	log.SetOutput(os.Stdout)
	log.SetOutput(os.Stderr)
	log.SetFormatter(&log.JSONFormatter{})

}

func setContentType(ctx *fasthttp.RequestCtx, contentType string) {
	ctx.Response.Header.Set("Content-Type", contentType)
}

func writeJSONResponse(ctx *fasthttp.RequestCtx, data interface{}, responseCode *int) {
	setContentType(ctx, jsontype)
	resp, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	if responseCode != nil {
		ctx.SetStatusCode(*responseCode)
	}
	ctx.Write(resp)
}

func handleInternalError(ctx *fasthttp.RequestCtx, msg string) {
	ctx.Error(msg, fasthttp.StatusInternalServerError)
}

func defaultHandler(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("It's Alive!!!")
}

func infoHandler(ctx *fasthttp.RequestCtx) {
	writeJSONResponse(ctx, info, nil)
}

func parseAPIUrl(requestPath string) (*string, *string, error) {
	splitPath := strings.SplitAfter(requestPath, "/")
	if len(splitPath) < 3 {
		return nil, nil, fmt.Errorf("Invalid API Configuration. Configure path should in format /{serviceID}/{apiURL} or /{serviceID}/")
	}
	serviceID := strings.Split(splitPath[1], "/")[0]
	apiURL := strings.TrimPrefix(requestPath, fmt.Sprintf("/%v", serviceID))
	log.Info(fmt.Sprintf("Parsed Path Url - serviceID=%v, apiUrl=%v", serviceID, apiURL))
	return &serviceID, &apiURL, nil
}
func fetchAPIInvocationDetails(ctx *fasthttp.RequestCtx) (*string, *MockableRequest, error) {
	log.Info(fmt.Sprintf("Invoked verb=%v, path=%v, uri=%v", string(ctx.Method()), string(ctx.Path()), string(ctx.RequestURI())))
	mockAPI := MockableRequest{}
	mockAPI.Method = string(ctx.Method())
	serviceID, apiURL, err := parseAPIUrl(string(ctx.RequestURI()))
	if err != nil {
		return nil, nil, err
	}
	bodyBytes := ctx.Request.Body()
	if len(bodyBytes) > 0 {
		var req map[string]interface{}
		err := json.Unmarshal(bodyBytes, &req)
		if err != nil {
			return nil, nil, err
		}
		mockAPI.RequestPayload = req
	}
	mockAPI.APIURL = *apiURL
	return serviceID, &mockAPI, nil
}

func proxyTheRequest(ctx *fasthttp.RequestCtx, service *core.Service, requestURI string) {
	serviceProxy := service.ReverseProxy
	ctx.Request.SetRequestURI(requestURI)
	if serviceProxy != nil {
		serviceProxy.ServeHTTP(ctx)
	} else {
		handleInternalError(ctx, "Service reverse proxy is not properly initialized")
	}
}

func bigFatHandler(ctx *fasthttp.RequestCtx) {
	//handler for all the registered mocks
	//Core logic to fetch the mocked API and interact with it based on mode
	// Modes are defined as below
	// MOCK (Default mode) - Mock                   -   Maps the incoming request to configured mock api, if found -  returns configured response in the mock
	// Service level Mode
	// SPT                 - ServicePassThrough 	- 	Maps the incoming request to configured mocks service/api, if service found with mode SPT and api mock not found,
	//													then proxies whatver request to actual endpoint/service configured in mock
	// API level mode
	// APT                 - APIPassThrough         -   Maps the incoming request to configured mock service/api, if found both service/api found and api configured with apt,
	//													then passes the request to actual endpoint/service configured in mock
	// TODO - Revisit the below two use cases / modes
	// ValReq - ValidateRequest 	- 	Invokes the actual service/api against the provided request payload instead of configured mock request payload,
	//									ignores configured request payload and validates the response against configured response payload
	// ValResp - ValidateResponse 	-  	Invokes the actual service/api against the configured request payload, ignores incoming
	//                            		request payload  and validates the response against configured response payload
	log.Info(fmt.Sprintf("Resolving the API for path  %s", ctx.RequestURI()))
	serviceID, apiDetails, err := fetchAPIInvocationDetails(ctx)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	log.Info(fmt.Sprintf("Mock request mapped : ServiceID=%v, APIDetails=%v", *serviceID, *apiDetails))
	service, err := core.GetServiceByID(*serviceID)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	verb, err := core.ResolveVerb(apiDetails.Method)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	var reqAsserted map[string]interface{} = nil
	if apiDetails.RequestPayload != nil {
		typedPayload, OK := apiDetails.RequestPayload.(map[string]interface{})
		if !OK {
			handleInternalError(ctx, "RequestPayload should be of json type")
			return
		}
		reqAsserted = typedPayload
	}
	log.Info(fmt.Sprintf("API Resolved, generatin APIID using Url=%v, Verb=%v, Payload=%v", apiDetails.APIURL, verb, reqAsserted))
	apiID, _, err := core.GenerateAPIID(apiDetails.APIURL, verb, reqAsserted)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	api, err := service.GetAPIByID(*apiID)
	//TODO Allow pass through/proxy for :
	// 1.  A non-registered api when service allows pass through (api==nil || err!=nil) or
	// 2.  A registered pass through api (second check)

	//non-registered api
	if err != nil {
		if !service.IsPassThroughAllowed() {
			handleInternalError(ctx, err.Error())
			return
		}
		proxyTheRequest(ctx, service, apiDetails.APIURL)
		return
	}
	if api.IsPassThroughAPI() {
		proxyTheRequest(ctx, service, apiDetails.APIURL)
		return
	}
	writeJSONResponse(ctx, api.APIResponse.ResponsePayload, &api.APIResponse.ResponseCode)
}

// MockableRequest - A valid API request payload that can be registered as mock
type MockableRequest struct {
	APIURL          string      `json:"api_url" validate:"required"`
	Method          string      `json:"method" validate:"required"`
	RequestPayload  interface{} `json:"request_payload"`
	ResponsePayload interface{} `json:"response_payload"`
	ResponseCode    int         `json:"response_code"`
	InvocationMode  *string     `json:"invocation_mode"`
}

func serviceRegistration(ctx *fasthttp.RequestCtx) {
	reqPayload := ctx.Request.Body()
	var req core.Service
	err := json.Unmarshal(reqPayload, &req)
	if err != nil {
		handleInternalError(ctx, "Unable to parse request payload")
		return
	}
	err = validate.Struct(&req)
	if err != nil {
		handleInternalError(ctx, fmt.Sprintf("Invalid Service registration payload :  %v", err.Error()))
		return
	}
	log.Info(fmt.Sprintf("Service Registration request %v", req))
	service, err := core.RegisterService(req.Name, req.Version, req.BaseURL, req.InvocationMode)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	if service.IsPassThroughAllowed() {
		serviceBaseURL := fmt.Sprintf("/%s/{mockedPath:*}", service.ID)
		r.ANY(serviceBaseURL, bigFatHandler)
	}
	writeJSONResponse(ctx, service, nil)
}

// Mandatory parameter in body
// API - Service URI to mock, omit the actual service name just provide the API uri/path that is to be mocked
// Method - Http Method supported by API - For each variation of HTTP method a separate registerService is to be called
// Data - Request Payload (if any)
// Response - Expected response for this method+Data combination
/*
 {
  "api_url":"/v1/helloworld",
  "method":"GET",
  "request_payload":{},
  "response_payload":{},
  "response_code":200,
  "invoke_mode":"MOCK"
 }
*/
func apiRegistration(ctx *fasthttp.RequestCtx) {
	setContentType(ctx, jsontype)
	var req MockableRequest
	err := json.Unmarshal(ctx.Request.Body(), &req)
	if err != nil {
		handleInternalError(ctx, "Unable to parse request payload")
		return
	}
	err = validate.Struct(&req)
	if err != nil {
		handleInternalError(ctx, fmt.Sprintf("Invalid API registration payload :  %v", err.Error()))
		return
	}
	service, err := getServiceFromCtx(ctx)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	verb, err := core.ResolveVerb(req.Method)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	respAsserted, OK := req.ResponsePayload.(map[string]interface{})
	if !OK {
		handleInternalError(ctx, "response_payload should be of json type")
		return
	}
	mockedResp := core.MockedResponse{ResponseCode: req.ResponseCode, ResponsePayload: respAsserted}
	reqAsserted, OK := req.RequestPayload.(map[string]interface{})
	if !OK {
		handleInternalError(ctx, "request_payload should be of json type")
		return
	}
	api, err := service.RegisterAPI(req.APIURL, verb, reqAsserted, &mockedResp, req.InvocationMode)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	log.Info(fmt.Sprintf("API configured as mock = %v %v", req.Method, api.SelfURL))

	// This is the first API registered for this service
	// hence we beed to register this route
	if service.RoutesRegistered() == 1 && !service.IsPassThroughAllowed() {
		serviceBaseURL := fmt.Sprintf("/%s/{mockedPath:*}", service.ID)
		r.ANY(serviceBaseURL, bigFatHandler)
	}
	writeJSONResponse(ctx, api, nil)
}

func getServiceFromCtx(ctx *fasthttp.RequestCtx) (*core.Service, error) {
	serviceID := ctx.UserValue(SERVICEID.String())
	return core.GetServiceByID(fmt.Sprintf("%v", serviceID))
}
func getService(ctx *fasthttp.RequestCtx) {
	service, err := getServiceFromCtx(ctx)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	writeJSONResponse(ctx, service, nil)
}

func getAllServices(ctx *fasthttp.RequestCtx) {
	registeredServices := core.GetRegisteredServices()
	writeJSONResponse(ctx, registeredServices, nil)
}

func getAPIFromCtx(ctx *fasthttp.RequestCtx) (*core.API, error) {
	service, err := getServiceFromCtx(ctx)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return nil, err
	}
	apiID := ctx.UserValue(APIID.String())
	return service.GetAPIByID(fmt.Sprintf("%v", apiID))
}

func getAllAPIs(ctx *fasthttp.RequestCtx) {
	service, err := getServiceFromCtx(ctx)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	writeJSONResponse(ctx, service.GetRegisteredAPIs(), nil)
}

func getAPI(ctx *fasthttp.RequestCtx) {
	api, err := getAPIFromCtx(ctx)
	if err != nil {
		handleInternalError(ctx, err.Error())
		return
	}
	writeJSONResponse(ctx, api, nil)
}

func main() {
	initLogger()
	r.Mutable(true)
	r.GET("/v1", defaultHandler)
	r.GET("/v1/health", defaultHandler)
	r.GET("/v1/info", infoHandler)
	// Core APIs
	// Resource - Service
	r.POST("/v1/service/register", serviceRegistration)
	r.GET("/v1/service/{serviceID}", getService)
	r.GET("/v1/service", getAllServices)

	// Resource - API a.k.a Mocked API
	r.POST("/v1/service/{serviceID}/api/register", apiRegistration)
	r.GET("/v1/service/{serviceID}/api", getAllAPIs)
	r.GET("/v1/services/{serviceID}/api/{apiID}", getAPI)

	log.Info("Server Started, listening on port 8080")
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}
