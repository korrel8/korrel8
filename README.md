# Overview

**⚠ Warning: Experimental ⚠** This code may change or vanish. It may not work. It may not even make sense.\
[API documentation is available at pkg.go.dev](https://pkg.go.dev/github.com/korrel8/korrel8/pkg/korrel8)

This project is building a *correlation engine*.

Given an initial staring context, korrel8 can find related data from multiple data stores in multiple domains.
Korrel8 works by navigating between related *objects*, using a set of *rules* that define how objects are related.
Given a *start* object (e.g. a cluster Alert) and a *goal* (e.g. "find related logs") the engine searches 
for goal objects that are related to the start object by rules.

The set of rules captures expert knowledge about troubleshooting in an executable form.
The engine aims to provide common rule-base that can be re-used in many settings:
as a service, embedded in graphical consoles or command line tools, or in offline data-processing systems.

The goals of this project include:

- Encode domain knowledge from SREs and other experts as re-usable rules.
- Automate navigation from symptoms to data that helps diagnose causes.
- Reduce multiple-step manual procedures to fewer clicks or references.
- Help tools that gather and analyze diagnostic data to focus on relevant information.

## Types of signals

A Kubernetes cluster generates many types of *observable signal*, including:

| Signal Type       | Description                                                             |
|-------------------|-------------------------------------------------------------------------|
| Metrics           | Counts and measurements of system behaviour.                            |
| Alerts            | Rules that fire when metrics cross important thresholds.                |
| Logs              | Application, infrastructure and audit logs from Pods and cluster nodes. |
| Kubernetes Events | Describe significant events in a cluster.                               |
| Traces            | Nested execution spans describing distributed requests.                 |
| Network Events    | TCP and IP level network information.                                   |

A cluster also contains objects that are not usually considered "signals",
but which can be correlated with signals and other objects:

| Object Type   | Description                                    |
|---------------|------------------------------------------------|
| k8s resources | Spec and status information.                   |
| Run books     | Problem solving guides associated with Alerts. |
| k8s probes    | Information about resource state.              |
| Operators     | Operators control other resources.             |


# Implentation Concepts

The following concepts are represented by interfaces in the korrel8 package.
Support for a new domain implements these interfaces:

**Domain** \
A family of signals or objects with common storage and representation.
Examples: resource, alert, metric, log, trace

**Store** \
A source of signal data from some Domain.
Examples: Loki, Prometheus, Kubernetes API server.

**Reference**  \
A relative URI reference that selects a set of signals from a store.
For stores with HTTP REST APIs, the URI reference can used directly with the store's base URL.
If non-REST stores are added in future, the URI reference format is flexible enough to encode other queries.

**Class**  \
A subset of signals in a Domain with a common schema (field definitions).
Examples: Pod (k8s), Event(k8s), `KubeContainerWaiting` (alert), `log_logged_bytes_total` (metric)

**Object** \
An instance of a signal or other correlation object.

**Rule**  \
A Rule applies to an instance of a *start* Class, and generates one or more references for a *goal* Class.
Rules are written in terms of domain-specific objects and query languages.
The start and goal of a rule can be in different domains (e.g. k8s/Pod → loki/application)
Rules are defined using Go templates, see ./rules for examples.

# Conflicting Vocabularies

Different signal and object domains may use different vocabularies to identify the same things.
For example:

- `k8s.pod.name` (traces)
- `pod` or `pod_name` (metrics)
- `kubernetes.pod_name` (logs)

The correlation problem would be simpler if there was a single vocabulary to describe signal attributes.
The [Open Telemetry Project](https://opentelemetry.io/) aims to create such a standard vocabulary.
Unfortunately, at least for now, we have to deal with multiple vocabularies in different domains.

For example, OpenShift tracing uses Open Telemetry notation. However, OpenShift logging and Prometheus metrics do not.
They each use a different convention that was established before Open Telemetry existed.
Historically, observability tools tended to develop in silos without standardization between domains.
Different conventions adopted in each domain are now entrenched and difficult to change.

A single vocabulary may eventually become universal, but in the short to medium term we have to handle mixed signals.
Korrle8 expresses rules in the native vocabulary of each domain, but allows rules to cross domains.

# Request for Feedback

If you work with OpenShift or kubernetes clusters, your experience can help to build a useful rule-base.
If you are interested, please [create a GitHub issue](https://github.com/korrel8/korrel8/issues/new), following this template:

## 1. When I am in this situation: ＿＿＿＿

Situations where:
- you have some information, and want to use it to jump to related information
- you know how get there, but it’s not trivial: you have to click many console screens, type many commands, write scripts or other automated tools.

The context could be
- interacting with a cluster via graphical console or command line.
- building services that will run in a cluster to collect or analyze data.
- out-of-cluster analysis of cluster data.

## 2. And I am looking at: ＿＿＿＿

Any type of signal or cluster data: metrics, traces, logs alerts, k8s events, k8s resources, network events, add your own…

The data could be viewed on a console, printed by command line tools, available from files or stores (loki, prometheus …)

## 3. I would like to see: ＿＿＿＿

Again types of information include: metrics, traces, logs alerts, k8s events, k8s resources, network events, add your own…

Describe the desired data, and the steps needed to get from the starting point in step 2.

Examples:
- I’m looking at this alert, and I want to see …
- I’m looking at this k8s Event, and I want to see …
- There are reports of slow responses from this Service, I want to see…
- CPU/Memory is getting scarce on this node, I want to see…
- These PVs are filling up, I want to see…
- Cluster is using more storage than I expected, I want to see…

