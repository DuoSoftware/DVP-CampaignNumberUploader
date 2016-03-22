// DVP-CampaignNoUploader project main.go

package main

import (
	"DVP-CampaignNoUploader/models"
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
	http.Handle("/DVP/API/6.0/CampaignManager/NumberUpload", &ContextAdapter{ctx, ContextHandlerFunc(models.UploadContactsToCampaignAndAttachSchedule)})
	http.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/RemoveComplete", &ContextAdapter{ctx, ContextHandlerFunc(models.RemoveCompleteProcess)})
	http.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/TrackerInfo", &ContextAdapter{ctx, ContextHandlerFunc(models.GetTrackingInfo)})
	http.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{TrackerId}", &ContextAdapter{ctx, ContextHandlerFunc(models.TrackNumberUpload)})
	http.Handle("/DVP/API/6.0/CampaignManager/NumberUpload/{CampaignId}/Numbers/Existing", &ContextAdapter{ctx, ContextHandlerFunc(models.AssingExssitingNumbersToCampaign)})
	http.ListenAndServe(":3268", nil)

}
