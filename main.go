// DVP-CampaignNoUploader project main.go

package main

import (
	"DVP-CampaignNoUploader/models"
	"github.com/gorilla/mux"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

type ContextHandler interface {
	ServeHTTPContext(context.Context, http.ResponseWriter, *http.Request)
}

type ContextHandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

func (h ContextHandlerFunc) ServeHTTPContext(ctx context.Context, rw http.ResponseWriter, req *http.Request) {
	h(ctx, rw, req)
}

type ContextAdapter struct {
	ctx     context.Context
	handler ContextHandler
}

func (ca *ContextAdapter) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	ca.handler.ServeHTTPContext(ca.ctx, rw, req)
}

func main() {

	db, err := models.NewDB()
	if err != nil {
		log.Panic(err)
	}
	ctx := context.WithValue(context.Background(), "db", db)

	r := mux.NewRouter()
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload", &ContextAdapter{ctx, ContextHandlerFunc(models.UploadContactsToCampaignAndAttachSchedule)})
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{CampaignId}/Numbers/{CategoryID}/Assign", &ContextAdapter{ctx, ContextHandlerFunc(models.AssingExssitingNumbersToCampaign)})
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/RemoveComplete", &ContextAdapter{ctx, ContextHandlerFunc(models.RemoveCompleteProcess)})
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{TrackerId}/Tracker", &ContextAdapter{ctx, ContextHandlerFunc(models.TrackNumberUpload)})
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/Trackers/Data/Print", &ContextAdapter{ctx, ContextHandlerFunc(models.GetTrackingInfo)})
	http.ListenAndServe(":3268", r)

	/*r := mux.NewRouter()
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{CampaignId}/Numbers/{CategoryID}/Assign", &ContextAdapter{ctx, ContextHandlerFunc(models.AssingExssitingNumbersToCampaign)})
	r.Handle("/DVP/API/6.0/CampaignManager/NumberUpload", &ContextAdapter{ctx, ContextHandlerFunc(models.UploadContactsToCampaignAndAttachSchedule)})
	r.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/RemoveComplete", models.RemoveCompleteProcess)
	r.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/{TrackerId}/Tracker", models.TrackNumberUpload)
	r.HandleFunc("/DVP/API/6.0/CampaignManager/NumberUpload/Trackers/Data/Print", models.GetTrackingInfo)

	http.ListenAndServe(":3268", r)
	*/
}
