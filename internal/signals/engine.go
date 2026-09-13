package signals

import (
	"fmt"
	"time"

	"github.com/Hardonian/CEO-G-Canada-Economic-Opportunity-Graph/internal/domain"
	"github.com/google/uuid"
)

// MomentumReport computes short-, medium-, and long-term activity velocities for a project.
type MomentumReport struct {
	ProjectID   string    `json:"project_id"`
	ProjectName string    `json:"project_name"`
	Momentum7d  float64   `json:"momentum_7d"`  // -1.0 to +1.0
	Momentum30d float64   `json:"momentum_30d"` // -1.0 to +1.0
	Momentum90d float64   `json:"momentum_90d"` // -1.0 to +1.0
	SignalCount int       `json:"signal_count"`
	Signals     []*domain.Signal `json:"signals"`
	Velocity    string    `json:"velocity"`     // "ACCELERATING", "STEADY", "DECELERATING", "STALLED"
}

// DetectSignals inspects project history and produces typed economic signals.
func DetectSignals(project *domain.Project, events []*domain.Event, capital []*domain.CapitalItem, procurements []*domain.Procurement) []*domain.Signal {
	var signals []*domain.Signal

	// 1. Regulatory progress & stage progression
	for _, ev := range events {
		if ev.NewStage != nil && ev.PreviousStage != nil {
			if *ev.NewStage == domain.StageFID || *ev.NewStage == domain.StageConstruction {
				signals = append(signals, &domain.Signal{
					ID:            uuid.New().String(),
					ProjectID:     project.ID,
					ProjectName:   project.Name,
					Type:          domain.SignalConstructionSignal,
					Timestamp:     ev.EventDate,
					Magnitude:     0.95,
					Confidence:    0.98,
					PreviousState: string(*ev.PreviousStage),
					NewState:      string(*ev.NewStage),
					Description:   fmt.Sprintf("Major milestone achieved: Transitioned from %s to %s.", *ev.PreviousStage, *ev.NewStage),
					EvidenceID:    ev.EvidenceID,
				})
			} else if *ev.NewStage == domain.StagePermitting || *ev.NewStage == domain.StageEnvironmentalReview {
				signals = append(signals, &domain.Signal{
					ID:            uuid.New().String(),
					ProjectID:     project.ID,
					ProjectName:   project.Name,
					Type:          domain.SignalRegulatoryProgress,
					Timestamp:     ev.EventDate,
					Magnitude:     0.70,
					Confidence:    0.92,
					PreviousState: string(*ev.PreviousStage),
					NewState:      string(*ev.NewStage),
					Description:   fmt.Sprintf("Regulatory milestone reached: entered %s.", *ev.NewStage),
					EvidenceID:    ev.EvidenceID,
				})
			}
		}

		if ev.EventType == "indigenous_agreement" || ev.EventType == "impact_benefit_agreement" {
			signals = append(signals, &domain.Signal{
				ID:            uuid.New().String(),
				ProjectID:     project.ID,
				ProjectName:   project.Name,
				Type:          domain.SignalIndigenousPartnership,
				Timestamp:     ev.EventDate,
				Magnitude:     0.88,
				Confidence:    0.95,
				Description:   fmt.Sprintf("Indigenous economic partnership or mutual benefit agreement confirmed: %s", ev.Title),
				EvidenceID:    ev.EvidenceID,
			})
		}
	}

	// 2. Capital & Financing acceleration
	for _, c := range capital {
		if c.Status == domain.CapitalCommitted || c.Status == domain.CapitalClosed {
			signals = append(signals, &domain.Signal{
				ID:          uuid.New().String(),
				ProjectID:   project.ID,
				ProjectName: project.Name,
				Type:        domain.SignalFinancingAcceleration,
				Timestamp:   c.CreatedAt,
				Magnitude:   0.85,
				Confidence:  0.95,
				NewState:    string(c.Status),
				Description: fmt.Sprintf("Capital commitment: $%d CAD secured via %s from %s.", c.AmountCAD, c.Category, c.ProviderName),
				EvidenceID:  c.EvidenceID,
			})
		}
	}

	// 3. Procurement acceleration
	if len(procurements) > 0 {
		signals = append(signals, &domain.Signal{
			ID:          uuid.New().String(),
			ProjectID:   project.ID,
			ProjectName: project.Name,
			Type:        domain.SignalProcurementAcceleration,
			Timestamp:   time.Now(),
			Magnitude:   0.75,
			Confidence:  0.90,
			Description: fmt.Sprintf("Active procurement pipeline: %d tender notices published.", len(procurements)),
		})
	}

	return signals
}

// CalculateMomentum evaluates detected signals across 7d, 30d, and 90d rolling windows.
func CalculateMomentum(project *domain.Project, signals []*domain.Signal) *MomentumReport {
	now := time.Now()
	w7 := now.Add(-7 * 24 * time.Hour)
	w30 := now.Add(-30 * 24 * time.Hour)
	w90 := now.Add(-90 * 24 * time.Hour)

	var score7, score30, score90 float64

	for _, s := range signals {
		val := s.Magnitude
		if s.Type == domain.SignalTimelineSlip || s.Type == domain.SignalProjectDelay || s.Type == domain.SignalPoliticalSupportLoss {
			val = -val
		}

		if s.Timestamp.After(w7) {
			score7 += val * 0.5
		}
		if s.Timestamp.After(w30) {
			score30 += val * 0.3
		}
		if s.Timestamp.After(w90) {
			score90 += val * 0.2
		}
	}

	// Bound to [-1.0, 1.0]
	clampM := func(v float64) float64 {
		if v > 1.0 {
			return 1.0
		}
		if v < -1.0 {
			return -1.0
		}
		return v
	}

	m7 := clampM(score7)
	m30 := clampM(score30)
	m90 := clampM(score90)

	velocity := "STEADY"
	if m30 > 0.4 {
		velocity = "ACCELERATING"
	} else if m30 < -0.2 {
		velocity = "DECELERATING"
	}

	return &MomentumReport{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Momentum7d:  m7,
		Momentum30d: m30,
		Momentum90d: m90,
		SignalCount: len(signals),
		Signals:     signals,
		Velocity:    velocity,
	}
}
