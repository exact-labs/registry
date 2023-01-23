package helpers

import "github.com/pocketbase/pocketbase/models"

func PackagePrivacyStatus(record *models.Record) bool {
   if record.GetString("visibility") == "private" {
      return true
   }

   return false
}

func PackageHasLicense(record *models.Record) string {
   if record.GetString("license") == "" {
      return "none"
   }

   return record.GetString("license")
}