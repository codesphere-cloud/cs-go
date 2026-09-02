# Skill Family Conventions

> Shared rules for every skill in the `codesphere` family (`codesphere`,
> `codesphere-create-cluster-deployment`, `codesphere-create-container-deployment`,
> `codesphere-create-reactive-deployment`, `codesphere-add-managed-service`,
> `codesphere-create-provider`, `codesphere-run-deployment`). Each skill's own Hard Gate points
> here instead of restating these rules in full — read this once, apply it wherever a
> `references/*.md` path or `ci.yml` placement comes up in that skill's own instructions.

## Locating and reading `codesphere`'s `references/*.md` files

Only the `codesphere` skill ships a `references/` folder. Every other skill in this family has
none of its own — every `references/<file>.md` path mentioned anywhere in this family's skills
lives inside the sibling `codesphere` skill's own directory. Resolving such a path relative to
the *calling* skill's own installed location fails, or silently looks in the wrong place.

- **Locate it first.** If `codesphere`'s exact install path isn't already known, `Glob` for
  `**/codesphere/references/*.md` to find it, then read `<codesphere-path>/references/<file>.md`
  directly.
- **A `references/*.md` read is a plain file read** (`Read`/`Glob`/`Grep`) — it does **not**
  require `codesphere` to appear in the available-skills list, or be triggered/loaded as an
  active skill. It costs only the tokens of the specific file read, never `codesphere`'s own
  `SKILL.md` content or any other sibling skill's content. Don't conflate "needs one of
  `codesphere`'s reference files" with "needs the `codesphere` skill loaded as an active skill" —
  this is what keeps the family's per-invocation cost flat as more skills are added to it.
- **If `codesphere` can't be located at all**, stop and tell the user it needs to be installed
  alongside this skill — don't guess at field/schema content from memory.

## `ci.yml` placement

`ci.yml` always belongs at the repository root — never in a subdirectory, even in a monorepo
with multiple components/charts. Every skill that generates or edits a `ci.yml` follows this
without exception.
