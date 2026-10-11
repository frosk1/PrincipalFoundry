<p align="center">
  <img src="docs/assets/logo.png" alt="PrincipalFoundry" width="760">
</p>

<p align="center">
  <strong>Git-native software factories for principal engineering teams.</strong>
</p>

PrincipalFoundry is a Git-native framework and bootstrap CLI for senior-led AI software teams.

The core idea is simple: a small number of senior engineers work with a strong architect agent to define product direction and architecture, while implementation agents perform most coding, testing, verification, and routine review.

The CLI executable is `pfoundry`.

## The team model

Traditional software delivery spreads product, architecture, implementation, QA, and operations across many specialized roles. PrincipalFoundry explores a different model: one or two senior engineers retain product and engineering judgment while an architect agent continuously decomposes, reviews, and coordinates work with implementation, tester, reviewer, and automation agents.

<p align="center">
  <img src="docs/assets/team_setup.png" alt="Traditional software team evolving into a Principal Engineering Team and AI software factory" width="100%">
</p>

The human core stays intentionally small. The software factory is the execution system around it: the architect agent and execution agents work in a Git-native feedback loop, returning implementation evidence to the senior engineer instead of replacing senior engineering judgment.

## Operating model

PrincipalFoundry separates architectural judgment from implementation throughput.

```text
Senior engineer
      ↕
Architect agent
      │
      ├── product requirements
      ├── architecture
      ├── architecture decisions
      ├── task decomposition
      └── final implementation review
      │
      ▼
Git repository
      │
      ▼
Implementation agent ◄───────────────────────────────┐
      │                                              │
      ├── impact analysis                            │
      ├── coding                                     │
      ├── targeted tests                             │
      ├── integration verification                   │
      └── implementation evidence                    │
      │                                              │
      ├──────────────┐                               │
      ▼              ▼                               │
 Tester agent    Reviewer agent                      │
      │              │                               │
      └──────┬───────┘                               │
             ▼                                       │
 deterministic verification                         │
             │                                       │
             ▼                                       │
       Architect review                              │
             │                                       │
       accept / revise                               │
             │                                       │
             ▼                                       │
        next small task ─────────────────────────────┘
```

Git is the shared state, handoff mechanism, and audit trail between the architect and execution side.

## Starting a new project

The intended onboarding flow starts with **product and architecture discovery before the software repository is bootstrapped**.

The software factory should begin from an approved design baseline, not from an empty repository where implementation agents are expected to discover fundamental product and architecture decisions themselves.

### 1. Start the architect discovery session

Begin with a strong architect model and the senior engineer.

Use the neutral kickoff prompt in:

```text
docs/prompts/discovery-kickoff.md
```

The prompt instructs the architect to act as a senior software architect and technical product owner, work iteratively with the user, challenge assumptions, clarify the most important product and architecture questions, and avoid implementation work during this phase.

The planned CLI experience is:

```bash
pfoundry start my-project \
  --architect chatgpt \
  --executor opencode
```

`pfoundry start` will print the kickoff prompt and the expected output package. This command is not implemented yet; today the prompt can be copied directly from `docs/prompts/discovery-kickoff.md`.

### 2. Run discovery iteratively

The architect and senior engineer work conversationally until the important product and architecture uncertainties are sufficiently clear.

The architect should explore the areas that materially affect the system, including:

- product goal and target users;
- core workflows and first-version scope;
- explicit non-goals;
- business and technical constraints;
- scale, reliability, performance, security, and privacy requirements;
- external integrations;
- deployment environment;
- major technology constraints or preferences;
- system boundaries and data ownership;
- unresolved decisions.

This is deliberately not a one-shot questionnaire. The architect should use each answer to determine the next highest-value questions and actively provide recommendations and trade-offs.

### 3. Approve the initial design baseline

When the architect believes the initial design is sufficiently clear, it summarizes the proposed baseline and remaining uncertainties.

The senior engineer explicitly approves or corrects that baseline before the final design package is produced.

No implementation task should exist yet.

### 4. Produce the design package

After approval, the architect produces:

```text
design-package/
├── PROJECT_BRIEF.md
├── PRODUCT_SPEC.md
├── ARCHITECTURE.md
├── OPEN_QUESTIONS.md
└── decisions/
    ├── INDEX.md
    └── ADR-*.md
```

ADRs are created only for material architectural decisions that were actually discussed and approved. Open or intentionally deferred decisions remain in `OPEN_QUESTIONS.md`.

This package is the handoff between the discovery/design phase and the software factory.

During repository bootstrap, the files map into the repository as follows:

```text
PROJECT_BRIEF.md        → .factory/PROJECT_BRIEF.md
PRODUCT_SPEC.md         → docs/PRODUCT_SPEC.md
ARCHITECTURE.md         → docs/ARCHITECTURE.md
OPEN_QUESTIONS.md       → docs/OPEN_QUESTIONS.md
decisions/*             → docs/decisions/*
```

The project brief remains bootstrap/orientation context, while product, architecture, open questions, and ADRs become durable design knowledge.

### 5. Initialize the project repository from the approved design

The planned repository bootstrap command is:

```bash
pfoundry init my-project \
  --from ./design-package \
  --architect chatgpt \
  --executor opencode
```

The intended behavior is:

```text
validate design package
        ↓
create repository structure
        ↓
install approved product + architecture documents
        ↓
create .factory/ control files
        ↓
create AGENTS.md
        ↓
create executor-specific tester/reviewer agents
        ↓
git init
        ↓
bootstrap commit on main
```

The repository is therefore created from an already-understood product and architecture baseline.

`pfoundry init --from` is not implemented yet. The current `pfoundry new` command still creates the repository directly in `DISCOVERY`; this will evolve toward the discovery-first flow described above.

### 6. Connect the remote repository

Remote creation stays separate from the local bootstrap so the workflow remains provider-independent.

After initialization, connect an empty GitHub, GitLab, or other Git remote:

```bash
git remote add origin REMOTE_URL
git push -u origin main
```

Do not initialize the remote with a README, `.gitignore`, or license when the local repository already contains the bootstrap commit.

The repository becomes the shared control plane:

```text
Git repository
   ├── senior engineer / architect agent
   └── selected executor
       ├── implementer
       ├── tester
       └── reviewer
```

### 7. Architect issues the first implementation task

Only after the approved design has been installed in the repository does the architect create the first bounded implementation task in:

```text
.factory/CURRENT_TASK.md
```

The first task should normally establish the minimum project scaffolding and verification harness required by the approved architecture.

A project should then evolve through small coherent increments:

```text
TASK-0001  project skeleton and build/test tooling
     ↓
TASK-0002  minimal runnable application
     ↓
TASK-0003  required persistence/infrastructure
     ↓
TASK-0004  first thin end-to-end product capability
     ↓
...
```

The architect should never issue a task equivalent to “implement the entire product”.

### 8. Run the implementation cycle manually

Before enabling automation, run the executor manually.

The implementation agent follows `AGENTS.md` and processes exactly one active task.

Its cycle is:

```text
CURRENT_TASK
     ↓
read architecture/spec/ADRs
     ↓
impact analysis
     ↓
tester PREFLIGHT
     ↓
baseline tests
     ↓
implementation
     ↕
targeted test feedback
     ↓
required verification gates
     ↓
tester FINAL
     ↓
independent reviewer
     ↓
deterministic final verification
     ↓
factory evidence reports
```

The executor records evidence in:

```text
.factory/
├── IMPLEMENTATION_REPORT.md
├── TESTER_REPORT.md
├── INTERNAL_REVIEW.md
└── VERIFICATION_REPORT.md
```

### 9. Architect reviews the implementation

The architect reads the active task, implementation diff, implementation report, tester evidence, independent review, deterministic verification results, and relevant product and architecture documents.

It then chooses one of three outcomes:

```text
ACCEPT
  → close the current task
  → define the smallest useful next task

REQUEST REVISION
  → keep the same task id
  → increment the revision
  → incorporate blocking findings into CURRENT_TASK

BLOCK / ESCALATE
  → require senior-engineer product or architecture input
```

The architect should judge the actual code and evidence rather than trusting reports blindly.


## Architect cycle

Once a project is established, the architect loop is intentionally small and repeatable:

```text
current repository state
        +
PRODUCT_SPEC
        +
ARCHITECTURE / ADRs
        +
OPEN_QUESTIONS when relevant
        +
previous execution evidence
        ↓
review previous task
        ↓
accept / revise / escalate
        ↓
choose the smallest coherent next capability
        ↓
write exactly one CURRENT_TASK
        ↓
commit architect handoff
```

The architect is responsible for:

- product clarity;
- architecture;
- architecture decisions;
- task decomposition;
- task acceptance criteria;
- required verification gates;
- reviewing executor evidence;
- deciding whether work is accepted;
- choosing the next incremental task.

The architect should not normally perform the implementation itself, even if technically capable of editing the repository.

## Executor cycle

The executor is responsible for implementation throughput within the architect-defined boundaries.

It may:

- inspect and understand the repository;
- choose local implementation details;
- implement the current task;
- create or update tests;
- invoke tester and reviewer agents;
- run builds and verification;
- update evidence files.

It may not independently:

- change product requirements;
- change acceptance criteria;
- change major architecture;
- introduce major infrastructure not authorized by the task;
- start another task;
- self-approve architecture changes.

If a task requires such a decision, the executor returns `ARCHITECT_DECISION_REQUIRED`.

## Lifecycle

The overall process moves through explicit phases:

```text
DISCOVERY
   ↓
DESIGN
   ↓
REPOSITORY_BOOTSTRAP
   ↓
SCAFFOLDING
   ↓
MANUAL_VALIDATION
   ↓
AUTONOMOUS
   ↕
PAUSED
```

### DISCOVERY

The senior engineer and architect establish the product, users, scope, constraints, and important unknowns. This phase can happen before a project repository exists.

### DESIGN

The architect converts the approved discovery outcome into the design package: product specification, architecture, open questions, and material ADRs.

### REPOSITORY_BOOTSTRAP

The approved design package is turned into the Git repository, factory control files, executor contract, and executor-specific subagents.

### SCAFFOLDING

The implementation agent creates the initial application structure according to the approved architecture.

### MANUAL_VALIDATION

Architect and executor cycles are triggered manually to validate task sizing, agent behavior, tests, Git handoffs, and verification.

### AUTONOMOUS

The same architect/executor loop can now be triggered on schedules or other automation.

### PAUSED

No autonomous implementation should proceed.

## Generated project structure

After repository bootstrap, the project structure is intended to contain:

```text
my-project/
├── AGENTS.md
│
├── docs/
│   ├── PRODUCT_SPEC.md
│   ├── ARCHITECTURE.md
│   ├── OPEN_QUESTIONS.md
│   └── decisions/
│       └── INDEX.md
│
├── .factory/
│   ├── FACTORY_STATE.md
│   ├── PROJECT_BRIEF.md
│   ├── CURRENT_TASK.md
│   ├── REVIEW.md
│   ├── IMPLEMENTATION_REPORT.md
│   ├── TESTER_REPORT.md
│   ├── INTERNAL_REVIEW.md
│   └── VERIFICATION_REPORT.md
│
└── executor-specific subagent configuration
```

Executor-specific subagents are generated only for the selected executor:

| Executor | Generated project files |
| --- | --- |
| `opencode` | `.opencode/agents/tester.md`, `.opencode/agents/reviewer.md` |
| `claude-code` | `.claude/agents/tester.md`, `.claude/agents/reviewer.md` |
| `codex` | `.codex/agents/tester.toml`, `.codex/agents/reviewer.toml` |
| `none` | No executor-specific subagent files |

The conceptual split is:

```text
docs/
= durable product and architecture knowledge

.factory/
= current PrincipalFoundry workflow state and evidence

AGENTS.md
= executor operating contract

executor-specific agent directory
= tester/reviewer definitions in the selected executor's native format
```

## Why Git?

PrincipalFoundry deliberately uses Git as the initial factory control plane.

A typical history can look like:

```text
bootstrap: import approved product and architecture baseline
architect: issue TASK-0001 r1
executor: complete TASK-0001 r1
architect: accept TASK-0001 and issue TASK-0002
executor: complete TASK-0002 r1
architect: request TASK-0002 r2
executor: complete TASK-0002 r2
```

This provides:

- atomic handoffs;
- exact code versions;
- diffs;
- audit history;
- rollback;
- branch protection;
- human review;
- CI integration;
- no additional workflow database for the initial design.

## Status

PrincipalFoundry is currently in early bootstrap.

Today, `pfoundry new` can generate the repository structure, executor-specific subagents, initialize Git, and create the bootstrap commit. The neutral discovery prompt is already versioned in `docs/prompts/discovery-kickoff.md`.

The onboarding model is now moving to the discovery-first flow described above. The next CLI milestone is to implement `pfoundry start` and `pfoundry init --from`, so an approved design package becomes the input to repository bootstrap. After that, the main remaining work is stronger structured task/report contracts, lifecycle commands, optional remote-provider integration, and automation support.
