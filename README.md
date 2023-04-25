# Cortex Gateway

Cortex Gateway is a microservice which strives to help you administrating and operating your [Cortex](https://github.com/cortexproject/cortex) Cluster in multi tenant environments.

## Features

- [x] Authentication of Prometheus & Grafana instances with JSON Web Tokens
- [x] Prometheus & Jager instrumentation, compatible with the rest of the Cortex microservices

#### Authentication Feature

If you run Cortex for multiple tenants you need to identify your tenants every time they send metrics or query them. This is needed to ensure that metrics can be ingested and queried separately from each other. For this purpose the Cortex microservices require you to pass a Header called `X-Scope-OrgID`. Unfortunately the Prometheus Remote write API has no config option to send headers and for Grafana you must provision a datasource to do so. Therefore the Cortex k8s manifests suggest deploying an NGINX server inside of each tenant which acts as reverse proxy. It's sole purpose is proxying the traffic and setting the `X-Scope-OrgID` header for your tenant.

We try to solve this problem by adding a Gateway which can be considered the entry point for all requests towards Cortex (see [Architecture](#architecture)). Prometheus and Grafana can both send a self contained JSON Web Token (JWT) along with each request. This JWT carries a claim which is the tenant's identifier. Once this JWT is validated we'll set the required `X-Scope-OrgID` header and pipe the traffic to the upstream Cortex microservices (distributor / query frontend).

## Configuration

- The default `config.yaml` content is:

```yaml
secret: "xxxx"  
basic:  
  username:      
    password: "password"       
    tenant: "orgid"                          
routes:              
  - path: "/api/v1/push"       
    target: "distributor"
  - path: "/prometheus"
    prefix: true
    target: "query-frontend"
  - path: "/api/v1/alerts"
    target: "ruler"
  - path: "/api/v1/rules"
    prefix: true
    target: "ruler"
  - path: "/alertmanager"
    prefix: true
    target: "alertmanager"    
targets:                
  distributor: "http://127.0.0.1:8004"            
  query-frontend: "http://127.0.0.1:8004"
  alertmanager: "http://127.0.0.1:8004"
  ruler: "http://127.0.0.1:8004"
```

The details:

| Field | Description  |
| --- | --- |
| `secret` | Secret to sign JSON Web Token | 
| `basic` | Basic Auth users |
| `routes` | The proxy routes |
| `targets` | The backend services, like Query Frontend、Distributor|


### Expected JWT payload

The expected Bearer token payload can be found here: [pkg/org/tenant.go#L7](https://github.com/telemetrytower/cortex-gateway/blob/master/pkg/org/tenant.go#L7)

- "tenant_id"
- "aud"
- "version" (must be an integer)

The audience and version claim is currently unused, but might be used in the future (e. g. to invalidate tokens).

### Cortex Geteway Tool

You can use `cortex-gateway-tool` to generate Bearer token.

The comamnd options:

| Flag | Description | Default |
| --- | --- | --- |
| `-auth.jwt-secret` | Secret to sign JSON Web Token | (empty string) |
| `-tenant.id` | The tenant of JSON Web Token | (empty string) |
| `-tenant.aud` | The audience of JSON Web Token | (empty string) |
| `-tenant.version` | The version of JSON Web Token | (0) |

An example:

```bash
$ docker run --entrypoint /bin/cortex-gateway-tool --name cortex-gateway songjiayang/cortex-gateway:v0.1.0 -auth.jwt-secret=test -tenant.id=demo

eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJ0ZW5hbnRfaWQiOiJkZW1vIiwidmVyc2lvbiI6MX0.UnM-5mDK24xDkNPes4VMLzC1xBQ9tx3GKoEjrbdd4beY510t9Oj1w2IIfNO10Fe9QEowFchceJ95X-j30mO1Iw
```