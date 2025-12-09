package jobs

import (
	"lab02_mahoa/server/database"
	"lab02_mahoa/server/models"
	"log"
	"time"

	"gorm.io/gorm"
)

// StartCleanupJob starts a background job to clean up expired shares and links
func StartCleanupJob(db *gorm.DB) {
	log.Println("🧹 Starting cleanup job for expired shares and links...")

	// Run cleanup every hour
	ticker := time.NewTicker(1 * time.Hour)
	
	// Run immediately on start
	go cleanupExpiredData(db)

	// Then run periodically
	go func() {
		for range ticker.C {
			cleanupExpiredData(db)
		}
	}()
}

// cleanupExpiredData removes expired shared links and E2EE shares
func cleanupExpiredData(db *gorm.DB) {
	now := time.Now()
	
	// Clean up expired shared links
	var expiredLinks []models.SharedLink
	result := db.Where("expires_at < ?", now).Find(&expiredLinks)
	if result.Error != nil {
		log.Printf("❌ Error finding expired links: %v", result.Error)
	} else if len(expiredLinks) > 0 {
		// Delete expired links
		deleteResult := db.Where("expires_at < ?", now).Delete(&models.SharedLink{})
		if deleteResult.Error != nil {
			log.Printf("❌ Error deleting expired links: %v", deleteResult.Error)
		} else if deleteResult.RowsAffected > 0 {
			log.Printf("🧹 Cleaned up %d expired shared links", deleteResult.RowsAffected)
		}
	}

	// Clean up expired E2EE shares
	var expiredShares []models.E2EEShare
	result = db.Where("expires_at < ?", now).Find(&expiredShares)
	if result.Error != nil {
		log.Printf("❌ Error finding expired E2EE shares: %v", result.Error)
	} else if len(expiredShares) > 0 {
		// Delete expired shares
		deleteResult := db.Where("expires_at < ?", now).Delete(&models.E2EEShare{})
		if deleteResult.Error != nil {
			log.Printf("❌ Error deleting expired E2EE shares: %v", deleteResult.Error)
		} else if deleteResult.RowsAffected > 0 {
			log.Printf("🧹 Cleaned up %d expired E2EE shares", deleteResult.RowsAffected)
		}
	}

	// Clean up exhausted shared links (access_count >= max_access_count)
	var exhaustedLinks []models.SharedLink
	result = db.Where("max_access_count > 0 AND access_count >= max_access_count").Find(&exhaustedLinks)
	if result.Error != nil {
		log.Printf("❌ Error finding exhausted links: %v", result.Error)
	} else if len(exhaustedLinks) > 0 {
		// Delete exhausted links
		deleteResult := db.Where("max_access_count > 0 AND access_count >= max_access_count").Delete(&models.SharedLink{})
		if deleteResult.Error != nil {
			log.Printf("❌ Error deleting exhausted links: %v", deleteResult.Error)
		} else if deleteResult.RowsAffected > 0 {
			log.Printf("🧹 Cleaned up %d exhausted shared links", deleteResult.RowsAffected)
		}
	}
}

// CleanupExpiredDataNow immediately cleans up expired data (for manual trigger)
func CleanupExpiredDataNow() {
	db := database.GetDB()
	if db == nil {
		log.Println("❌ Database not initialized")
		return
	}
	log.Println("🧹 Manual cleanup triggered...")
	cleanupExpiredData(db)
}
