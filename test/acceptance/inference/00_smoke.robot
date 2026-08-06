*** Settings ***
Documentation     TODO(exercise): prove the packaged Router is ready and one
...               request crosses a real process boundary to a Model Server.
Resource          resources/inference.resource
Suite Setup       Connect To Inference Router


*** Test Cases ***
Router Is Live
    [Tags]    smoke    exercise
    Fail    TODO(exercise): implement liveness and readiness assertions

Chat Completion Crosses The Router
    [Tags]    smoke    exercise
    Fail    TODO(exercise): assert the response contract and selected Endpoint
