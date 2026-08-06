*** Settings ***
Documentation     Shared keywords for the order service acceptance suite.
...
...               A resource file is Robot's unit of reuse. Keywords defined
...               here are available to any suite that imports it.
...
...               Suites 01 to 04 use RequestsLibrary directly. From 05 onward,
...               move proven protocol duplication here. Every keyword below is
...               learner-owned and intentionally starts red.
Library           RequestsLibrary
Library           Collections


*** Variables ***
${BASE_URL}       http://localhost:8080
${ALIAS}          orderd


*** Keywords ***
Connect To Order Service
    [Documentation]    Opens the HTTP session every other keyword reuses.
    ...                Call this from Suite Setup, not from each test.
    Fail    TODO(exercise): create and name the reusable HTTP session

A Valid Order Payload
    [Documentation]    Returns a minimal order body that the service accepts.
    ...
    ...                Keep the default here boring and always valid. Tests that
    ...                need a broken variant should start from this and change
    ...                one thing, so the reader can see exactly what is under
    ...                test.
    [Arguments]    ${customer_id}=cust-1    ${quantity}=${2}
    Fail    TODO(exercise): return the smallest valid payload for the public API

Place An Order
    [Documentation]    POSTs an order and returns the parsed response body.
    ...
    ...                TODO (suite 05): POST ${payload} to /orders, assert the
    ...                status is 201, and return the decoded JSON.
    ...
    ...                Decide whether this keyword should assert the status
    ...                itself. A keyword that always asserts success cannot be
    ...                reused by a test about failure — you may need two.
    [Arguments]    ${payload}
    Fail    TODO: implement Place An Order

Order Should Have Status
    [Documentation]    Fetches an order and checks its status field.
    ...
    ...                TODO (suite 05): GET /orders/${order_id} and assert that
    ...                the status field equals ${expected}.
    [Arguments]    ${order_id}    ${expected}
    Fail    TODO: implement Order Should Have Status
