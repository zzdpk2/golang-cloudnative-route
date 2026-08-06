# Inference Routing Context

This context receives inference requests and selects suitable model-serving
endpoints while protecting latency, capacity, and model requirements.

## Language

**Inference Request**:
A request to generate model output, including the requested Model and routing
signals such as a reusable prompt prefix.
_Avoid_: Job, task, query

**Model**:
The named generative model requested by an Inference Request.
_Avoid_: Image, deployment, server

**Model Server**:
A running process that loads a Model and produces inference output.
_Avoid_: Worker, Endpoint, Pod

**Endpoint**:
A routable address and current serving state for one Model Server port.
_Avoid_: Pod, Model Server, instance

**Inference Pool**:
A group of Endpoints with the same model-serving and compute configuration that
can receive the same class of Inference Requests.
_Avoid_: Service, cluster, fleet

**Router**:
The inference entry point consisting of a Proxy and an Endpoint Picker.
_Avoid_: Load balancer, gateway

**Proxy**:
The part of the Router that receives and forwards an Inference Request.
_Avoid_: Endpoint Picker, scheduler

**Endpoint Picker (EPP)**:
The part of the Router that chooses an Endpoint using request and Endpoint
state.
_Avoid_: Kubernetes scheduler, Proxy

**Request Handler**:
The EPP component that runs protocol parsing, request-scoped data production,
admission, and response lifecycle processing around scheduling.
_Avoid_: Proxy, HTTP middleware

**Data Producer**:
An EPP plugin that derives request-scoped state used by admission or scheduling.
_Avoid_: Data Layer, parser

**Data Layer**:
The EPP subsystem that maintains shared Endpoint state from sources through
extractors into attributes.
_Avoid_: request body, Registry API

**Flow Control**:
The EPP subsystem that buffers and dispatches admitted requests according to
priority, fairness, ordering, and pool saturation.
_Avoid_: Endpoint picking, retry

**Scheduling Profile**:
An ordered Filter, Score, and Pick policy used by the Endpoint Picker.
_Avoid_: Algorithm, pipeline

**Admission**:
The decision that an Inference Request may consume capacity now.
_Avoid_: Authentication, scheduling

**Prefix Cache Affinity**:
The preference for an Endpoint that can reuse cached work for a matching prompt
prefix.
_Avoid_: Session stickiness, response cache

**Prefill**:
The compute-heavy processing of input tokens that creates reusable KV Cache
state.
_Avoid_: Warm-up, parsing

**Decode**:
The iterative generation of output tokens from prepared KV Cache state.
_Avoid_: Response formatting, detokenization

**Disaggregated Serving**:
Inference in which Prefill and Decode are handled by separately selected Model
Servers.
_Avoid_: Model sharding, load balancing
