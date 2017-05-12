package models

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func PdfReport(w http.ResponseWriter, r *http.Request, userId int64) {
	campaignID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		fmt.Println("PdfReport-1:", err)
		return
	}
	campaign, err := GetCampaignResults(int64(campaignID), userId)
	if err != nil {
		fmt.Println("PdfReport-2:", err)
		return
	}
	fmt.Println(campaign)
	pdf := gofpdf.New("L", "in", "Letter", "")
	pdf.SetFont("Arial", "", 12)
	// email,first_name,last_name,position,status,ip,latitude,longitude
	// fake@fake.com,Bob,Mark,Director,Error,,0,0
	header := []string{"First Name", "Last Name", "Position", "Department", "Status"}
	widths := []float64{1.0, 1.0, 1.0, 1.0, 1.5}
	pdf.AddPage()
	h := 0.4
	pdf.SetFont("Arial", "B", 12)

	imageOptions := gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}
	pdf.ImageOptions("./static/images/"+"phishaway-logo.png", 0.1, 0.1, 1.2, 0.5, false, imageOptions, 0, "")

	pdf.CellFormat(0, h, time.Now().Local().Format("Mon Jan 2 15:04:05 -0700 MST 2006"), "", 1, "R", false, 0, "")
	pdf.CellFormat(0, h, "Campaign Results", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, h, campaign.Name+" (#"+strconv.Itoa(campaignID)+")", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	sent := 0
	opened := 0
	clicked := 0
	phishSuccess := 0

	for _, r := range campaign.Results {
		switch r.Status {
		case EVENT_SENT:
			sent++
		case EVENT_OPENED:
			sent++
			opened++
		case EVENT_CLICKED:
			sent++
			opened++
			clicked++
		case STATUS_SUCCESS:
			sent++
			opened++
			clicked++
			phishSuccess++
		}
	}

	summaryWidth := 5.0
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, h, "Summary:", "", 1, "", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(summaryWidth, h, strconv.Itoa(sent)+" emails sent", "1", 0, "", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(summaryWidth, h, strconv.Itoa(phishSuccess)+" users gave away credentials", "1", 0, "", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(summaryWidth, h, trim(summaryWidth, strconv.Itoa(clicked)+" users browsed the phishing page"), "1", 0, "", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(summaryWidth, h, trim(summaryWidth, strconv.Itoa(opened)+" users opened the email"), "1", 0, "", false, 0, "")
	pdf.Ln(-1)

	pdf.Ln(-1)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, h, "Details:", "", 1, "", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	for i, str := range header {
		pdf.CellFormat(widths[i], h, str, "1", 0, "", false, 0, "")
	}
	for i, r := range campaign.Results {
		if (i+1)%13 == 0 {
			pdf.AddPage()
		}
		pdf.Ln(-1)
		pdf.CellFormat(widths[0], h, trim(widths[0], r.FirstName), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[1], h, trim(widths[1], r.LastName), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[2], h, trim(widths[2], r.Position), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[3], h, trim(widths[3], r.Department), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[4], h, trim(widths[4], r.Status), "1", 0, "", false, 0, "")
	}
	err = pdf.OutputFileAndClose("campaign-results.pdf")
	if err != nil {
		fmt.Println("PdfReport-3:", err)
		return
	}
	report, err := ioutil.ReadFile("campaign-results.pdf")
	if err != nil {
		fmt.Println("PdfReport-4:", err)
		return
	}
	w.Write(report)
}

func ExportRawEvents(w http.ResponseWriter, r *http.Request, userId int64) {
	campaignID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		fmt.Println("ExportRawEvents-1:", err)
		return
	}
	campaign, err := GetCampaignResults(int64(campaignID), userId)
	if err != nil {
		fmt.Println("ExportRawEvents-2:", err)
		return
	}
	fmt.Println(campaign)
	pdf := gofpdf.New("L", "in", "Letter", "")
	pdf.SetFont("Arial", "", 12)
	header := []string{"Email", "Time", "Message", "Details"}
	widths := []float64{2.0, 2.0, 3.0, 3.0}
	pdf.AddPage()
	h := 0.4
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, h, "Campaign Raw Events - "+time.Now().Local().String(), "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, h, "Campaign: "+trim(1.0, campaign.Name)+"(#"+strconv.Itoa(campaignID)+")", "", 1, "", false, 0, "")
	pdf.Ln(-1)
	for i, str := range header {
		pdf.CellFormat(widths[i], h, str, "1", 0, "", false, 0, "")
	}
	for i, e := range campaign.Events {
		if (i+1)%10 == 0 {
			pdf.AddPage()
		}
		pdf.Ln(-1)
		pdf.CellFormat(widths[0], h, trim(widths[0], e.Email), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[1], h, trim(widths[1], e.Time.Local().String()), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[2], h, trim(widths[2], e.Message), "1", 0, "", false, 0, "")
		pdf.CellFormat(widths[3], h, trim(widths[3], e.Details), "1", 0, "", false, 0, "")
	}
	err = pdf.OutputFileAndClose("campaign-rawevents.pdf")
	if err != nil {
		fmt.Println("ExportRawEvents-3:", err)
		return
	}
	file, err := ioutil.ReadFile("campaign-rawevents.pdf")
	if err != nil {
		fmt.Println("ExportRawEvents-4:", err)
		return
	}
	w.Write(file)
}

func trim(header float64, s string) string {
	max := int(10 * header)
	if len(s) >= max {
		return s[0:max] + "..."
	} else {
		return s
	}
}
