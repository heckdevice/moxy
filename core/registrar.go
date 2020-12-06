package core

import (
	"fmt"
)

// Verb - Represent the HTTP Verb enum type
type Verb int

// Service - Baseline struct for a mocker service
type Service struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	registeredAPIs map[string]*API
}

// API - Configure a mock api giving the URL and the http verb supported for the URL
type API struct {
	ID      string `json:"id"`
	URL     string `json:"url"`
	APIVerb Verb   `json:"api_verb"`
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

var (
	registeredServices = make(map[string]*Service)
	verbs              = [...]string{
		"GET",
		"POST",
		"DELETE",
		"PUT",
	}
)

// ServiceRegistration - Service Registration features
type ServiceRegistration interface {
	RegisterService(name string, version string) *Service
	UnregisterService(name, version string)
	GetService(name string, version string) (*Service, error)
}

// APIRegistration - API Registration features
type APIRegistration interface {
	RegisterAPI(url string, httpVerb Verb) (*API, error)
	RegisterAPIWithLatency(url string, httpVerb Verb) (*APIWithLatency, error)
}

// Feature Implementations

//
// Core Features
//

func getServiceKey(name, version string) string {
	return name + "." + version
}

func getAPIKey(url string, verb Verb) string {
	return fmt.Sprintf("%v %v", verb, url)
}

// RegisterService - Registers a specific service name and version
// name, version tuple is considered to uniquely identify a registered service
func RegisterService(name, version string) (*Service, error) {
	if service, OK := registeredServices[getServiceKey(name, version)]; OK {
		return nil, fmt.Errorf("%v already registered", service)
	}
	service := &Service{getServiceKey(name, version), name, version, make(map[string]*API)}
	registeredServices[service.ID] = service
	return service, nil
}

// UnregisterService - Service unregistration feature remove the service in a no-op fashion
func UnregisterService(name, version string) {
	delete(registeredServices, getServiceKey(name, version))
}

// GetService - return the service if found registered else errs
func GetService(name, version string) (*Service, error) {
	if service, OK := registeredServices[getServiceKey(name, version)]; OK {
		return service, nil
	}
	return nil, fmt.Errorf("Service with name=%s, version=%s tuple is not registered", name, version)
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

func (s *Service) String() string {
	return fmt.Sprintf("Service(ID=%v, Name=%v, Version=%v)", s.ID, s.Name, s.Version)
}

func (s *Service) registerAPI(api *API) {
	s.registeredAPIs[api.ID] = api
}

//RegisterAPI - Registers an API for a given service
func (s *Service) RegisterAPI(url string, verb Verb) (*API, error) {
	apiKey := getAPIKey(url, verb)
	if api, OK := s.registeredAPIs[apiKey]; OK {
		return nil, fmt.Errorf("%v already registered with %v", api, s)
	}
	api := &API{apiKey, url, verb}
	s.registerAPI(api)
	return api, nil
}

//RegisterAPIWithLatency - Registers an API for a given service with specified mocked latency
func (s *Service) RegisterAPIWithLatency(url string, verb Verb, latency float32) (*APIWithLatency, error) {
	apiKey := getAPIKey(url, verb)
	if api, OK := s.registeredAPIs[apiKey]; OK {
		return nil, fmt.Errorf("%v already registered with %v", api, s)
	}
	api := &APIWithLatency{API{apiKey, url, verb}, latency}
	s.registerAPI(&(api.API))
	return api, nil
}
