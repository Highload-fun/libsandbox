# libsandbox

`libsandbox` is a small Go library that provides a declarative API for running programs inside a sandbox environment.

It is a thin wrapper over the **sandbox** command-line tool and is responsible only for:

* describing how a process should be executed inside a sandbox
* assembling filesystem mappings, environment, and resource constraints
* invoking the sandbox executable via `os/exec`

All execution semantics, isolation guarantees, and validation rules are defined by the sandbox tool itself.

---

## Sandbox tool

This library relies on the external `sandbox` executable.

Canonical documentation, usage, and guarantees are defined in the sandbox repository:

👉 [https://github.com/Highload-fun/sandbox](https://github.com/Highload-fun/sandbox)

If something is unclear or unspecified in this library, always refer to the sandbox tool documentation.

---

## Design goals

* **Declarative API** — describe *what* should happen, not *how*
* **No duplication of CLI semantics** — the sandbox tool is the single source of truth
* **Low-level and predictable** — arguments are built explicitly and passed as-is
* **Minimal abstraction** — no hidden behavior, no magic defaults

This makes `libsandbox` suitable for infrastructure code, CI systems, judges, and high-load execution environments.

---

## Usage example

```go
sb := sandbox.New("/tmp/sandbox")

cmd := sb.
    AddFile("/usr/bin/go", "/usr/bin/go", true).
    MountDir("/workspace", "/workspace").
    AddEnv("PATH=/usr/bin").
    SetNoNewNet(true).
    SetMemLimit(512 * 1024 * 1024).
    ExecDir("/workspace").
    Command("go", "test", "./...")

cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr

if err := cmd.Run(); err != nil {
    log.Fatal(err)
}
```

---

## Package philosophy

* The Go API expresses **intent**, not command-line syntax
* The library does **not** attempt to validate sandbox configuration
* Any future changes in sandbox flags should not require changes in `libsandbox`

If you need behavior that is not exposed here, it should be added to the sandbox tool first.

---

## Concurrency

`Sandbox` is a mutable builder and is **not safe for concurrent use**.

If you need to reuse a base configuration across goroutines, create separate instances per execution.
