nginx-smtp-auth
===============

A simple Go server to act as a proof-of-concept for how to authenticate to resources protected by NGINX's auth module via an SMTP server.

> NOTE: no cookies/sessions baked in yet so every request necessarily authenticates with the smtp server which isn't cool.

Checkout the example config in the [example](./example/) directory.

## References

- https://docs.nginx.com/nginx/admin-guide/security-controls/configuring-subrequest-authentication/
- https://developer.okta.com/blog/2018/08/28/nginx-auth-request
