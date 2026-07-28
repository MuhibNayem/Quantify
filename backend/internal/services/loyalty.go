package services

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"inventory/backend/internal/domain"
)

func (s *SalesService) handleLoyalty(tx *gorm.DB, customerID *uint, pointsToRedeem int, totalAmount float64) (float64, int, error) {
	if customerID == nil || *customerID == 0 {
		if pointsToRedeem > 0 {
			return 0, 0, fmt.Errorf("cannot redeem points without a customer selected")
		}
		return 0, 0, nil
	}

	var loyalty domain.LoyaltyAccount
	result := tx.Where("user_id = ?", *customerID).First(&loyalty)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			loyalty = domain.LoyaltyAccount{
				UserID: *customerID,
				Points: 0,
				Tier:   "Bronze",
			}
			if err := tx.Create(&loyalty).Error; err != nil {
				return 0, 0, fmt.Errorf("failed to create loyalty account: %w", err)
			}
		} else {
			return 0, 0, fmt.Errorf("failed to fetch loyalty account: %w", result.Error)
		}
	}

	discountAmount := 0.0
	pointsRedeemed := 0

	if pointsToRedeem > 0 {
		if loyalty.Points < pointsToRedeem {
			return 0, 0, fmt.Errorf("insufficient loyalty points (Available: %d, Requested: %d)", loyalty.Points, pointsToRedeem)
		}

		redemptionRate := 0.01
		if val, err := s.Settings.GetSetting("loyalty_points_redemption_rate"); err == nil {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				redemptionRate = v
			}
		}

		discountAmount = float64(pointsToRedeem) * redemptionRate
		if discountAmount > totalAmount {
			return 0, 0, fmt.Errorf("loyalty discount amount (%.2f) exceeds order total (%.2f)", discountAmount, totalAmount)
		}

		loyalty.Points -= pointsToRedeem
		pointsRedeemed = pointsToRedeem
	}

	earningRate, _ := strconv.ParseFloat(getSetting(context.Background(), tx, "loyalty.earning_rate", "0.01"), 64)
	if earningRate == 0 {
		if val, err := s.Settings.GetSetting("loyalty_points_earning_rate"); err == nil {
			if v, err := strconv.ParseFloat(val, 64); err == nil {
				earningRate = v
			}
		}
	}

	netAmount := totalAmount - discountAmount
	if netAmount < 0 {
		netAmount = 0
	}

	pointsEarned := int(netAmount * earningRate)
	loyalty.Points += pointsEarned

	s.updateLoyaltyTier(&loyalty)

	if err := tx.Save(&loyalty).Error; err != nil {
		return 0, 0, fmt.Errorf("failed to update loyalty points: %w", err)
	}

	return discountAmount, pointsRedeemed, nil
}

func (s *SalesService) updateLoyaltyTier(loyalty *domain.LoyaltyAccount) {
	silverThreshold, _ := strconv.Atoi(getSetting(context.Background(), s.DB, "loyalty.tier_silver", "500"))
	goldThreshold, _ := strconv.Atoi(getSetting(context.Background(), s.DB, "loyalty.tier_gold", "2500"))
	platinumThreshold, _ := strconv.Atoi(getSetting(context.Background(), s.DB, "loyalty.tier_platinum", "10000"))

	if loyalty.Points >= platinumThreshold {
		loyalty.Tier = "Platinum"
	} else if loyalty.Points >= goldThreshold {
		loyalty.Tier = "Gold"
	} else if loyalty.Points >= silverThreshold {
		loyalty.Tier = "Silver"
	}
}

func getSetting(ctx context.Context, tx *gorm.DB, key, def string) string {
	var value string
	if err := tx.WithContext(ctx).Model(&domain.SystemSetting{}).Where("key = ?", key).Pluck("value", &value).Error; err != nil || value == "" {
		return def
	}
	return value
}
