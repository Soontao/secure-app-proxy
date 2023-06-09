# Secure App Proxy

> secure your app, with most simplest setup

[![Go Build](https://github.com/Soontao/secure-app-proxy/actions/workflows/go.yml/badge.svg)](https://github.com/Soontao/secure-app-proxy/actions/workflows/go.yml)
[![codecov](https://codecov.io/github/Soontao/secure-app-proxy/branch/main/graph/badge.svg?token=F24Lf6HEhp)](https://app.codecov.io/github/Soontao/secure-app-proxy)

## Configuration

> by system environment only

- [x] UPSTREAM
- [x] header modifications
  - [x] modify out request headers
    - [x] APPEND_FORWARD_HEADERS - `APPEND_FORWARD_HEADERS=false`
    - [x] APPEND_REQ_HEADERS - `APPEND_REQ_HEADERS_X-A=cccc`
    - [x] DELETE_REQ_HEADERS - `DELETE_REQ_HEADERS_authorization=true`
  - [x] modify out response headers
    - [x] APPEND_RES_HEADERS
    - [x] DELETE_RES_HEADERS
- [x] JWT_SECRET
  - [x] forward `X-User-Subject` to upstream
- [x] RATE_LIMIT - [document](https://github.com/ulule/limiter)
- [ ] FORM_LOGIN
  - [ ] STORAGE
- [x] odic integration
  - [x] ODIC_ISSUER
  - [x] ODIC_CLIENT_ID
  - [x] ODIC_CLIENT_SECRET
  - [x] ODIC_CALLBACK_URL
  - [x] ODIC_SESSION_SECRET
  - [ ] logout