# Design QA: responsive independent-fork notice

## Result

Passed on 2026-10-17.

The independent-fork notice now occupies the page inset beside the desktop
sidebar. At the narrow desktop width from the reported failure it wraps to two
centered lines, remains fully visible, and introduces no horizontal overflow.

## Reference and implementation captures

- Source reference: `/Users/kode/Desktop/Screenshot 2026-10-17 at 10.05.06 PM.png`
  (2026 × 278 physical pixels). The fixed sidebar obscures the beginning of the
  notice.
- Wider source reference: `/Users/kode/Desktop/Screenshot 2026-10-17 at 10.05.14 PM.png`
  (3006 × 358 physical pixels). The notice happens to fit because the viewport
  is wider.
- Implementation: `docs/grynvault-release/assets/banner-small-after.png`
  (2026 × 600 physical pixels; 1013 × 300 CSS-pixel viewport at 2× scale).

The implementation capture renders the production `GrynvaultForkNotice`
component in the corrected desktop sidebar/content nesting with synthetic
shell content. It does not connect to or copy a wallet.

## Comparison

| Check | Result |
| --- | --- |
| Notice begins after the sidebar boundary | Pass |
| Product name and full disclaimer are visible | Pass |
| Source-review link remains visible | Pass |
| Long copy wraps within the content inset | Pass |
| No horizontal document overflow | Pass |
| Existing dark theme, color, type, and spacing preserved | Pass |

## Iteration history

1. The original full-width notice rendered underneath the fixed desktop
   sidebar. Centering only concealed the defect at wider window sizes.
2. The notice was moved into `SidebarInset`, before the page content. The
   existing upstream banner offset was moved with it so the two notices cannot
   overlap.
3. The narrow-width capture was compared directly with the reported failure.
   No additional visual correction was required.
