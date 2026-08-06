*** Settings ***
Documentation     Level 1 - the happy path, written out longhand.
...
...               All tests are yours. They are deliberately repetitive so that
...               extracting keywords in suite 05 is an obvious relief rather
...               than an abstract instruction.
Library           RequestsLibrary
Library           Collections
Suite Setup       Create Session    orderd    ${BASE_URL}    verify=${False}


*** Variables ***
${BASE_URL}        http://localhost:8080
&{JSON_HEADERS}    Content-Type=application/json


*** Test Cases ***
Placing An Order Returns 201 And An Id
    [Documentation]    Build a valid order payload, POST it, and prove the
    ...                response contract without using private Go packages.
    [Tags]    orders
    Fail    TODO(exercise): implement the create-order contract from the public API

A Placed Order Can Be Read Back
    [Documentation]    TODO: place an order, keep its id, then GET
    ...                /orders/<id> and assert 200.
    ...
    ...                The thing to work out here is how a value travels from
    ...                one step to the next. ${...} assignment is scoped to the
    ...                test unless you say otherwise — which is exactly why the
    ...                next test cannot see what this one created.
    [Tags]    orders
    Fail    TODO: implement this test

A New Order Starts As Pending
    [Documentation]    TODO: place an order, read it back, and assert the
    ...                status field is "pending".
    ...
    ...                Note that you are asserting on the *wire* representation,
    ...                not on the Go enum. If the JSON says something other than
    ...                "pending", that is an API design question worth raising,
    ...                not something to work around in the test.
    [Tags]    orders
    Fail    TODO: implement this test

An Order Reports Its Line Count
    [Documentation]    TODO: place an order with two different products and
    ...                assert the response reports two lines.
    [Tags]    orders
    Fail    TODO: implement this test
