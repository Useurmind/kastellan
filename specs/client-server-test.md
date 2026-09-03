# Client Server communication test

We want to implement an integration test suite that proves that the client(agent) and server (operator) communication works.

The following steps need to be performed:

- complete the grpc client implementation for the agent
- complete the grpc server implementation for the operator
- implement test cases that connect these two and check different communication flows work

## Complete client implementation

The client implementation in pkg/agentprotocol/client must be completely implemented.
It should allow 

- connecting to a grpc operator server
- send the agent messages
- receive operator messages

## Complete server implementation

The server implementation in pkg/agentprotocol/server must be completely implemented.
It should allow 

- serving an endpoint for connection requests
- send the operator messages
- receive agent messages

## First test cases

The first test case should be 

- client connects to server and sends client hello and enroll message, check that server receives the correct messages
