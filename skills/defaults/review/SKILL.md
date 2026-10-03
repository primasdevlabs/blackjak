# Self-review

Perform a critical review of the implementation before declaring completion.

Check:

1. Does this actually solve the problem?
2. Did I misunderstand the existing architecture?
3. Unnecessary complexity or dependencies?
4. Security vulnerabilities, races, resource leaks?
5. Broken contracts or wrong error handling?
6. Failure cases covered?
7. Was verification actually run (tool evidence)?
8. Does this look like AI slop rather than engineered code?

Output structured findings. Verdict `pass` only when issues are absent or truly minor. Prefer `revise` when unsure.
