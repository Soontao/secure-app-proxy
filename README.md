# Secure App Proxy

> secure your app, with most simplest setup

## Configuration

> by system environment only

- [x] UPSTREAM
- [x] APPEND_FORWARD_HEADERS - `APPEND_FORWARD_HEADERS=false`
- [x] APPEND_CUSTOM_HEADERS - `APPEND_CUSTOM_HEADERS_X-A=cccc`
- [x] DELETE_SOURCE_HEADERS - `DELETE_SOURCE_HEADERS_authorization=true`
- [x] JWT_SECRET
  - [x] forward `X-Auth-Subject` to upstream
- [x] RATE_LIMIT - [document](https://github.com/ulule/limiter)
- [ ] FORM_LOGIN
  - [ ] STORAGE
- [ ] OIDC_LOGIN
  - [ ] ODIC_ISSUER
  - [ ] ODIC_CLIENT_ID
  - [ ] ODIC_CLIENT_SECRET