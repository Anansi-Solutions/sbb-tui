![GitHub contributors](https://img.shields.io/github/contributors/necrom4/sbb-tui?style=for-the-badge&link=https%3A%2F%2Fgithub.com%2FNecrom4%2Fsbb-tui%2Fgraphs%2Fcontributors)
![GitHub Release](https://img.shields.io/github/v/release/necrom4/sbb-tui?sort=semver&style=for-the-badge)
![GitHub License](https://img.shields.io/github/license/necrom4/sbb-tui?style=for-the-badge)

## How to contribute to SBB-TUI

### **Issues**

- **Ensure the bug was not already reported** by searching on GitHub under [Issues](https://github.com/necrom4/sbb-tui/issues).
- If you're unable to find an open issue addressing the problem, [open a new one](https://github.com/necrom4/sbb-tui/issues/new). Be sure to include:
  - **Title and clear description** with as much relevant information as possible
  - A **code sample** or an **executable test case** demonstrating the expected behavior that is not occurring. (**screenshots** appreciated)

### **Pull Requests**

- Start by creating an **Issue** addressing the bug/feature, create a **Fork** and start coding. If the idea for the change is accepted, you'll have the green light to open a PR with your code.
- Ensure the PR description clearly describes the problem and solution. Include the relevant issue number.
- **WARNING**: Install [mise](https://mise.jdx.dev/), it will help you setup your environment to properly start committing. The [mise.toml](https://github.com/Necrom4/sbb-tui/blob/master/mise.toml) file has a command to setup the pre-commit hooks and install the formatter, linter and docs checker, which the hooks rely on.
- Write **granular** commits each defining a single change. Titles must be meaningful and commit bodies should include in depth explanations if necessary. (No one should have to look at your code to understand what it does). Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/)
- Be aware that we are using [SemVer](https://semver.org/), your commit types must hence follow that logic. (e.g. `feat:` bumps a minor `vX.+1.X` version, `fix:` bumps a patch `vX.X.+1` and other commits do not generate a new release of the TUI.)
- Comment your code according to [godoc](https://go.dev/blog/godoc) but don't overdo it, a good function name should in theory be enough. Also comment small chunks of code that aren't easily understandable at first sight, if applicable.

Thanks for wanting to improve this fabulous tool!
