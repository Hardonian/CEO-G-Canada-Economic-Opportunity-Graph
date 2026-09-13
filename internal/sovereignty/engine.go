package sovereignty

import (
	"fmt"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
)

const VersionAISovereignty = "cai-sovereignty-v1.0"

// AIProfile defines the technical, legal, and operational attributes of an AI model, platform, or compute facility.
type AIProfile struct {
	SubjectID          string
	SubjectName        string
	DataResidencyCA    bool    // 100% data at rest stored within Canadian borders
	ComputeResidencyCA bool    // GPUs / TPUs physically executing within Canada
	CanadianOwnership  float64 // 0.0 to 1.0 equity / control
	ForeignLegalRisk   float64 // 0.0 to 1.0 (1.0 = zero foreign CLOUD Act exposure, 0.0 = total exposure)
	LocalDeployment    bool    // Can be deployed on-premise without phoning home
	OfflineCapability  bool    // Operates in air-gapped / disconnected mode
	OpenWeights        bool    // Weights published under permissive or open license
	BilingualCapacity  float64 // 0.0 to 1.0 native Canadian English & French parity
	QuebecLaw25Ready   bool    // Strict default consent, DPIA, right to erasure implemented
	CleanEnergySource  float64 // 0.0 to 1.0 clean hydro / nuclear / non-emitting power share
}

// EvaluateSovereignty computes the Canadian AI Sovereignty Index (0-100) deterministically.
func EvaluateSovereignty(profile *AIProfile) *domain.AISovereignty {
	dims := make(map[string]float64)

	// 1. Data Residency (15%)
	dataScore := 10.0
	if profile.DataResidencyCA {
		dataScore = 100.0
	}
	dims["data_residency"] = dataScore

	// 2. Compute Residency (15%)
	computeScore := 10.0
	if profile.ComputeResidencyCA {
		computeScore = 100.0
	}
	dims["compute_residency"] = computeScore

	// 3. Ownership & Corporate Governance (15%)
	ownScore := profile.CanadianOwnership * 100.0
	dims["canadian_ownership"] = ownScore

	// 4. Foreign Extraterritorial Legal Exposure (15%)
	// (Assesses immunity from US CLOUD Act, Foreign Intelligence Surveillance Act sec 702)
	foreignScore := profile.ForeignLegalRisk * 100.0
	dims["foreign_legal_protection"] = foreignScore

	// 5. Local & Air-Gapped Deployment Autonomy (15%)
	localScore := 20.0
	if profile.LocalDeployment {
		localScore += 40.0
	}
	if profile.OfflineCapability {
		localScore += 40.0
	}
	dims["deployment_autonomy"] = localScore

	// 6. Linguistic Sovereignty & Bilingual Parity (10%)
	bilingualScore := profile.BilingualCapacity * 100.0
	dims["bilingual_capacity"] = bilingualScore

	// 7. Privacy & Quebec Law 25 Compliance Posture (10%)
	privScore := 30.0
	if profile.QuebecLaw25Ready {
		privScore = 100.0
	}
	dims["privacy_law25"] = privScore

	// 8. Clean Energy Provenance (5%)
	cleanScore := profile.CleanEnergySource * 100.0
	dims["clean_energy"] = cleanScore

	overall := (dims["data_residency"] * 0.15) +
		(dims["compute_residency"] * 0.15) +
		(dims["canadian_ownership"] * 0.15) +
		(dims["foreign_legal_protection"] * 0.15) +
		(dims["deployment_autonomy"] * 0.15) +
		(dims["bilingual_capacity"] * 0.10) +
		(dims["privacy_law25"] * 0.10) +
		(dims["clean_energy"] * 0.05)

	if overall > 100.0 {
		overall = 100.0
	}

	notes := fmt.Sprintf("Calculated under %s: Data residency (%.0f), Compute residency (%.0f), Ownership (%.0f), Foreign immunity (%.0f), Autonomy (%.0f), Bilingualism (%.0f), and Law 25 (%.0f). DISCLAIMER: Research score only; not legal or compliance advice.",
		VersionAISovereignty, dims["data_residency"], dims["compute_residency"], dims["canadian_ownership"], dims["foreign_legal_protection"], dims["deployment_autonomy"], dims["bilingual_capacity"], dims["privacy_law25"])

	return &domain.AISovereignty{
		ScoreVersion:      VersionAISovereignty,
		OverallScore:      round(overall),
		DataResidency:     dims["data_residency"] / 10.0,
		ComputeResidency:  dims["compute_residency"] / 10.0,
		CanadianOwnership: dims["canadian_ownership"] / 10.0,
		ForeignLegalRisk:  dims["foreign_legal_protection"] / 10.0,
		LocalDeployment:   dims["deployment_autonomy"] / 10.0,
		BilingualCapacity: dims["bilingual_capacity"] / 10.0,
		QuebecLaw25:       dims["privacy_law25"] / 10.0,
		CleanEnergy:       dims["clean_energy"] / 10.0,
		Dimensions:        dims,
		Notes:             notes,
		CalculatedAt:      time.Now(),
	}
}

func round(val float64) float64 {
	return float64(int(val*10)) / 10.0
}
