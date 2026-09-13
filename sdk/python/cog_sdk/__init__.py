"""
CanadaOpportunityGraph Python SDK

A client library for interacting with the CanadaOpportunityGraph API
and working with CEGS (Canada Economic Graph Schema) data.
"""

from .client import COGClient
from .models import (
    Project,
    Entity,
    Event,
    Relationship,
    Procurement,
    CapitalItem,
    Opportunity,
    Signal,
    RadarStats,
    ProjectScore,
    Evidence,
    CapitalProgram,
    StackingEvaluation,
    ReconciliationMatch,
    IndigenousBusiness,
    MerkleProof,
    MerkleTree,
)
from .exceptions import (
    COGError,
    COGNotFoundError,
    COGValidationError,
    COGRateLimitError,
)

__version__ = "0.2.0"
__all__ = [
    "COGClient",
    "Project",
    "Entity",
    "Event",
    "Relationship",
    "Procurement",
    "CapitalItem",
    "Opportunity",
    "Signal",
    "RadarStats",
    "ProjectScore",
    "Evidence",
    "CapitalProgram",
    "StackingEvaluation",
    "ReconciliationMatch",
    "IndigenousBusiness",
    "MerkleProof",
    "MerkleTree",
    "COGError",
    "COGNotFoundError",
    "COGValidationError",
    "COGRateLimitError",
]