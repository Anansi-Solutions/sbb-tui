![GitHub contributors](https://img.shields.io/github/contributors/necrom4/sbb-tui?style=for-the-badge&link=https%3A%2F%2Fgithub.com%2FNecrom4%2Fsbb-tui%2Fgraphs%2Fcontributors)
![GitHub Release](https://img.shields.io/github/v/release/necrom4/sbb-tui?sort=semver&style=for-the-badge)
![GitHub License](https://img.shields.io/github/license/necrom4/sbb-tui?style=for-the-badge)

## How to contribute to SBB-TUI

### **Issues**

- **Ensure the bug was not already reported** by searching on GitHub under [Issues](https://github.com/necrom4/sbb-tui/issues).
- If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/necrom4/sbb-tui/issues/new). Be sure to include:
  - **Title and clear description** with as much relevant information as possible
  - A **code sample** or an **executable test case** demonstrating the expected behavior that is not occurring. (**screenshots** appreciated)

### **Development setup**

The project uses [mise](https://mise.jdx.dev/) to manage its toolchain (Go, formatters, linters) and [hk](https://hk.jdx.dev/) to run the git hooks. mise installs everything at pinned versions, so your local checks match CI exactly. Set it up once:

1. **Install mise** by following the [official guide](https://mise.jdx.dev/getting-started.html#installing-mise-cli).
2. **Activate mise** in your shell, also per the [official guide](https://mise.jdx.dev/getting-started.html#activate-mise). This lets mise auto-load the project's environment when you `cd` into it.
3. **Clone the repo** (or your fork), then `cd` into it. If mise prompts you, run `mise trust` to allow this project's `mise.toml`.
4. **Run `mise install`.** This installs every tool and then automatically installs the hk git hooks (via the `postinstall` hook). No separate init step is needed.
5. **Verify the setup:** `cd` out of the project and back in. If you see no `[WARNING]` about missing tools or hooks, you're ready.

> [!TIP]
> If you don't activate mise (step 2), you can still run any tool with `mise exec -- <command>` (e.g. `mise exec -- go test ./...`).

**Checking your code:**

- `hk check` runs all formatters, linters, `go vet`, the build and the test suite in read-only mode (add `--all` to check the whole repo, not just changed files).
- `hk fix` does the same but auto-applies formatter/linter fixes.
- On `git commit`, the **pre-commit** hook auto-runs the fast file-scoped formatters and linters; on `git push`, the **pre-push** hook runs `go build` and the test suite. Both run automatically once `mise install` has set them up.

### **Pull Requests**

- Start by creating an **Issue** addressing the bug/feature, create a **Fork** and start coding. If the idea for the change is accepted, you'll have the green light to open a PR with your code.
- Ensure the PR description clearly describes the problem and solution. Include the relevant issue number.
- Make sure you've completed the [Development setup](#development-setup) above so the git hooks lint and test your commits.
- Write **granular** commits each defining a single change. Titles must be meaningful and commit bodies should include in depth explanations if necessary. (No one should have to look at your code to understand what it does). Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- Be aware that we are using [SemVer](https://semver.org/), your commit types must hence follow that logic. (e.g. `feat:` bumps a minor `vX.+1.X` version, `fix:` bumps a patch `vX.X.+1` and other commits do not generate a new release of the TUI.)
- Comment your code according to [godoc](https://go.dev/blog/godoc) but don't overdo it, a good function name should in theory be enough. Also comment small chunks of code that aren't easily understandable at first sight, if applicable.

Thanks for wanting to improve this fabulous tool!
