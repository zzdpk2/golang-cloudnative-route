*** Settings ***
Documentation     Level 2 - error paths, driven by a table instead of by
...               copy-paste.
...
...               Suite 02 made you write the same six lines per test. This one
...               introduces [Template]: one keyword describing the *shape* of
...               the test, and a table supplying the cases. Adding a seventh
...               case becomes one row, not one more test.
...
...               This is the same idea as a table-driven test in Go. Notice
...               that both languages converge on it, and ask why.
Library           RequestsLibrary
Library           Collections
Suite Setup       Create Session    orderd    ${BASE_URL}    verify=${False}


*** Variables ***
${BASE_URL}        http://localhost:8080
&{JSON_HEADERS}    Content-Type=application/json


*** Test Cases ***    CUSTOMER      QUANTITY    EXPECTED
Rejects Missing Customer
    [Template]    Placing An Order Should Return
    ${EMPTY}      ${2}        400

Rejects Zero Quantity
    [Template]    Placing An Order Should Return
    cust-1        ${0}        400

Rejects Negative Quantity
    [Template]    Placing An Order Should Return
    cust-1        ${-1}       400

Accepts A Valid Order
    [Documentation]    The control case. A table of failures with no success in
    ...                it can pass while the endpoint rejects everything.
    [Template]    Placing An Order Should Return
    cust-1        ${2}        201


*** Test Cases ***
Rejects An Order With No Lines
    [Documentation]    TODO: POST an order whose lines array is empty and
    ...                assert 400.
    ...
    ...                This one is not in the table above because the shape is
    ...                different — the template varies quantity, not the
    ...                presence of lines. Knowing when a case does *not* fit the
    ...                template is part of the skill.
    Fail    TODO: implement this test

Rejects An Unknown Currency
    [Documentation]    TODO: POST a line with currency "XYZ" and assert the
    ...                request is rejected.
    ...
    ...                Then go and check what the service actually returns. If
    ...                it is a 500 rather than a 400, you have found a real bug
    ...                in the error mapping from L4 — fix the service, not the
    ...                test.
    Fail    TODO: implement this test

Rejects Mixed Currencies Within One Order
    [Documentation]    TODO: two lines, one AUD and one USD. The aggregate
    ...                forbids this. Assert the API says so too.
    ...
    ...                This is the interesting one: the rule lives deep in the
    ...                domain, and this test proves it survives all the way out
    ...                to the edge. A rule that is enforced but not observable
    ...                from outside is a rule nobody can rely on.
    Fail    TODO: implement this test


*** Keywords ***
Placing An Order Should Return
    [Documentation]    Implement the template keyword invoked by each table row.
    ...                Vary only the two input fields and assert through HTTP.
    [Arguments]    ${customer_id}    ${quantity}    ${expected_status}
    Fail    TODO(exercise): build the request, call the public API, and assert the expected status
