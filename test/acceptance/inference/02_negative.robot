*** Settings ***
Documentation     TODO(exercise): verify stable external failures and capacity cleanup.
Resource          resources/inference.resource
Suite Setup       Connect To Inference Router


*** Test Cases ***
Invalid Requests Have Stable Error Codes
    [Tags]    negative    contract    exercise
    Fail    TODO(exercise): cover malformed JSON, unsupported Model, and bad priority

Client Cancellation Releases Capacity
    [Tags]    negative    cancellation    concurrency    exercise
    Fail    TODO(exercise): cancel a slow stream and prove the next request is admitted

Mid Stream Failure Is Not Retried
    [Tags]    negative    streaming    retry    exercise
    Fail    TODO(exercise): prove no duplicate inference after downstream bytes were written
