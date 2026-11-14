<!-- SPDX-License-Identifier: Apache-2.0 OR MIT -->

# 0004. Retries are opt-in and limited to GET

- **Status:** Accepted
- **Date:** 2026-11-14 (release 0.0.88)
- **Deciders:** maintainer

## Context

Until 0.0.88 the client never retried: it cannot know which calls are
safe to repeat against a given deployment. Services behind a load
balancer still see transient failures (a refused connection during a
rollout, a 503 from a draining replica, a 429 from a rate limiter) and
each had to write its own loop. The suite's SDKs agreed to offer the
same opt-in policy.

## Options considered

1. Keep no retries. Safest, but pushes the same loop into every caller.
2. Retry every call. Repeating a `Generate` or `Validate` POST whose
   response was lost could submit a payment file twice downstream.
3. Retry idempotent calls only, and only when asked.

## Decision

Option 3. `WithRetry(RetryPolicy{MaxAttempts, BaseDelay, MaxDelay})`:

- applies to GET only: `Health`, `Faces` and `Request` with GET. A POST
  (`Validate`, `Generate`, `Check`, `ProjectFace`, `Request` with any
  other method) is never retried;
- retries a failed connection (no response, context still live) and a
  `429` or `503`; any other status is returned at once;
- waits with exponential backoff (BaseDelay doubled per attempt,
  capped at MaxDelay) and jitter drawn from the upper half of the wait,
  or for the response's `Retry-After` (seconds or an HTTP date);
- gives up when a `Retry-After` exceeds MaxDelay, returning that
  response rather than sleeping past the caller's budget;
- shares one budget across attempts: the client timeout and the
  caller's context bound them all, and the last attempt's result is
  returned when either ends.

A zero `RetryPolicy`, or `MaxAttempts` below 2, means one attempt, which
is also the default without the option.

## Consequences

- Behaviour without `WithRetry` is unchanged from 0.0.87.
- Callers that need to retry a POST must decide its safety themselves
  and loop outside the client.
- `retry_test.go` covers every branch, including overflow of the
  doubling and a cancelled context mid-wait.
