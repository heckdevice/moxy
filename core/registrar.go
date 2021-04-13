package core

import (
	"crypto/sha256"
	"encoding/hex"
	json "encoding/json"
	"fmt"
	proxy "github.com/yeqown/fasthttp-reverse-proxy/v2"
)

// Verb - Represent the HTTP Verb enum type
type Verb int

// Payload - Represent generic HTTP JSON payload
type Payload map[string]interface{}

// PayloadFeatures - Represent generic Payload functions
type PayloadFeatures interface {
	getSeed() ([]byte, error)
	String() string
}

// MockedResponse - Represents mocked response for an API
type MockedResponse struct {
	ResponseCode    int
	ResponsePayload Payload
}

// Service - Baseline struct for a mocker service
type Service struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name" validate:"required"`
	Version        string `json:"version" validate:"required"`
	registeredAPIs map[string]*API
	BaseURL        *string `json:"base_url,omitempty"`
	InvocationMode *string `json:"invocation_mode,omitempty" default:"mock"`
	ReverseProxy   *proxy.ReverseProxy
}

// API - Configure a mock api giving the URL and the http verb supported for the URL
type API struct {
	ID             string  `json:"id"`
	URL            string  `json:"url" validate:"required"`
	APIVerb        Verb    `json:"api_verb" validate:"required"`
	APIPayload     Payload `json:"api_payload"`
	ServiceID      string  `json:"service_id"`
	idSeeds        *IDSeeds
	APIResponse    *MockedResponse `json:"api_response"`
	SelfURL        string          `json:"self_url"`
	InvocationMode *string         `json:"invocation_mode" default:"mock"`
}

// IDSeeds - Various elements that seed the API Id generation hash
// TODO - Placeholder for future optimisations of API mock reverse lookups
type IDSeeds struct {
	CoreSeed    []byte `json:"core_seed"`
	PayloadSeed []byte `json:"payload_seed"`
}

// APIWithLatency - Configure a mock api like base API struct but with simulated Latency
type APIWithLatency struct {
	API
	Latency float32 `json:"latency"`
}

const (
	// GET - HTTP Get Verb enum int value = 0
	GET Verb = iota
	// POST - HTTP Post Verb enum int value = 1
	POST
	// DELETE - HTTP Delete Verb enum int value = 2
	DELETE
	// PUT - HTTP Put Verb enum int value = 3
	PUT
)

// String - string valueof the Verb enum
func (v Verb) String() string { return verbs[v] }

// ResolveVerb - Resolves HTTP method string to moxy Verb
func ResolveVerb(httpMethod string) (Verb, error) {
	switch httpMethod {
	case "GET":
		return GET, nil
	case "POST":
		return POST, nil
	case "DELETE":
		return DELETE, nil
	case "PUT":
		return PUT, nil
	default:
		return -1, fmt.Errorf("HTTP Method %v not supported", httpMethod)
	}
}

var (
	registeredServices = make(map[string]*Service)
	verbs              = [...]string{
		"GET",
		"POST",
		"DELETE",
		"PUT",
	}
	defaultMode = "mock"
	apiModes    = map[string]string{
		"mock":     "mock",
		"apt":      "apt",
		"validate": "validate",
	}
	serviceModes = map[string]string{
		"mock": "mock",
		"spt":  "spt",
	}
)

//GetRegisteredServices - Gets all the registered services list
func GetRegisteredServices() map[string]*Service {
	return registeredServices
}

// ServiceRegistration - Service Registration features
type ServiceRegistration interface {
	RegisterService(name, version, baseURL *string) (*Service, error)
	UnregisterService(name, version string)
	GetService(name string, version string) (*Service, error)
	RoutesRegistered() int
}

// APIRegistration - API Registration features
type APIRegistration interface {
	RegisterAPI(url string, httpVerb Verb, payload Payload, resp *MockedResponse, invocationMode *string) (*API, error)
	RegisterAPIWithLatency(url string, httpVerb Verb, payload Payload, resp *MockedResponse, invocationMode *string) (*APIWithLatency, error)
}

// Feature Implementations

//
// Core Features
//

// InterfaceToJSONString - converts to a JSON string
func InterfaceToJSONString(data interface{}) (*string, error) {
	dataBytes, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return nil, err
	}
	dataJSON := string(dataBytes)
	return &dataJSON, nil
}

func getServiceKey(name, version string) string {
	return name + "." + version
}

func (s *Service) validateServiceMode() (*string, error) {
	if s.InvocationMode == nil {
		defMode, _ := serviceModes[defaultMode]
		return &defMode, nil
	}
	if *s.InvocationMode != defaultMode && s.BaseURL == nil {
		return nil, fmt.Errorf(fmt.Sprintf("Mode %s is not supported if services' base_url is not set", *s.InvocationMode))
	}
	if supporteMode, OK := serviceModes[*s.InvocationMode]; OK {
		return &supporteMode, nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("Invalid mode %s", *s.InvocationMode))
}

// RegisterService - Registers a specific service name and version
// name, version is considered to uniquely identify a registered service
func RegisterService(name, version string, baseURL *string, mode *string) (*Service, error) {
	serviceKey := getServiceKey(name, version)
	if service, OK := registeredServices[serviceKey]; OK {
		return nil, fmt.Errorf("%v already registered", service)
	}
	service := &Service{serviceKey, name, version, make(map[string]*API), baseURL, mode, nil}
	serviceMode, err := service.validateServiceMode()
	if err != nil {
		return nil, err
	}
	service.InvocationMode = serviceMode
	if *service.InvocationMode == "spt" {
		service.ReverseProxy = proxy.NewReverseProxy(*service.BaseURL)
	}
	registeredServices[service.ID] = service
	return service, nil
}

// UnregisterService - Service unregistration feature remove the service in a no-op fashion
func UnregisterService(name, version string) {
	delete(registeredServices, getServiceKey(name, version))
}

// GetService - Lookup for registered service by name,version; errs if not found
func GetService(name, version string) (*Service, error) {
	if service, OK := registeredServices[getServiceKey(name, version)]; OK {
		return service, nil
	}
	return nil, fmt.Errorf("Service with name=%s, version=%s tuple is not registered", name, version)
}

// GetServiceByID - Lookup for registerd service by id, errs if not found
func GetServiceByID(serviceID string) (*Service, error) {
	if service, OK := registeredServices[serviceID]; OK {
		return service, nil
	}
	return nil, fmt.Errorf("Service with id=%v is not registered", serviceID)
}

//
// API Features supported by Service
//

func (api *API) String() string {
	return fmt.Sprintf("API(ID=%v, URL=%v, APIVerb=%v)", api.ID, api.URL, api.APIVerb)
}

func (api *APIWithLatency) String() string {
	return fmt.Sprintf("APIWithLatency(ID=%v, URL=%v, APIVerb=%v, Latency=%v)", api.ID, api.URL, api.APIVerb, api.Latency)
}

// IsPassThroughAPI - Checks if API is configured for pass-through / proxy
func (api *API) IsPassThroughAPI() bool {
	return *api.InvocationMode == "apt"
}

// IsPassThroughAPI - Checks if API is configured for pass-through / proxy
func (api *APIWithLatency) IsPassThroughAPI() bool {
	return *api.InvocationMode == "apt"
}

func (s *Service) String() string {
	return fmt.Sprintf("Service(ID=%v, Name=%v, Version=%v)", s.ID, s.Name, s.Version)
}

func (p Payload) String() string {
	val, err := InterfaceToJSONString(p)
	if err != nil {
		return err.Error()
	}
	return *val
}

func (p Payload) getSeed() ([]byte, error) {
	var v interface{}
	jsonDoc := p.String()
	err := json.Unmarshal([]byte(jsonDoc), &v)
	if err != nil {
		return nil, err
	}
	payloadSeed, _ := json.Marshal(v)
	return payloadSeed, nil
}

//GetCoreSeed - Gets the Core seed made up of url,verb part in API
func GetCoreSeed(url string, verb Verb) []byte {
	return []byte(fmt.Sprintf("%s %s", verb, url))
}

//GetPayloadSeed - Gets the payload seed made up of payload part in API
func GetPayloadSeed(payload Payload) ([]byte, error) {
	payloadSeed, err := payload.getSeed()
	if err != nil {
		return nil, fmt.Errorf("Error getting payload seed :: %v", err.Error())
	}
	return payloadSeed, nil
}

// GenerateAPIID - Generates the unique API ID hased using - API method, API Url and Request Payload (if any)
// returns APIID, api id generation seeds / hashes
func GenerateAPIID(url string, verb Verb, payload Payload) (*string, *IDSeeds, error) {
	var idSeeds IDSeeds
	apiIDSeed := GetCoreSeed(url, verb)
	idSeeds.CoreSeed = apiIDSeed
	if payload != nil {
		payloadSeed, err := GetPayloadSeed(payload)
		if err != nil {
			return nil, nil, err
		}
		idSeeds.PayloadSeed = payloadSeed
		apiIDSeed = append(apiIDSeed, payloadSeed...)
	}
	sum := sha256.Sum256(apiIDSeed)
	apiID := hex.EncodeToString(sum[0:])
	return &apiID, &idSeeds, nil
}

func (s *Service) registerAPI(api *API) {
	s.registeredAPIs[api.ID] = api
}

// RoutesRegistered - Returns numbers of APIs registered for the service
func (s *Service) RoutesRegistered() int {
	return len(s.registeredAPIs)
}

// validateAPIMode - Validates against supported API modes
func (s *Service) validateAPIMode(apiMode *string) (*string, error) {
	if apiMode == nil {
		defMode, _ := serviceModes[defaultMode]
		return &defMode, nil
	}
	if *s.InvocationMode == defaultMode && *apiMode != defaultMode && s.BaseURL == nil {
		return nil, fmt.Errorf(fmt.Sprintf("API mode %s is not supported if services' base_url is not set", *apiMode))
	}
	if supporteMode, OK := apiModes[*apiMode]; OK {
		return &supporteMode, nil
	}
	return nil, fmt.Errorf(fmt.Sprintf("API mode %s not supported", *apiMode))
}

//RegisterAPI - Registers an API for a given service
func (s *Service) RegisterAPI(url string, verb Verb, payload Payload, response *MockedResponse, mode *string) (*API, error) {
	apiKey, apiSeeds, err := GenerateAPIID(url, verb, payload)
	if err != nil {
		return nil, fmt.Errorf("Error generating API ID :: %v", err.Error())
	}
	if api, OK := s.registeredAPIs[*apiKey]; OK {
		return nil, fmt.Errorf("%v already registered with %v", api, s)
	}
	selfURL := fmt.Sprintf("/%s%s", s.ID, url)
	api := &API{*apiKey, url, verb, payload, s.ID, apiSeeds, response, selfURL, mode}
	apiMode, err := s.validateAPIMode(api.InvocationMode)
	if err != nil {
		return nil, err
	}
	api.InvocationMode = apiMode
	if *api.InvocationMode == "apt" && s.ReverseProxy == nil {
		s.ReverseProxy = proxy.NewReverseProxy(*s.BaseURL)
	}
	s.registerAPI(api)
	return api, nil
}

//RegisterAPIWithLatency - Registers an API for a given service with specified mocked latency
func (s *Service) RegisterAPIWithLatency(url string, verb Verb, payload Payload, latency float32, response *MockedResponse, mode *string) (*APIWithLatency, error) {
	apiKey, apiSeeds, err := GenerateAPIID(url, verb, payload)
	if err != nil {
		return nil, fmt.Errorf("Error generating API ID :: %v", err.Error())
	}
	if api, OK := s.registeredAPIs[*apiKey]; OK {
		return nil, fmt.Errorf("%v already registered with %v", api, s)
	}
	selfURL := fmt.Sprintf("/%s%s", s.ID, url)
	api := &APIWithLatency{API{*apiKey, url, verb, payload, s.ID, apiSeeds, response, selfURL, mode}, latency}
	apiMode, err := s.validateAPIMode(api.InvocationMode)
	if err != nil {
		return nil, err
	}
	api.InvocationMode = apiMode
	if *api.InvocationMode == "apt" && s.ReverseProxy == nil {
		s.ReverseProxy = proxy.NewReverseProxy(*s.BaseURL)
	}
	s.registerAPI(&(api.API))
	return api, nil
}

// GetAPIByID - Fetches registered api by api id, errs if not found
func (s *Service) GetAPIByID(apiID string) (*API, error) {
	if api, OK := s.registeredAPIs[apiID]; OK {
		return api, nil
	}
	return nil, fmt.Errorf("API, Service (%v, %v) combination not found", apiID, s.ID)
}

// GetAPI - Fetches registered api by api id, errs if not found
func (s *Service) GetAPI(url string, verb Verb, payload Payload) (*API, error) {
	apiID, _, err := GenerateAPIID(url, verb, payload)
	if err != nil {
		return nil, err
	}
	if api, OK := s.registeredAPIs[*apiID]; OK {
		return api, nil
	}
	return nil, fmt.Errorf("API, Service (%v, %v) combination not found", apiID, s.ID)
}

// GetRegisteredAPIs - Returns a map of all the registered APIs in the service
func (s *Service) GetRegisteredAPIs() map[string]*API {
	return s.registeredAPIs
}

// IsPassThroughAllowed - Checks if the services is configured for pass-through / proxy mode
func (s *Service) IsPassThroughAllowed() bool {
	return *s.InvocationMode == "spt"
}
