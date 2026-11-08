#!/usr/bin/env python3
"""
Conclusive residual ranking for the developer-portal carrier.
Requires: scikit-learn, numpy

Candidate analytical procedure only.
Does not authorize any mutation of repositories, CI, or live systems.
"""

from __future__ import annotations

import math
from dataclasses import dataclass, field
from typing import Any, Dict, List, Sequence

import numpy as np
from sklearn.ensemble import IsolationForest

# ---------------------------------------------------------------------------
# Locked calibrated parameters
# ---------------------------------------------------------------------------

ALPHA, BETA, GAMMA = 0.55, 0.30, 0.15
K1 = 1.2
AGE_WEIGHT = 1.0
LAMBDA_BOOST = 0.40
WINSOR_LO, WINSOR_HI = 5.0, 95.0
A_CAP = 0.95

# ---------------------------------------------------------------------------
# Math helpers
# ---------------------------------------------------------------------------

def bm25_tf(o: float, k1: float = K1) -> float:
    o = max(0.0, float(o))
    return (o * (k1 + 1.0)) / (o + k1)

def normalize(vals: Sequence[float]) -> List[float]:
    peak = max(vals) or 1.0
    return [v / peak for v in vals]

def hybrid_severity(sm: float, sig: float, f: float) -> float:
    return ALPHA * (1.0 - sm) + BETA * sig + GAMMA * f

def surface_priority(s_gap: float, oga: float) -> float:
    return s_gap * (1.0 + AGE_WEIGHT * math.log1p(max(0.0, oga)))

def winsorized_minmax(vals: Sequence[float]) -> np.ndarray:
    arr = np.asarray(vals, dtype=float)
    lo, hi = np.percentile(arr, [WINSOR_LO, WINSOR_HI])
    clipped = np.clip(arr, lo, hi)
    if hi - lo < 1e-12:
        return np.zeros_like(arr)
    a = (clipped - lo) / (hi - lo)
    return np.minimum(a, A_CAP)

# ---------------------------------------------------------------------------
# Residual container
# ---------------------------------------------------------------------------

@dataclass
class ResidualItem:
    id: str
    content: str
    sm_forward: float
    specificity: float
    occurrence: float
    oga: float
    mechanism_id: str = ""
    f_bm25: float = 0.0
    s_gap: float = 0.0
    surface_priority: float = 0.0
    anomaly_score: float = 0.0
    attention_score: float = 0.0
    boundary: str = "candidate analytical ranking only — no mutation authorized"

# ---------------------------------------------------------------------------
# Pipeline
# ---------------------------------------------------------------------------

def rank_and_score(raw: List[Dict[str, Any]]) -> List[ResidualItem]:
    # 1. Hybrid severity + SurfacePriority
    occ = [float(r["occurrence"]) for r in raw]
    f_vals = normalize([bm25_tf(o) for o in occ])

    items: List[ResidualItem] = []
    for r, f in zip(raw, f_vals):
        sm = max(0.0, min(1.0, float(r["sm_forward"])))
        sig = max(0.0, min(1.0, float(r["specificity"])))
        oga = max(0.0, float(r["oga"]))
        s_gap = hybrid_severity(sm, sig, f)
        sp = surface_priority(s_gap, oga)
        items.append(
            ResidualItem(
                id=r["id"],
                content=r["content"],
                sm_forward=sm,
                specificity=sig,
                occurrence=float(r["occurrence"]),
                oga=oga,
                mechanism_id=r.get("mechanism_id", ""),
                f_bm25=f,
                s_gap=s_gap,
                surface_priority=sp,
            )
        )

    # 2. Isolation Forest on [S_gap, log1p(OGA), σ]
    X = np.array([[it.s_gap, math.log1p(it.oga), it.specificity] for it in items])
    clf = IsolationForest(
        n_estimators=200,
        contamination="auto",
        random_state=42,
        n_jobs=1,
    )
    clf.fit(X)
    raw_scores = clf.score_samples(X)          # higher = more normal
    inverted = (-raw_scores).tolist()
    A = winsorized_minmax(inverted)

    # 3. AttentionScore = SurfacePriority × (1 + λ A)
    for it, a in zip(items, A):
        it.anomaly_score = float(a)
        it.attention_score = it.surface_priority * (1.0 + LAMBDA_BOOST * it.anomaly_score)

    items.sort(key=lambda r: (-r.attention_score, -r.anomaly_score, -r.surface_priority, r.id))
    return items

# ---------------------------------------------------------------------------
# Concrete portal residual archetypes
# ---------------------------------------------------------------------------

PORTAL_RESIDUALS = [
    {
        "id": "res-proxy-404",
        "content": "frontend proxy missing → 404 while backend reports up",
        "sm_forward": 0.55,
        "specificity": 0.88,
        "occurrence": 4,
        "oga": 11,
        "mechanism_id": "KIMI-MECH-004",
    },
    {
        "id": "res-smoke-divergence",
        "content": "smoke-all failed after production-ready claim",
        "sm_forward": 0.58,
        "specificity": 0.82,
        "occurrence": 3,
        "oga": 6,
        "mechanism_id": "KIMI-MECH-004",
    },
    {
        "id": "res-cert-drift",
        "content": "clientCA / certificate freshness mismatch across restarts",
        "sm_forward": 0.55,
        "specificity": 0.80,
        "occurrence": 2,
        "oga": 18,
        "mechanism_id": "KIMI-MECH-004",
    },
    {
        "id": "res-scaffold-409",
        "content": "scaffolder name collision returned 409",
        "sm_forward": 0.70,
        "specificity": 0.72,
        "occurrence": 2,
        "oga": 1,
        "mechanism_id": "KIMI-MECH-002",
    },
    {
        "id": "res-generic-unresolved",
        "content": "unresolved",
        "sm_forward": 0.80,
        "specificity": 0.22,
        "occurrence": 9,
        "oga": 4,
        "mechanism_id": "KIMI-MECH-001",
    },
]

# ---------------------------------------------------------------------------
# Consumers
# ---------------------------------------------------------------------------

def readiness_gate(ranked: List[ResidualItem], top_k: int = 5, min_att: float = 0.55) -> Dict:
    surface = [r for r in ranked if r.attention_score >= min_att][:top_k]
    return {
        "boundary": "candidate analytical residual surface — no mutation authorized",
        "gate_status": "WARN" if surface else "CLEAR",
        "residuals": [
            {
                "id": r.id,
                "attention": round(r.attention_score, 4),
                "A": round(r.anomaly_score, 3),
                "S_gap": round(r.s_gap, 4),
                "OGA": r.oga,
                "content": r.content,
            }
            for r in surface
        ],
    }

def restart_checklist(ranked: List[ResidualItem], max_items: int = 5, min_oga: float = 3.0) -> Dict:
    aged = [r for r in ranked if r.oga >= min_oga and r.attention_score > 0.4][:max_items]
    return {
        "boundary": "candidate restart checklist seed — no mutation authorized",
        "items": [
            f"[OGA={r.oga:.0f}] {r.content}  (Attention={r.attention_score:.3f})"
            for r in aged
        ],
    }

# ---------------------------------------------------------------------------
# Execute
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    print("=" * 78)
    print("CONCLUSIVE RESIDUAL RANKING — Hybrid + Isolation Forest + Winsorized 5-95")
    print("Boundary: candidate analytical ranking only — no mutation authorized")
    print("=" * 78)

    ranked = rank_and_score(PORTAL_RESIDUALS)

    print("\nFinal AttentionScore ranking")
    print(f"{'Rank':<5} {'Attention':>10} {'A':>7} {'SP':>8} {'S_gap':>7} {'σ':>5} {'OGA':>5}  ID")
    print("-" * 78)
    for i, r in enumerate(ranked, 1):
        print(
            f"{i:<5} {r.attention_score:10.4f} {r.anomaly_score:7.3f} "
            f"{r.surface_priority:8.4f} {r.s_gap:7.4f} {r.specificity:5.2f} "
            f"{r.oga:5.1f}  {r.id}"
        )

    gate = readiness_gate(ranked)
    print("\nReadiness residual gate:", gate["gate_status"])
    for r in gate["residuals"]:
        print(f"  • {r['id']:<22} Att={r['attention']:.4f}  A={r['A']:.3f}  OGA={r['OGA']}")

    checklist = restart_checklist(ranked)
    print("\nRestart checklist seeds:")
    for line in checklist["items"]:
        print("  •", line)

    print("\n" + "=" * 78)
    print("All outputs are candidate analytical artifacts only.")
    print("No portal, CI, or repository changes are authorized.")
    print("=" * 78)
