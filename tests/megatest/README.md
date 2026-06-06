# megatest scenarios

User-journey end-to-end tests for the `megaport-cli` binary. The scenarios
in this directory drive the actual built binary against the staging API
through the megatest `api-tester` YAML runner.

These are deliberately separate from the per-package Go integration tests
under `internal/commands/*/*_integration_test.go`. Those tests live next
to the code and exercise individual command actions in process. The
scenarios here are out of process, story shaped, and verify that the
shipped binary behaves correctly when chained together over a real API.

## Running locally

You need a clone of `megaport/megatest` next to this repo (or anywhere on
disk) and a working `python3` with `api-tester/requirements.txt` installed.

```bash
# 1. Build the CLI
go build -o megaport-cli .
export PATH="$PWD:$PATH"

# 2. Clone megatest as a sibling
git clone https://github.com/megaport/megatest.git ../megatest
pip install -r ../megatest/api-tester/requirements.txt

# 3. Configure megatest's env (api keys, base URL)
cp ../megatest/tests/credentials-example.yml ../megatest/api-tester/credentials.yml
# Fill in staging credentials in credentials.yml

# 4. Run a scenario
cd ../megatest
MEGAPORT_ENV=staging python3 api-tester/runner.py \
    ../megaport-cli/tests/megatest/lifecycle/port.yml
```

The runner writes archived results under `runs/run-NN/` next to the
scenario file.

## Authoring new scenarios

For ad hoc lifecycle smokes keep them under `lifecycle/<resource>.yml`.
For ticket-driven scenarios mirror megatest's own convention and add a
folder per ticket:

```
tests/megatest/
├── lifecycle/
│   └── port.yml
└── stories/
    └── ESD-XXXX/
        ├── users.yml
        ├── products.yml
        └── XXXX_01_*.yml
```

Conventions to follow:

- Name every resource you create `megatest-{run_uid}` so concurrent runs
  do not collide. `{run_uid}` is auto-seeded by the runner.
- Always wire a teardown step that deletes anything the suite created,
  guarded by `stop_on_failure: true` so a mid-suite failure still hits
  cleanup.
- Never touch a resource the scenario did not provision in this run.
- Use `--output json` on every CLI invocation. Capture downstream
  values via JSONPath against the parsed stdout. Note that
  `ports buy` emits the new UID through stderr; capture by name from a
  follow-up `ports list` JSON instead.

## What each file is for

| File | Purpose |
|------|---------|
| `lifecycle/port.yml` | Canonical buy/list/delete smoke for a 1Gbps port. |
| `users.yml` | Documents which user aliases scenarios expect (no credentials). |
| `products.yml` | Shared placeholder constants (location IDs, defaults). |
