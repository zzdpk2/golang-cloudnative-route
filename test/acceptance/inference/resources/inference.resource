*** Settings ***
Documentation     Shared black-box keywords for the inference Router.
Library           RequestsLibrary
Library           Collections


*** Variables ***
${BASE_URL}        %{INFERENCE_BASE_URL=http://localhost:8080}
${ALIAS}           inference-router


*** Keywords ***
Connect To Inference Router
    Fail    TODO(exercise): create the bounded HTTP session used by black-box cases

A Chat Request
    [Arguments]    ${model}=tiny-llm    ${prompt}=hello router
    Fail    TODO(exercise): build a request without importing Go domain types

Route Chat
    [Arguments]    ${payload}    ${prefix}=${EMPTY}
    Fail    TODO(exercise): call only the public completion endpoint and return its response

The Selected Endpoint Should Be
    [Arguments]    ${response}    ${endpoint}
    Fail    TODO(exercise): assert selection through stable public response evidence

Read Last Decision
    Fail    TODO(exercise): retrieve and return the public decision trace
