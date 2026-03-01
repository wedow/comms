---
id: com-mwpq
status: closed
deps: [com-zl1e]
links: []
created: 2026-03-01T13:18:20Z
type: task
priority: 2
assignee: Greg Wedow
tags: [phase-3]
---
# comms-v1: Callback rate limiting

## Summary
Enforce minimum interval between callback invocations using the `config.callback.delay` value (e.g. "5s"). If messages arrive faster than the delay, callbacks are skipped (not queued). Simple time-based gate: store the last callback invocation time, skip if less than `delay` has elapsed. Global scope (not per-channel).

## Files
- `internal/daemon/callback.go` -- add `CallbackRunner` struct with `lastRun time.Time` and `delay time.Duration`; method `Run(command string, env CallbackEnv)` that checks delay before executing
- `internal/daemon/callback_test.go` -- test that rapid invocations are throttled, test that invocations after delay pass through

## Implementation Notes
- Dependencies: Task 3.6 (callback execution)
- Struct definition:
  ```go
  type CallbackRunner struct {
      lastRun time.Time
      delay   time.Duration
  }
  ```
- Method: `Run(command string, env CallbackEnv)` -- checks if `time.Since(lastRun) >= delay` before calling `ExecCallback`
- Rate limiting is global (not per-channel)
- If messages arrive faster than delay, callbacks are SKIPPED (not queued)
- The `delay` value comes from `config.callback.delay`

## Verification
```sh
cd /home/greg/p/comms && go test ./internal/daemon/... -run TestCallbackRate -v
```

## Acceptance Criteria
- Rapid invocations are throttled
- Invocations after delay proceed

