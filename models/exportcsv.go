package models

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func matchFilter(campaignName string, filterTerms []string) bool {
	for _, filter := range filterTerms {
		if !strings.Contains(campaignName, filter) {
			return false
		}
	}
	return true
}

func MatchingCampaigns(w http.ResponseWriter, r *http.Request, userId int64) []Campaign {
	filter := ""
	matchExact := "off"
	filter = r.URL.Query().Get("filter")
	if r.URL.Query().Get("matchExact") != "" {
		matchExact = r.URL.Query().Get("matchExact")
	}
	filterTerms := []string{}
	if matchExact == "on" {
		filterTerms = []string{filter}
	} else {
		filterTerms = strings.Split(filter, " ")
	}
	allCampaigns, err := GetCampaigns(userId, false)
	if err != nil {
		Logger.Println(err)
	}
	cs := []Campaign{}
	cs = allCampaigns
	csFiltered := []Campaign{}
	for _, c := range cs {
		if matchFilter(c.Name, filterTerms) || len(filterTerms) == 0 {
			if err := c.GetDetails(true, true, true, true); err != nil {
				Logger.Println(err)
			} else {
				csFiltered = append(csFiltered, c)
			}
		}
	}
	return csFiltered
}

func ExportCsv(w http.ResponseWriter, r *http.Request, userId int64) {
	campaigns := MatchingCampaigns(w, r, userId)
	csv := ""
	sent := 0
	opened := 0
	clicked := 0
	phishSuccess := 0
	submittedData := 0
	csvResults := "\n"
	csvCampaignNames := ""
	fmt.Println("EXPORT CSV - NUM CAMPAIGNS:", len(campaigns))
	for i, campaign := range campaigns {
		fmt.Println("EXPORT CSV:", i, len(campaigns))
		csvCampaignNames += campaign.Name + " (" + strconv.Itoa(int(campaign.Id)) + ")" + ","
		for _, result := range campaign.Results {
			csvResult := strconv.Itoa(int(result.CampaignId)) + "," +
				result.FirstName + "," +
				result.LastName + "," +
				result.Email + ", " +
				result.Position + ", " +
				result.Department
			csvResultStatus := ""
			switch result.Status {
			case "Email Sent":
				sent++
				csvResultStatus = "Sent"
				break
			case "Email Opened":
				sent++
				opened++
				csvResultStatus = "Opened"
				break
			case "Clicked Link":
				sent++
				opened++
				clicked++
				csvResultStatus = "Clicked"
				break
			case "Success":
				sent++
				opened++
				clicked++
				phishSuccess++
				break
			}
			for _, event := range campaign.Events {
				if event.Email == result.Email && event.Message == "Submitted Data" {
					submittedData++
					csvResultStatus = "Submitted Credentials"
					break
				}
			}
			csvResult = csvResult + ", " + csvResultStatus
			csvResults += csvResult + "\n"
		}
	}
	csv += csvCampaignNames + "\n\n"
	csv += csvResults
	w.Header().Add("Content-Type", "application/csv")
	w.Header().Add("Content-Disposition", "attachment; filename=Export.csv")
	w.Write([]byte(csv))
}
