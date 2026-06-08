# Lox – Durable Blocking Lock Service

## Overview

**Lox** is a single-instance, durable, blocking lock service.

It provides mutual exclusion for distributed or multi-process systems
with the following characteristics:

- Single instance (no HA for now)
- Durable lock state (survives restart)
- Blocking lock acquisition (FIFO queue)
- TTL-based expiration
- Fencing tokens for correctness

This document defines the development roadmap and feature breakdown.

---

# 1. Goals

## Primary Goals

- Provide correct mutual exclusion
- Support blocking lock acquisition
- Persist lock state across restarts
- Provide predictable and documented semantics
- Be extensible toward distributed/HA architecture

## Non-Goals (For Now)

- High availability
- Multi-instance coordination
- Horizontal scalability
- Consensus algorithms (Raft/Paxos)

---

# 2. High-Level Architecture

Lox consists of three main layers:

1. API Layer (gRPC)
2. Lock Manager (in-memory state + concurrency control)
3. Persistence Layer (SQlite)
