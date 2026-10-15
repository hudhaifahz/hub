# Exact-channel rebalance launch drafts

Status: **PREPARED — NOT SIGNED — NOT PUBLISHED**

These drafts are not publication authorization. The public release, portal, and
canonical companion URLs below were independently read back on 2026-10-14.
Upstream pull requests do not yet exist and are deliberately not claimed.

## Nostr kind-1 note

Timestamp policy: set `created_at` only when the owner approves signing. Do not
pre-sign or reuse an expired signing request.

Proposed tags:

```json
[
  ["t", "bitcoin"],
  ["t", "lightning"],
  ["t", "albyhub"]
]
```

Proposed content:

```text
Grynvault Rebalance is an experimental, independently maintained macOS fork of Alby Hub for local circular rebalancing when one channel has enough outbound balance and another has enough inbound capacity.

It binds the exact first and final channels inside the local Hub/LDK process, with non-paying route discovery, typed confirmation, no fallback route, and no automatic payment retry. It cannot manufacture liquidity or guarantee a viable Lightning route.

Source and PGP-authenticated release:
https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.2

Read-only documentation and diagnostics:
https://rebalance.grynvault.com

Portal:
https://frontiercrown.com/portal

The macOS application is unsigned by Apple and unnotarized. Verify the PGP-signed manifest and expect Gatekeeper to warn or block it. Not independently audited.

Not an official Alby release or endorsement.
```

Before requesting signing approval:

- verify every linked artifact, source tag, manifest signature, documentation
  page, and portal link from a fresh profile;
- do not add an upstream-contribution claim until the exact PR chain is public;
- compare with prior owner posts to avoid duplicating an unchanged announcement;
- show the owner this exact text, kind `1`, tags, relay set, and timestamp policy;
- exclude node identifiers, balances, wallet history, routing tactics, keys, and
  operational logs; and
- obtain separate approval for the exact event, then read the signed event back
  from at least one target relay.

## Stacker News link post

This remains a template until the owner approves submission. The immutable
public GitHub release is the current canonical link; replace it only with a
freshly verified long-form article.

- Title: `Grynvault Rebalance: exact-channel circular rebalancing for Alby Hub`
- URL: `https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.2`
- Territory: `bitcoin` (reconfirm that this is the narrowest available relevant
  territory immediately before submission)
- Crosspost to Nostr: off
- Description:

```text
An experimental independent macOS fork that adds fail-closed circular rebalancing through owner-selected outgoing and incoming channels. Route discovery is non-paying; execution requires a fresh exact confirmation and never falls back to another channel. Includes source, a PGP-signed manifest, checksums, migration guidance, and security notes. The app is unsigned by Apple, unnotarized, and not independently audited.
```

Before requesting submission or payment approval:

- open and verify the final canonical article, title, and author identity;
- perform a fresh exact-URL duplicate check on Stacker News;
- confirm the final territory and keep Nostr crossposting disabled;
- read the current posting fee and show the owner the exact title, URL,
  territory, description, and sat amount;
- obtain approval for that specific fee and post; and
- after one submission/payment attempt, reconcile settlement before any retry
  and record the live item URL plus the actual credit/sat breakdown.
