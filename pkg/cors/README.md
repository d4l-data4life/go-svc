# go-cors

[![Build Status](https://travis-ci.com/gesundheitscloud/go-cors.svg?token=R1P7ZqnP8S5syLsxefpy&branch=master)](https://travis-ci.com/gesundheitscloud/go-cors)
[![codecov](https://codecov.io/gh/gesundheitscloud/go-cors/branch/master/graph/badge.svg?token=R36hAMWaBF)](https://codecov.io/gh/gesundheitscloud/go-cors)

A Go package for injecting Cross-Origin Resource Sharing headers.

## Usage

The package entrypoint is the `Wrap` function.

`Wrap` accepts a `http.Handler` as an argument, and therefore can both apply CORS headers to a single endpoint or to the whole `http.ServeMux`, as needed.

Example of applying CORS to the whole webserver:

```Go
	router := http.NewServeMux()
	router.Handle("/handshake", handler)

	withCORS := cors.Wrap(router,
		cors.WithDomainList(os.Getenv("CORS_ALLOWED_ORIGINS")),
		cors.WithHeaderList("Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token, GC-SDK-Version, X-User-Language"),
	)

	log.Fatal(http.ListenAndServe(":8080", withCORS))
```

Example of applying CORS to a single endpoint:

```Go
	handler1WithCors := cors.Wrap(handler1,
		cors.WithDomains(handler1AllowedDomains),
		cors.WithHeaderList(os.Getenv("CORS_ALLOWED_HEADERS")),
	)

	router := http.NewServeMux()
	router.Handle("/endpoint1", handler1WithCors)
	router.Handle("/endpoint2", handler2)

	log.Fatal(http.ListenAndServe(":8080", router))
```

For the full API, use `godoc`.
