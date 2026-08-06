*** Settings ***
Documentation     Level ● — you design the cases.
...
...               No TODO list this time. Below are the business rules the
...               service is supposed to enforce. Which of them are worth an
...               acceptance test, what each test asserts, and how many tests it
...               takes are your decisions.
...
...               Constraints:
...                 - black box only. No Go type names, no internal knowledge.
...                 - every test must be able to fail. A test asserting
...                   something that is true no matter what the service does is
...                   worse than no test, because it buys false confidence.
...                 - each test must pass on a fresh service and pass again
...                   immediately after, with no manual cleanup.
...
...               The rules:
...
...               1. An order must have at least one line.
...               2. All lines in one order share a currency.
...               3. The same product cannot appear twice on one order.
...               4. Only a pending order may be modified.
...               5. Transitions run pending → confirmed → shipped → delivered,
...                  and no step may be skipped.
...               6. A delivered order cannot be cancelled. A pending or
...                  confirmed one can.
...               7. Quantities are positive whole numbers.
...               8. Money is never negative, and a total is the sum of its
...                  lines to the cent.
...
...               Before writing anything, answer these three in writing:
...
...               - Which of these eight are already covered by the Go tests in
...                 test/e2e, and does covering them again here buy anything?
...                 "It is tested at a different level" is a real answer, but so
...                 is "this is duplication I will have to maintain twice".
...
...               - Which are not observable over HTTP at all? Each one is
...                 either a rule that needs no API surface, or a gap in the
...                 API. Say which.
...
...               - Which would still pass if the service silently stopped
...                 persisting orders? Those tests are weaker than they look.
Resource          resources/orders.resource
Suite Setup       Connect To Order Service


*** Test Cases ***
Placeholder
    [Documentation]    Delete this and write your own suite.
    ...
    ...                A note on scope: it is tempting to write one test per
    ...                rule. Resist it. Some rules deserve several tests, some
    ...                deserve none at this level, and mechanically mapping
    ...                rules to tests is how suites become unmaintainable.
    Fail    Replace this suite with the acceptance tests you designed
