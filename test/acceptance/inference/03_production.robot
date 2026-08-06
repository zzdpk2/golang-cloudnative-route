*** Settings ***
Documentation     TODO(exercise): collect production failure and recovery evidence.
Resource          resources/inference.resource
Suite Setup       Connect To Inference Router


*** Test Cases ***
Bounded Overload Recovers Without Capacity Leak
    [Tags]    production    overload    concurrency    exercise
    Fail    TODO(exercise): saturate admission, assert the hard bound, then prove recovery

Endpoint Loss And Router Eviction Preserve The Declared Availability
    [Tags]    production    kubernetes    disruption    exercise
    Fail    TODO(exercise): remove an Endpoint and evict a Router Pod through public operations

Alerts Fire And Resolve During A Controlled Failure
    [Tags]    production    prometheus    recovery    exercise
    Fail    TODO(exercise): inject one documented failure and assert firing plus recovery
