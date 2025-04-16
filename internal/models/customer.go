package models

import "time"

// Customer represents a customer in the POS system
type Customer struct {
	ID                int       `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Address           string    `json:"address"`
	JoinDate          time.Time `json:"join_date"`
	LastPurchaseDate  time.Time `json:"last_purchase_date"`
	TotalPurchases    float64   `json:"total_purchases"`
	Notes             string    `json:"notes"`
	LoyaltyPoints     int       `json:"loyalty_points"`
	LoyaltyTier       string    `json:"loyalty_tier"`
	Birthday          string    `json:"birthday"`
	PreferredProducts string    `json:"preferred_products"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// CustomerSummary provides a simplified view of customer data
type CustomerSummary struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	LoyaltyPoints int    `json:"loyalty_points"`
	LoyaltyTier   string `json:"loyalty_tier"`
}

// LoyaltyTier defines a tier in the loyalty program
type LoyaltyTier struct {
	ID                 int     `json:"id"`
	Name               string  `json:"name"`
	MinPoints          int     `json:"min_points"`
	DiscountPercentage float64 `json:"discount_percentage"`
	PointsMultiplier   float64 `json:"points_multiplier"`
	Benefits           string  `json:"benefits"`
}

// LoyaltyReward defines a reward in the loyalty program
type LoyaltyReward struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	PointsCost    int     `json:"points_cost"`
	DiscountValue float64 `json:"discount_value"`
	IsPercentage  bool    `json:"is_percentage"`
	ValidDays     int     `json:"valid_days"`
	Active        bool    `json:"active"`
}

// CustomerSale links a sale to a customer
type CustomerSale struct {
	SaleID      int       `json:"sale_id"`
	CustomerID  int       `json:"customer_id"`
	PointsEarned int       `json:"points_earned"`
	PointsUsed   int       `json:"points_used"`
	RewardID     int       `json:"reward_id"`
	CreatedAt    time.Time `json:"created_at"`
}

// CalculatePointsForPurchase determines how many loyalty points to award for a purchase
func CalculatePointsForPurchase(amount float64, multiplier float64) int {
	// Base calculation: $1 = 1 point, multiplied by tier multiplier
	if multiplier <= 0 {
		multiplier = 1.0
	}
	
	points := int(amount * multiplier)
	return points
}

// GetLoyaltyTierName returns the name of the loyalty tier based on points
func GetLoyaltyTierName(points int) string {
	// Default tier names, would typically come from database
	if points >= 1000 {
		return "Platinum"
	} else if points >= 500 {
		return "Gold"
	} else if points >= 200 {
		return "Silver"
	} else {
		return "Bronze"
	}
}

// GetLoyaltyTierDiscount returns the discount percentage for a loyalty tier
func GetLoyaltyTierDiscount(tierName string) float64 {
	// Default tier discounts, would typically come from database
	switch tierName {
	case "Platinum":
		return 0.15 // 15% discount
	case "Gold":
		return 0.10 // 10% discount
	case "Silver":
		return 0.05 // 5% discount
	default:
		return 0.0 // No discount for bronze
	}
}

// GetLoyaltyTierMultiplier returns the points multiplier for a loyalty tier
func GetLoyaltyTierMultiplier(tierName string) float64 {
	// Default tier multipliers, would typically come from database
	switch tierName {
	case "Platinum":
		return 2.0 // 2x points
	case "Gold":
		return 1.5 // 1.5x points
	case "Silver":
		return 1.2 // 1.2x points
	default:
		return 1.0 // 1x points for bronze
	}
}