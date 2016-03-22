// DVP-CampaignNoUploader project main.go

package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload", UploadContactsToCampaignAndAttachSchedule).Methods("POST")
	router.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/RemoveComplete", RemoveCompleteProcess).Methods("POST")
	router.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/TrackerInfo", GetTrackingInfo).Methods("GET")
	router.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/{TrackerId}", TrackNumberUpload).Methods("GET")
	router.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/{CampaignId}/Numbers/Existing", AssingExssitingNumbersToCampaign).Methods("POST")
	http.ListenAndServe(":8080", router)
}
