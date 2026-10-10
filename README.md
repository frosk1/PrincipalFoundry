# PrincipalFoundry

PrincipalFoundry is a Git-native framework and bootstrap CLI for senior-led AI software teams.

The project is built around a simple operating model:

- senior engineers and a strong architect model define product intent, architecture, and the next small task;
- implementation agents execute one task at a time;
- tester and reviewer agents provide independent evidence;
- deterministic build, unit, integration, API, and optional performance gates decide whether work is ready for architect review;
- Git is the shared state, audit trail, and handoff mechanism.

## Status

Early bootstrap. The first milestone is a portable CLI that can create a new PrincipalFoundry project with the appropriate control files and agent configuration.

Planned first workflow:

```text
principalfoundry new my-project \
  --architect chatgpt \
  --executor opencode
```

The initial generated project starts in `DISCOVERY` and does not allow autonomous implementation until its product specification and architecture have been established and manually validated.
