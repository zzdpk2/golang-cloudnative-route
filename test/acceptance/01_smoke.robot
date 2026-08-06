*** Settings ***
Documentation     Level 0 - build the smallest possible acceptance suite.
...
...               Both cases are learner-owned. Keep the proof narrow: exercise
...               routing and decoding without depending on private Go code.
...
...               Settings    imports and suite-level hooks
...               Variables   suite-scoped values
...               Test Cases  the tests themselves
Library           RequestsLibrary
Suite Setup       Create Session    orderd    ${BASE_URL}    verify=${False}


*** Variables ***
${BASE_URL}        http://localhost:8080
&{JSON_HEADERS}    Content-Type=application/json


*** Test Cases ***
The Service Answers An Unknown Order With 404
    [Documentation]    The cheapest possible proof that the service is up and
    ...                routing. It deliberately asserts a *failure* response:
    ...                a 404 means the router matched and a handler ran, which
    ...                is more than a 200 on a health endpoint would tell you.
    ...
    ...                There is no health endpoint in this Commerce phase;
    ...                readiness is introduced later in the Router journey.
    [Tags]    smoke
    Fail    TODO(exercise): send an unknown-order request and prove the public response is 404

The Service Rejects A Malformed Body
    [Documentation]    Proves the request actually reaches JSON decoding, so
    ...                the route, the middleware chain, and the handler all ran.
    [Tags]    smoke
    Fail    TODO(exercise): send malformed JSON and prove the public response is 400
