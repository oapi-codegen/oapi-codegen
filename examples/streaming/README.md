OpenAPI Code Generation Example - Streaming
-------------------------------------------

This directory contains an example server using our code generator which implements
a simple streaming API with a single endpoint.

This is the structure:
- `sse.yaml`: Contains the OpenAPI 3.0 specification
- `stdhttp/`: Contains the written and generated code for the server using the standard http package
- `client/`: Contains a client which reads the server stream and prints out the messages

You can run both together to demonstrate the end-to-end behavior from
both client and server side. Run these commands in parallel:

    go run ./stdhttp
    go run ./client

