# Richard Merry ATOS Tech Test Submission

## API Documentation

View the auto-generated Swagger documentation [here](http://test).

## Requirements

Create a REST service that can meet the below requirements and publish the source on github.

Return a list of supported symmetric encryption algorithms.

Support the decryption of encrypted text by accepting and returning the response text Base64 encoded:
    - symmetric cipher algorithm name
    - cipher text
    - cipher key

The key and associated algorithm should be reusable for upto 10 minutes with additional cipher texts. Publish the open api spec for the REST service.

## Documentation

I've used [swag](https://github.com/swaggo/swag) for autogeneration of docs from doc tag annotations over the handlers (see `./internal/api/handlers.go`). When making changes to these tags you can reformat them via `make swag-fmt` and regenerate the documentation via `make swag`. The generated docs are in `./docs/swagger`.


## Building

Running `make build` will place the executable in the `./bin` directory.

## Running

Either build and then run using via issuing the command `./bin/server` or simply using the Makefile target: `make run`.



