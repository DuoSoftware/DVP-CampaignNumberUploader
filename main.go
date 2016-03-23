package main

import (
	"DVP-CampaignNoUploader/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/DuoSoftware/gorest"
	"github.com/go-contrib/uuid"
	"net/http"
)

var db *sql.DB // global variable to share it between main and the HTTP handler

type HttpRouter struct {
	gorest.RestService    `root:"/DVP/API/6.0/CampaignManager/NumberUpload" consumes:"application/json" produces:"application/json"`
	uploadContacts        gorest.EndPoint `method:"POST" path:"/" postdata:"UploadData"`
	assignContacts        gorest.EndPoint `method:"POST" path:"/{(CampaignId):int}/Numbers/{(CategoryID):int}/Assign" postdata:"ExistingData"`
	removeCompleteProcess gorest.EndPoint `method:"POST" path:"/RemoveComplete" postdata:"ExistingData"`
	trackerData           gorest.EndPoint `method:"GET" path:"/{(TrackerId):string}/Tracker" output:"string"`
	printData             gorest.EndPoint `method:"POST" path:"/Trackers/Data/Print" postdata:"UploadData"`
	/*
			r := mux.NewRouter()
		r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload", &ContextAdapter{ctx, ContextHandlerFunc(models.UploadContactsToCampaignAndAttachSchedule)})
		r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{CampaignId}/Numbers/{CategoryID}/Assign", &ContextAdapter{ctx, ContextHandlerFunc(models.AssingExssitingNumbersToCampaign)})
		r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/RemoveComplete", &ContextAdapter{ctx, ContextHandlerFunc(models.RemoveCompleteProcess)})
		r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{TrackerId}/Tracker", &ContextAdapter{ctx, ContextHandlerFunc(models.TrackNumberUpload)})
		r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/Trackers/Data/Print", &ContextAdapter{ctx, ContextHandlerFunc(models.GetTrackingInfo)})
		http.ListenAndServe(":3268", r)
	*/
}

func main() {
	fmt.Println("starting up")

	var err error
	db, err = sql.Open("postgres", "host=localhost dbname=dvpdb sslmode=disable user=duouser password=DuoS123")
	if err != nil {
		fmt.Println("Error on initializing database connection: %s", err.Error())
	}
	if err = db.Ping(); err != nil {
		return
	}
	gorest.RegisterService(new(HttpRouter))
	http.Handle("/", gorest.Handle())
	http.ListenAndServe(":3268", nil)
}

func GenerateReply(uid uuid.UUID) []byte {
	reply := models.UploadData{}
	reply.TrackerId = uid
	outgoingJSON, err := json.Marshal(reply)
	if err != nil {
		fmt.Println(err.Error())
		return []byte(err.Error())
	}
	return outgoingJSON
}

func (router HttpRouter) UploadContacts(data models.UploadData) {

	uid := uuid.NewV4()
	fmt.Printf("UploadContacts : %s\n", uid)

	go models.UploadContactsToCampaignAndAttachSchedule(db, data, uid)

	router.RB().SetResponseCode(202)
	router.RB().Write([]byte("{ \"Data\": " + string(GenerateReply(uid)) + ",\"IsSuccess\":\"true\"}"))

}

func (router HttpRouter) PrintData(data models.UploadData) {

	go models.GetTrackingInfo()
	router.RB().SetResponseCode(200)
	router.RB().Write([]byte("{ \"msg\":\"Done.\",\"IsSuccess\":\"true\"}"))
	return
}

func (router HttpRouter) AssignContacts(data models.ExistingData, campaignId int, categoryID int) {

	uid := uuid.NewV4()
	fmt.Printf("AssignContacts : %s\n", uid)

	go models.AssingExssitingNumbersToCampaign(db, data, uid, campaignId, categoryID)

	router.RB().SetResponseCode(202)
	router.RB().Write([]byte("{ \"Data\": " + string(GenerateReply(uid)) + ",\"IsSuccess\":\"true\"}"))
	return
}

func (router HttpRouter) RemoveCompleteProcess(data models.ExistingData) {

	fmt.Printf("RemoveCompleteProcess : %s\n")

	reply := models.RemoveCompleteProcess()
	router.RB().SetResponseCode(200)
	router.RB().Write([]byte("{ \"Data\": " + string(reply) + ",\"IsSuccess\":\"true\"}"))
	return
}

func (router HttpRouter) TrackerData(TrackerId string) string {
	uid := uuid.NewV4()
	fmt.Printf("TrackerData : %s\n", uid)

	reply := models.TrackNumberUpload(TrackerId)
	router.RB().SetResponseCode(200)
	//router.RB().Write([]byte("{ \"Data\": " + string(reply) + ",\"IsSuccess\":\"true\"}"))
	return "{ \"Data\": " + string(reply) + ",\"IsSuccess\":\"true\"}"
}
