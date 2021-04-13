# moxy
A simple service proxy layer with downstream mocking capabilities inbuilt

## Docker build

Assumption you have docker installed, these command are tested and verified for Docker engine 20.10.5

- Build Docker image "moxy"

 ```
 docker build --no-cache -t moxy:latest .
 ```
- Run the moxy service, by default it listens on port 8080

 ```
 docker run -dit -p 8080:8080 --name moxy moxy
 ```
- Confirm its running
 
 ```
  docker logs -f moxy
 ```

 if all good it should show the following log lines

 ```
  {"level":"info","msg":"Server Started, listening on port 8080",......}
 ```

## Sample Moxy Flow

Assuming moxy is up and running, listening on port 8080, sample flows showing how to register a google search api as a mock

### Indices

* [APIs](#apis)

  * [Register Google search service](#1-register-google-search-service)
  * [Register Google CustomSearch API](#2-register-google-customsearch-api)
  * [Fetch Service](#3-fetch-service)
  * [Fetch Google CustomSearch registered APIs](#4-fetch-google-customsearch-registered-apis)
  * [Invoke CustomSearch API](#5-invoke-customsearch-api)
 
--------


### APIs

Step 1 - Register a service - Lets name this as Google and version at 1.0

#### 1. Register Google search service


The simple payload for registering a Service


***Endpoint:***

```bash
Method: POST
Type: RAW
URL: http://localhost:8080/v1/service/register
```



***Body:***

```js        
{
	"name":"google",
	"version":"1.0"
}
```



#### 2. Register Google CustomSearch API


Step 2 - Registering an API with a service, here we will use already registered service Google 1.0,
as per Step 1 above, we would get the ServiceID in response to service registration and we will use that as follows to register an API against that service

POST /v1/service/{serviceID}/api/register and provide the API level details in the payload


***Endpoint:***

```bash
Method: POST
Type: RAW
URL: http://localhost:8080/v1/service/google.1.0/api/register
```



***Body:***

```js        
{
	"api_url":"/customsearch/v1?key=INSERT_YOUR_API_KEY&cx=017576662512468239146:omuauf_lfve&q=lecturesnewtestasd4",
	"method":"GET",
	"request_payload":{"me":"Hecky","doby":1978},
	"response_code":200,
	"response_payload":{"msg":"Greetings Hecky who was born on 1978"}
}
```

#### 3. Fetch Service


This API lists all the services registered with Moxy


***Endpoint:***

```bash
Method: GET
Type: RAW
URL: http://localhost:8080/v1/service
```


#### 4. Fetch Google CustomSearch registered APIs


This lists all the APIs registered under a registered service

GET /v1/service/{serviceID}/api


***Endpoint:***

```bash
Method: GET
Type: RAW
URL: http://localhost:8080/v1/service/google.1.0/api
```


#### 5. Invoke CustomSearch API


Last Step - Execute a registered mock API

In the step above we registered a dummy google search api with Signature

GET /google.1.0/{apiURL}

The request has to exactly the same to find a match and return the registered mock

- Verb / HTTP method should be same
- API url including the service ID should match - following the pattern of /{serviceID}/{apiURL}
- Request Payload should be same


***Endpoint:***

```bash
Method: GET
Type: RAW
URL: http://localhost:8080/google.1.0/customsearch/v1
```



***Query params:***

| Key | Value | Description |
| --- | ------|-------------|
| key | INSERT_YOUR_API_KEY |  |
| cx | 017576662512468239146:omuauf_lfve |  |
| q | lecturesnewtestasd4 |  |



***Body:***

```js        
{
            "doby": 1978,
            "me": "Hecky"
        }
```

---
[Back to top](#sample-moxy-flow)
