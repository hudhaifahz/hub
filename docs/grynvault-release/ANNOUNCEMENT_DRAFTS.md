# RC.6 announcement package

Status: **PREPARED — NOT SIGNED — NOT PUBLISHED**

Prepared: 2026-10-18

This package is editorial preparation only. It does not authorize a Nostr
signature, relay publication, Stacker News submission, posting-fee payment,
cross-post, zap, boost, release edit, tag, or artifact upload.

## Evidence boundary

- Release: `grynvault-rebalance-v0.1.0-rc.6`
- Source commit: `408c74f9b06664b156df298e6fd6841a3a4db2b7`
- Upstream base: Alby Hub `v1.24.3`
- Release URL:
  <https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.6>
- Documentation: <https://rebalance.grynvault.com/>
- Current public binary: universal macOS only, requiring macOS 15.4 or newer
- Authenticity: detached PGP signature over the release manifest, fingerprint
  `D131 A3D0 EAB0 CF98 A841 3853 CD9F DBF9 D075 C126`
- Distribution: experimental, independently maintained, ad-hoc signed, not
  Apple Developer ID signed or notarized, and not an official Alby release
- Live-use wording: the public release record contains owner-run live-path
  evidence for the retained exact-channel engine. Do not represent it as an
  independent audit or as a payment made during network-isolated RC.6 migration
  acceptance.

The application cannot manufacture liquidity or guarantee a route. It needs
one active channel with enough outbound liquidity, a different active channel
with enough inbound capacity, and a viable public Lightning route between the
selected peers.

### Internal post-release readback

The latest saved operation at preparation time independently read back as
`succeeded` for 10,000 sats with a 20.502-sat routing fee and zero unidentified
receiving channels. The installed application's native LDK library and
`Info.plist` matched the public RC.6 disk image, but its executable bytes did
not, even after removing the ad-hoc code signatures from temporary copies. The
installed executable also lacked the public RC.6 tag and source-commit strings
embedded in the published build.

Therefore, do not call this an execution of the downloadable RC.6 artifact.
It is additional owner-run live-path evidence only. Binding a live test to the
public artifact would require a separately approved, PGP-verified installation
and a separately approved rebalance; neither is authorized by this document.

## Nostr kind-1 note

Label: **PREPARED — NOT SIGNED — NOT PUBLISHED**

Timestamp policy: set `created_at` only when the owner approves the exact event
for signing. Do not reuse a stale prepared timestamp.

Proposed tags:

```json
[
  ["t", "lightning"],
  ["t", "bitcoin"],
  [
    "r",
    "https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.6"
  ]
]
```

Exact content:

```text
Grynvault Rebalance v0.1.0-rc.6 is available: an experimental, independently maintained Alby Hub v1.24.3 fork for exact-channel circular Lightning rebalances.

Choose the outgoing and return channels, find a non-paying local route, review the fresh quote and balance evidence, then confirm execution once. It is a local circular-payment alternative when the former swap-backed rebalance flow is unavailable.

It still needs real outbound liquidity, inbound capacity, and a viable route. macOS 15.4+ only for this RC; PGP-signed manifest; not Apple-notarized and not an official Alby release.

Source/download: https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.6
Docs: https://rebalance.grynvault.com/
```

Signing and publication remain separate gates:

1. Recheck every link and time-sensitive claim.
2. Show the owner the final serialized kind-1 event.
3. Obtain exact-event editorial approval.
4. Request signing through the installed Nostr signer; the owner personally
   confirms its prompt.
5. Obtain separate intent to publish that signed event.
6. Publish once, then read the event back from the selected relays before
   calling it live.

## Stacker News link post

Label: **PREPARED — NOT SUBMITTED — NOT PAID**

- Title: `Grynvault Rebalance RC.6: exact-channel circular rebalancing for Alby Hub`
- URL:
  `https://github.com/hudhaifahz/hub/releases/tag/grynvault-rebalance-v0.1.0-rc.6`
- Territory: `~lightning`
- Cross-post to Nostr: off
- Boost: none

This draft links directly to the immutable GitHub release because no canonical
Nostr long-form article exists. If the owner later approves a long-form Nostr
article, verify its stable public `naddr` URL and prefer that as the Stacker
News source instead of creating duplicate submissions.

Description:

```text
An experimental macOS release candidate built from Alby Hub v1.24.3. It lets an embedded-LDK node operator choose the exact outgoing and returning channels, find a non-paying local route, review fee and balance evidence, and execute only after a fresh exact confirmation. It also includes a fail-closed guided migration flow for an existing Alby Hub embedded-LDK wallet.

The manifest is PGP-signed, but the app is not Apple Developer ID signed or notarized. This is an independently maintained fork, not an official Alby release, and it cannot create missing liquidity or guarantee a route.
```

The 2026-10-18 preliminary public search found no Stacker News result for the
exact release URL, tag, or title. This is not the required final duplicate
check. Immediately before submission:

1. Verify that the release URL resolves and still matches the title and source
   commit above.
2. Run a fresh exact-URL duplicate query.
3. Recheck that `~lightning` remains the narrowest suitable territory.
4. Read the current posting fee from the authenticated submission screen or
   API. The public territory page showed 22 sats during preparation, but that
   value is dynamic and is not approved spending.
5. Present the exact title, URL, territory, description, and fee to the owner.
6. Submit only after explicit post approval, and pay only after separate
   approval of the exact fee or a stated maximum.
7. Verify the item is public and not pending, and record the final URL plus the
   actual cowboy-credit/sat payment breakdown.

## Public-response guardrails

- Say **release candidate**, not stable or production-guaranteed.
- Say **PGP-signed manifest**, not “PGP-signed application.”
- Say **exact local first and last channels**, not guaranteed network success.
- Do not disclose private node identities, wallet balances, NWC credentials,
  payment hashes, route bytes, private channel data, or operational strategy.
- Do not describe the informational hosted companion as a wallet-connected
  rebalance service. It cannot unlock, quote, sign, pay, or move sats.
- Linux and Windows are planned and unsupported until their own acceptance
  records are complete.
