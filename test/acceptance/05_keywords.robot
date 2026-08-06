*** Settings ***
Documentation     Level ◐ — extract the duplication.
...
...               By now you have written the same eight-line order payload
...               five times. This suite has no new test scenarios: the job is
...               to move that repetition into resources/orders.resource and
...               rewrite these tests in terms of it.
...
...               Finish the TODO keywords in resources/orders.resource first,
...               then make the tests below read like business language.
...
...               The target: a reader who knows nothing about HTTP should be
...               able to follow every test in this file. If ${JSON_HEADERS} or
...               a status code still appears here, the abstraction is not done.
Resource          resources/orders.resource
Suite Setup       Connect To Order Service


*** Test Cases ***
A Customer Places An Order
    [Documentation]    TODO: rewrite using A Valid Order Payload and Place An
    ...                Order. Two lines should be enough.
    [Tags]    keywords
    Fail    TODO: implement using the resource keywords

A New Order Is Pending
    [Documentation]    TODO: place an order, then use Order Should Have Status.
    [Tags]    keywords
    Fail    TODO: implement using the resource keywords

A Confirmed Order Reports Its New Status
    [Documentation]    TODO: place, confirm, assert.
    ...
    ...                You will want a Confirm Order keyword. Add it to the
    ...                resource file — that is the point of this suite.
    [Tags]    keywords
    Fail    TODO: implement using the resource keywords

An Order For Many Items Is Priced Correctly
    [Documentation]    TODO: place an order for 10 units and assert the total
    ...                the API reports.
    ...
    ...                If the API does not currently return a total, stop and
    ...                think before adding one. Is a total something the client
    ...                should be told, or something it should compute? Both are
    ...                defensible; the wrong move is adding a field without
    ...                deciding.
    [Tags]    keywords
    Fail    TODO: implement using the resource keywords


*** Keywords ***
# Keywords used by more than one suite belong in resources/orders.resource.
# Keywords only this suite needs can live here. Knowing which is which is the
# skill — the same judgement as deciding whether a Go helper belongs in the
# package under test or in a shared testutil package.
