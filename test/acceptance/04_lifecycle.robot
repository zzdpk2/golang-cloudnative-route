*** Settings ***
Documentation     Level 3 - a stateful, multi-step flow.
...
...               Everything so far was one request per test. An order's life is
...               a sequence, and testing a sequence raises a question the
...               earlier suites let you dodge: how much state may leak between
...               tests?
...
...               The strict answer is none. A test that only passes when it
...               runs after another test is not a test, it is a step. But
...               setting up a fresh order for every case costs a request, and
...               there is a real trade-off here. Suite Setup versus Test Setup
...               is that trade-off, made explicit.
...
...               Read how this suite resolves it, then argue with the choice.
Library           RequestsLibrary
Library           Collections
Suite Setup       Create Session    orderd    ${BASE_URL}    verify=${False}
Test Setup        Place A Fresh Order


*** Variables ***
${BASE_URL}        http://localhost:8080
&{JSON_HEADERS}    Content-Type=application/json
${ORDER_ID}        ${EMPTY}


*** Test Cases ***
A Fresh Order Is Pending
    [Documentation]    Prove that Test Setup created a readable pending order.
    [Tags]    lifecycle
    Fail    TODO(exercise): fetch the fresh order and prove its initial state

Confirming Moves The Order To Confirmed
    [Documentation]    TODO: POST /orders/${ORDER_ID}/confirm, then read the
    ...                order back and assert its status is "confirmed".
    [Tags]    lifecycle
    Fail    TODO: implement this test

Confirming Twice Is Rejected
    [Documentation]    TODO: confirm the order, then confirm it again, and
    ...                assert the second attempt fails.
    ...
    ...                Which status code should a repeated confirm return? 409
    ...                Conflict and 422 Unprocessable Entity are both defensible.
    ...                Pick one, make the service agree, and write down why. The
    ...                answer matters less than the fact that it is consistent.
    [Tags]    lifecycle
    Fail    TODO: implement this test

Confirming An Unknown Order Is A 404
    [Documentation]    TODO: confirm an order id that was never created.
    ...
    ...                Careful: this test does not need the order that Test
    ...                Setup created. That is a hint that Test Setup is doing
    ...                work some tests do not want. Is that acceptable, or is it
    ...                a sign this test belongs in a different suite?
    [Tags]    lifecycle
    Fail    TODO: implement this test

A Confirmed Order Cannot Be Modified
    [Documentation]    TODO: confirm the order, then try to change it, and
    ...                assert the change is refused.
    ...
    ...                You may find the API has no endpoint for this yet. If so,
    ...                that is the finding: the aggregate enforces a rule that
    ...                the API gives no way to exercise. Write down what the
    ...                endpoint should look like — you will need it in L15.
    [Tags]    lifecycle
    Fail    TODO: implement this test


*** Keywords ***
Place A Fresh Order
    [Documentation]    Creates a new order before every test and stores its id
    ...                in a suite-scoped variable.
    ...
    ...                Set Suite Variable is what makes ${ORDER_ID} visible to
    ...                the test that runs next. Look up Set Test Variable and
    ...                work out why it would not work here.
    Fail    TODO(exercise): create a fresh order and expose its id to the current test
