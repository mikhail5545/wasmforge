# Contributing to WasmForge

Thanks for considering a contribution.

## Ways to contribute

- Report bugs and edge cases
- Improve docs and examples
- Add features from the roadmap
- Improve performance, reliability, and tests

## Development setup

1. Fork and clone the repo.
2. Install prerequisites:
   - Go 1.25+
   - Node.js 18+
   - make
3. Build the project:

```bash
make build
```

4. Run tests:

```bash
go test ./...
```

5. Build admin UI:

```bash
npm --prefix ui/adminv2 run build
```

## Workflow

1. Create a branch from `master`
2. Keep changes focused and small
3. Add/adjust tests for behavior changes
4. Ensure build and tests pass
5. Open a pull request with:
   - What changed
   - Why it changed
   - Any migration or compatibility notes

## Pull request guidelines

- Keep PRs reviewable (avoid giant mixed changes)
- Update docs if API/UI behavior changes
- Don’t include unrelated refactors in the same PR
- Prefer clear commit messages

## Bug reports

When filing an issue, include:

- Expected behavior
- Actual behavior
- Reproduction steps
- Logs/errors/screenshots (if available)

## Security

If you find a security issue, please avoid public disclosure before maintainers can triage and patch it.
