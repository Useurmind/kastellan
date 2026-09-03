# Implementation of agent enrollment

I want to start to implement the enrollment of an agent in the operator.

For this we need several components:

- Agent run command with enrollment token
- Operator server that listens to enrollment requests

## Agent run command

The agent run command should take an enrollment token and start connecting to the server.
The command should not fail, the agent should continue to try connecting to the operator if the operator is not available.
It should send an enrollment request and expect a confirmation.
After that it should just wait.

## Operator Server

The Operator Server should listen for enrollment requests and if one is received it should update the status of the external host that the agent is managing.

The status should look like this:

```yaml
status:
  active: <true or false, depending on whether the agent is connected>
  capabilities: [ <list of agent capabilities> ]
  agent:
    id: <agent id>
    version: <agent version>
  host:
    name: <agent host name>
    hostname: <agent hostname>
    ip: <agent hostIp>
  runtime:
    name: <agent runtime name, e.g. podman>
    version: <agent runtime version, e.g. 5.9.0>
```

if the agent disconnects for some reason the active flag in the status should be set to `false`, on reconnect it should be set to `true`.