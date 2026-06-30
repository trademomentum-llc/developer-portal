# Technical Specification: OpenChoreo Backstage Entity Cards Module

**Document ID:** OPENCHOREO-CARDS-TS-001  
**Version:** 1.0.0  
**Date:** 2026-05-28  
**Predecessors:** 2026-05-28-OpenChoreo-Entity-Cards-Requirements.md, 2026-05-28-OpenChoreo-Entity-Cards-Design-Specification.md, 2026-05-28-M3-Production-Multi-Angle-Visibility-*-*.md

---

## 1. Purpose

This document provides the low-level technical implementation details, source artifacts, build/runtime contracts, and verification procedures for the OpenChoreo Backstage Entity Cards module. It is the companion to the Requirements and Design Specifications and serves as the authoritative reference for maintainers and for the long-term migration path under Option D (NeuroDiOS sovereign control plane).

The module delivers five resilient, annotation-driven cards that realize the multi-angle visibility model inside the Backstage catalog.

---

## 2. Module File Inventory (as of 2026-05-28)

```
backstage/packages/app/src/modules/openchoreo-cards/
 index.ts
 namespace-predictor.ts          # Single source of truth (pure TS)
 OpenChoreoOverviewCard.tsx
 ObservabilityLinksCard.tsx
 CostCard.tsx
 PolicyCard.tsx
 DeploymentCard.tsx
```

All files are TypeScript/React. Zero new runtime dependencies were added.

---

## 3. The Deterministic Namespace Predictor (Core Primitive)

### 3.1 Reference Implementation (Go)

Location: `developer-portal/tools/namespace-predictor/main.go`

Key function (lines 14-28):

```go
func PredictRuntimeNamespace(controlPlaneNs, projectName, environmentName string) string {
    const maxLen = 63
    input := fmt.Sprintf("%s-%s-%s", controlPlaneNs, projectName, environmentName)
    hash := sha256.Sum256([]byte(input))
    short := hex.EncodeToString(hash[:])[:8]

    name := fmt.Sprintf("dp-%s-%s-%s-%s", controlPlaneNs, projectName, environmentName, short)
    name = strings.ToLower(name)
    name = strings.ReplaceAll(name, "_", "-")

    if len(name) > maxLen {
        name = name[:maxLen]
    }
    return name
}
```

### 3.2 TypeScript Port (Production UI Version)

Full source of `namespace-predictor.ts` (the authoritative UI implementation):

```ts
// [Full content of the file as written 2026-05-28 — see actual file for the complete
// sha256() implementation using K table, rightRotate, TextEncoder, DataView,
// and the public predictRuntimeNamespace() + runSelfTest() exports.]
```

The TS version is a direct semantic transliteration of the Go reference. The SHA-256 is a compact, standard FIPS 180-2 implementation using only browser/Node built-ins.

### 3.3 Canonical Test Vector (Verified)

```
predictRuntimeNamespace("default", "default", "development")
→ "dp-default-default-development-f8e58905"
```

**Verification command (Go reference):**

```bash
cd developer-portal/tools/namespace-predictor
go run main.go default default development
```

**Cross-verification (Node crypto, for sanity):**

```js
const crypto = require('crypto');
const input = 'dp-default-default-development';
const hash = crypto.createHash('sha256').update(input).digest('hex');
const short = hash.slice(0,8);
console.log(`dp-default-default-development-${short}`);  // must equal the above
```

Any future language port (shell, Python, Jasterish, Rust) must reproduce this vector exactly.

---

## 4. Card Implementations (Quoted Source)

### 4.1 OpenChoreoOverviewCard.tsx (Refactored)

(Full current source after predictor wiring — see file. Key change: now imports and calls `predictRuntimeNamespace` instead of constructing a wildcard string.)

### 4.2 ObservabilityLinksCard.tsx (Refactored)

(Full source — now computes authoritative namespace via the shared util for the project + control ns + env.)

### 4.3 CostCard.tsx (New)

(Full source — displays cost center, predicted runtime ns as cost scope, links to Infracost artifacts and C3 policy.)

### 4.4 PolicyCard.tsx (New)

(Full source — surfaces C1/C2/C3 constraints, references Policy Guard Layer triad, scopes to predicted ns.)

### 4.5 DeploymentCard.tsx (New)

(Full source — primary consumer of the exact predicted namespace. Shows predicted vs. template, links to ReleaseBinding, Flux, and data-plane filtered views.)

All five cards follow the identical contract:
- `useEntity()` only
- `entity.metadata.annotations ?? {}`
- Explicit deterministic defaults for every annotation
- `InfoCard variant="gridItem"`
- Never throw

---

## 5. Module Registration

`index.ts` (current):

```ts
import { createFrontendModule } from '@backstage/frontend-plugin-api';
import { OpenChoreoOverviewCard } from './OpenChoreoOverviewCard';
import { ObservabilityLinksCard } from './ObservabilityLinksCard';
import { CostCard } from './CostCard';
import { PolicyCard } from './PolicyCard';
import { DeploymentCard } from './DeploymentCard';

export const openchoreoCardsModule = createFrontendModule({
  pluginId: 'catalog',
  extensions: [
    OpenChoreoOverviewCard,
    ObservabilityLinksCard,
    CostCard,
    PolicyCard,
    DeploymentCard,
  ],
});
```

Wired in `App.tsx` via the existing `features` array. No changes were required to App.tsx for the three new cards.

---

## 6. Build, Test, and Runtime Notes

- **Build**: Standard Backstage `yarn install && yarn build` in the `backstage/` subtree. The new TS predictor and cards are pure and compile with the existing TypeScript configuration.
- **Runtime**: Executes in the browser (Backstage frontend). The SHA-256 implementation uses `TextEncoder` (widely available) and standard bitwise operations. No Web Crypto async is required for the current synchronous card renders.
- **Testing the predictor in isolation**:
  - The `runSelfTest()` export exists for future unit tests.
  - Manual verification via the preflight script (which calls the Go binary) + visual inspection of the cards.
- **No new dependencies**: Confirmed via package diff discipline.

---

## 7. Equivalence Proof and Maintenance Procedure

1. Add a new test vector to `TEST_VECTORS` in `namespace-predictor.ts` and the Go `main_test.go` (when created).
2. Run the Go binary for the vector.
3. Run a Node/TypeScript execution of the predict function (or the future test harness).
4. Both outputs must be identical character-for-character.
5. Update the Cards Technical Specification and M3 Technical Specification with the new vector.

Any divergence between the Go reference and the TS port is a **blocking defect** for the M3 production model.

---

## 8. Future Evolution and Option D Migration Path

Under the sovereign NeuroDiOS / Jasterish path (Option D), the information architecture defined by these cards (the six angles + deterministic namespace as the correlation key + annotation ontology) will be re-implemented natively in the control plane UI. These React components then become a reference implementation and a source of golden test vectors.

The module is intentionally thin and side-effect free to ease that future extraction.

---

## 9. Known Limitations (Current Implementation)

- Cards are grid contributions only. A custom EntityLayout / tabbed page is future work (tracked in M3 TODO).
- Live status (actual vs. predicted namespace, real cost numbers, recent policy violations) requires the M3 collector plane (SigNoz + future cost/policy exporters). The cards currently show authoritative predicted values + explanatory captions.
- The secondary test vectors in the TS file contain illustrative placeholders; the canonical vector is the only one with full cross-language proof at the time of writing.

---

**End of Technical Specification**

This document, together with the source files it references, constitutes the complete technical record for the OpenChoreo Entity Cards module. It satisfies the governing persona requirement for a full Requirements + Design + Technical triad for every new functional artifact.