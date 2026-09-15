package propagation

import (
	"fmt"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/identity"
)

// DependencyRule defines an upstream project requirement that propagates into downstream market opportunities.
type DependencyRule struct {
	Sector           domain.Sector
	SubsectorKeyword string
	Category         string
	TitleTemplate    string
	Description      string
	Class            domain.RequirementClass
	TriggerMilestone domain.LifecycleStage
	CapexPercentMin  float64
	CapexPercentMax  float64
}

// DefaultOntology returns the canonical Canadian economic dependency rules across major infrastructure sectors.
func DefaultOntology() []DependencyRule {
	return []DependencyRule{
		// Critical Minerals & Mining
		{
			Sector:           domain.SectorCriticalMinerals,
			Category:         "transmission_and_power",
			TitleTemplate:    "High-Voltage Transmission & On-Site Substation Interconnect",
			Description:      "Dedicated grid interconnection or clean power generation (e.g. SMR, hydro tie-in, microgrid) to energize mill and extraction fleet.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageFeasibility,
			CapexPercentMin:  0.05,
			CapexPercentMax:  0.12,
		},
		{
			Sector:           domain.SectorCriticalMinerals,
			Category:         "civil_and_roads",
			TitleTemplate:    "All-Weather Access Road & Heavy Haul Corridors",
			Description:      "Construction of industrial transport roads, bridges, and culverts linking deposit to provincial highway network or rail heads.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StagePermitting,
			CapexPercentMin:  0.03,
			CapexPercentMax:  0.08,
		},
		{
			Sector:           domain.SectorCriticalMinerals,
			Category:         "epcm_and_engineering",
			TitleTemplate:    "EPCM Detailed Engineering & Project Management Contract",
			Description:      "Turnkey Engineering, Procurement, and Construction Management (EPCM) package for processing mill and tailings management facility.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageFIDLikely,
			CapexPercentMin:  0.08,
			CapexPercentMax:  0.15,
		},
		{
			Sector:           domain.SectorCriticalMinerals,
			Category:         "environmental_monitoring",
			TitleTemplate:    "Long-Term Environmental Baseline & Water Quality Monitoring",
			Description:      "Aquatic habitat monitoring, acid rock drainage surveillance, and Indigenous community guardian reporting systems.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageEnvironmentalReview,
			CapexPercentMin:  0.01,
			CapexPercentMax:  0.03,
		},
		{
			Sector:           domain.SectorCriticalMinerals,
			Category:         "worker_accommodation",
			TitleTemplate:    "Remote Workforce Lodging & Modular Camp Infrastructure",
			Description:      "Turnkey 500-1,500 bed modular camp accommodation, catering, medical clinic, and satellite communications.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StagePermitting,
			CapexPercentMin:  0.02,
			CapexPercentMax:  0.05,
		},
		{
			Sector:           domain.SectorCriticalMinerals,
			Category:         "battery_supply_chain",
			TitleTemplate:    "Downstream Precursor Cathode Active Material (pCAM) Integration",
			Description:      "Offtake allocation and specialized chemical refining integration into domestic North American battery supply chains.",
			Class:            domain.RequirementSpeculative,
			TriggerMilestone: domain.StageFID,
			CapexPercentMin:  0.10,
			CapexPercentMax:  0.25,
		},

		// Nuclear & Clean Power
		{
			Sector:           domain.SectorNuclearEnergy,
			Category:         "nuclear_grade_components",
			TitleTemplate:    "CSA N285 / N299 Certified Pressure Vessel & Piping Fabrication",
			Description:      "Nuclear-quality certified valves, calandria components, heat exchangers, and primary heat transport piping.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageEarlyDevelopment,
			CapexPercentMin:  0.15,
			CapexPercentMax:  0.25,
		},
		{
			Sector:           domain.SectorNuclearEnergy,
			Category:         "instrumentation_and_control",
			TitleTemplate:    "Safety-Critical Digital Instrumentation & Control (I&C) Architecture",
			Description:      "Triple-redundant control systems, reactor protection logic, cybersecurity monitoring, and physical simulator suites.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StagePermitting,
			CapexPercentMin:  0.05,
			CapexPercentMax:  0.10,
		},
		{
			Sector:           domain.SectorNuclearEnergy,
			Category:         "heavy_civil_works",
			TitleTemplate:    "Nuclear Containment Heavy Civil & Specialized Pour Concrete",
			Description:      "Specialized seismic foundation, heavy aggregate shielding concrete, and intake/outfall cooling water structures.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageFID,
			CapexPercentMin:  0.12,
			CapexPercentMax:  0.20,
		},

		// AI Compute & Data Centres
		{
			Sector:           domain.SectorAICompute,
			Category:         "power_interconnection",
			TitleTemplate:    "Multi-Hundred Megawatt High-Voltage Interconnection & Substation",
			Description:      "Dedicated 230kV/500kV redundant utility substation, switchgear, and on-site utility firm capacity allocation.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageEarlyDevelopment,
			CapexPercentMin:  0.08,
			CapexPercentMax:  0.15,
		},
		{
			Sector:           domain.SectorAICompute,
			Category:         "liquid_cooling",
			TitleTemplate:    "Direct-to-Chip Liquid Cooling Distribution & Thermal Dissipation",
			Description:      "High-density coolant distribution units (CDUs), closed-loop water chillers, and zero-water dissipation systems.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StagePermitting,
			CapexPercentMin:  0.10,
			CapexPercentMax:  0.18,
		},
		{
			Sector:           domain.SectorAICompute,
			Category:         "sovereign_cybersecurity",
			TitleTemplate:    "Sovereign Perimeter Defense, PBMM Encryption & Air-Gap Architecture",
			Description:      "Protected B / Medium / Medium (PBMM) federal compliance, hardware security modules (HSM), and zero-trust orchestration.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageProcurement,
			CapexPercentMin:  0.02,
			CapexPercentMax:  0.05,
		},

		// Defence & Arctic
		{
			Sector:           domain.SectorDefenceArctic,
			Category:         "permafrost_civil_engineering",
			TitleTemplate:    "Arctic Thermosyphon Foundation & Cold-Weather Runway Hardening",
			Description:      "Passive refrigeration thermosyphon foundations, all-weather gravel aggregate stabilization, and insulated utility corridors.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageEarlyDevelopment,
			CapexPercentMin:  0.12,
			CapexPercentMax:  0.22,
		},
		{
			Sector:           domain.SectorDefenceArctic,
			Category:         "surveillance_and_c4isr",
			TitleTemplate:    "Over-The-Horizon Sensor Interconnect & Secure Communications",
			Description:      "LEO satellite uplinks, secure encrypted tactical datalinks, and radar power generation resilience.",
			Class:            domain.RequirementDerived,
			TriggerMilestone: domain.StageProcurement,
			CapexPercentMin:  0.15,
			CapexPercentMax:  0.25,
		},

		// Transportation & Ports
		{
			Sector:           domain.SectorTransportation,
			Category:         "marine_deepwater_berths",
			TitleTemplate:    "Marine Terminal Dredging & Deep-Water Wharf Construction",
			Description:      "Geotechnical marine dredging, sheet pile berthing walls, and high-capacity gantry crane rail installations.",
			Class:            domain.RequirementConfirmed,
			TriggerMilestone: domain.StageEnvironmentalReview,
			CapexPercentMin:  0.15,
			CapexPercentMax:  0.30,
		},
	}
}

// PropagateOpportunities evaluates a project against the dependency ontology and generates downstream opportunities.
func PropagateOpportunities(project *domain.Project) []*domain.Opportunity {
	rules := DefaultOntology()
	var results []*domain.Opportunity

	for _, rule := range rules {
		if rule.Sector != project.Sector {
			continue
		}

		opp := &domain.Opportunity{
			ID:               identity.StableID("opportunity", "dependency-template-v1", project.ID+":"+rule.Category),
			ProjectID:        project.ID,
			ProjectName:      project.Name,
			Title:            fmt.Sprintf("%s: %s", project.Name, rule.TitleTemplate),
			Sector:           project.Sector,
			RequirementClass: rule.Class,
			Category:         rule.Category,
			EstimatedCAD:     0,
			EstimateStatus:   domain.ConfidenceUnknown,
			Description:      rule.Description,
			TriggerMilestone: string(rule.TriggerMilestone),
			OpportunityKind:  "PROCUREMENT",
			Visibility:       domain.VisibilityPublic,
			Publishable:      true,
			PublicationState: domain.PublicationPublicCanonical,
			CreatedAt:        project.UpdatedAt,
		}
		results = append(results, opp)
	}

	return results
}
