# Accessibility checks

Three scripts that run against a **running** instance, because what
matters is the rendered page, not the template.

| Script | What it asserts |
|---|---|
| `audit.mjs` | axe-core, WCAG 2.1 A and AA, on every page in both themes |
| `dialogs.mjs` | the same on each command dialog while it is open, and that Escape closes it |
| `keyboard.mjs` | every tab stop is visible and has a focus ring, and the skip link lands on the main region |

```bash
npm install            # in this directory, once
SEI_URL=http://127.0.0.1:8090 SEI_USER=ops SEI_PASS=... node audit.mjs
```

The account needs every read permission, so `operator` or `admin`.

These are deliberately not part of `make test`: they need a live server
with data behind it. `make test-a11y` runs all three.

## What they will not tell you

An automated audit checks what a machine can check. It does not know
whether a label is the right word, whether the reading order makes
sense, or whether an operator can find the thing they came for. The
colour tokens are also computed rather than audited - see the contrast
note in `frontend/src/styles.css`, which is what keeps every text colour
above 4.5:1 against all three surfaces of its theme.
