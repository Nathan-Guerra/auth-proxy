# Auth Proxy
## Scope

This project will be designed with the intention of smooth the API integration
with my local perfect-pay api. To do so it will receive a request, validate if
it is a api call and, if so, it will relay my request with a bearer token.

### The Bearer Token
To get it, I will send a request to my `/api/auth/login` endpoint with some
variables set in my environment variables file, save the token that i receive
and, based on some time rule use this token until it expires. When it comes to
an end, the proxy should be able to resend the request to the login endpoint
and recreate the token automatically. Thus not being necessary me to do it
manually.

## Constraints
Every argument used in this project should be stored at an env file. This proxy
should be able to try to connect to any API through any endpoint and any host.

The path of the authentication token should be given as a dot path where each
dot is an index/acessor to the api response data. For example, the path for the
JSON below should be `data.token`.
```json
{
    "data": {
        "token": "T0K3N"
    }
}
```
---
The bearer token will be stored in a file for a defined set of time and should
be read until it expires.
