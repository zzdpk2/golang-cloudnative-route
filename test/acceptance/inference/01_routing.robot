*** Settings ***
Documentation     TODO(exercise): verify routing only through public behavior.
Resource          resources/inference.resource
Suite Setup       Connect To Inference Router


*** Test Cases ***
A Cached Prefix Competes With Load
    [Tags]    routing    prefix    exercise
    Fail    TODO(exercise): prove affinity yields when the cached Endpoint is saturated

Tenant Fairness Prevents Starvation
    [Tags]    routing    fairness    concurrency    exercise
    Fail    TODO(exercise): hold capacity and prove another tenant makes progress

The Decision Trace Explains Every Candidate
    [Tags]    routing    explainability    exercise
    Fail    TODO(exercise): assert total and per-plugin scores without Go types
