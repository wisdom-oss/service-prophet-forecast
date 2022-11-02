# WISdoM OSS - Prophet Forecast Microservice

The prophet forecast microservice uses the prophet forecasting library published by Facebook to forecast
the water usages. Furthermore, the microservice will use the current population forecasts to calculate a
per-person water usage.

Information on how to access this service are available in the openapi.yaml

## Configuration

The microservice template is configurable via the following environment variables:
- `CONFIG_LOGGING_LEVEL` &#8594; Set the logging verbosity [optional, default `INFO`]
- `CONFIG_API_GATEWAY_HOST` &#8594; Set the host on which the API Gateway runs on **[required]**
- `CONFIG_API_GATEWAY_ADMIN_PORT` &#8594; Set the port on which the API Gateway listens on **[required]**
- `CONFIG_API_GATEWAY_SERVICE_PATH` &#8594; Set the path under which the service shall be reachable. _Do not prepend the path with `/api`. Only set the last part of the desired path_ **[required]**
- `CONFIG_HTTP_LISTEN_PORT` &#8594; The port on which the built-in webserver will listen on [optional, default `8000`]
- `CONFIG_SCOPE_FILE_PATH` &#8594; The location where the scope definition file is stored [optional, default `/microservice/res/scope.json]

