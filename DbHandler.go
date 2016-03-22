package main

import (
	"database/sql"
	"encoding/json"

	"fmt"
	"github.com/go-contrib/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/net/context"
	"log"
	"net/http"
)

type UploadData struct {
	Contacts      []string  `json:"contacts"`
	CampaignId    int       `json:"campaignId"`
	CamScheduleId int       `json:"camScheduleId"`
	TenantId      int       `json:"tenantId"`
	CompanyId     int       `json:"companyId"`
	CategoryId    int       `json:"categoryId"`
	ExtraData     string    `json:"extraData"`
	TrackerId     uuid.UUID `json:"trackerId"`
}

type ExistingData struct {
	CamScheduleId int    `json:"camScheduleId"`
	CategoryID    int    `json:"CategoryID"`
	ExtraData     string `json:"extraData"`
}

type TrackInfo struct {
	Message   string   `json:"message"`
	ErrorList []string `json:"errorList"`
}

const (
//	gophers = 10
//entries = 10
)

var TrackList = make(map[uuid.UUID][]string)
var WorkersCount = make(map[uuid.UUID]int)
var errorList = make(map[uuid.UUID][]string)

func AppendToTrackList(uid uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in AppendToTrackList", r)
		}
	}()
	TrackList[uid] = append(TrackList[uid], "done")
	fmt.Println("AppendToTrackList : ", TrackList[uid])
}

//ctx context.Context, w http.ResponseWriter, r *http.Request
func UploadContactsToCampaignAndAttachSchedule(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in UploadContactsToCampaignAndAttachSchedule", r)
		}
	}()
	res.Header().Set("Content-Type", "application/json")
	if req.Method != "POST" {
		fmt.Println("405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	db, ok := ctx.Value("db").(*sql.DB)
	if !ok {
		fmt.Println("Recovered in UploadContactsToCampaignAndAttachSchedule")
	}

	var gophers = 10

	fmt.Println("------------------ UploadContactsToCampaignAndAttachSchedule ------------------")
	uploadData := UploadData{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&uploadData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Upload Data : ", uploadData)
	entries := len(uploadData.Contacts)
	page := entries / gophers

	u1 := uuid.NewV4()
	fmt.Printf("UUIDv4: %s\n", u1)
	TrackList[u1] = []string{}
	WorkersCount[u1] = page

	if page == 0 {
		WorkersCount[u1] = 1
	}

	fmt.Println("page Count : ", page)
	if 1 >= page {
		gophers = 1
	}
	j := 0
	// run the insert function using 10 go routines
	for i := 0; i < gophers; i++ {
		contacts := uploadData.Contacts[j : j+page]

		if (i == gophers-1) && (j < entries) {
			contacts = uploadData.Contacts[j:entries]
		}

		// spin up a gopher
		go SaveToDb(i, contacts, u1, uploadData, gophers, db)
		j = j + page
	}

	uploadData.TrackerId = u1
	outgoingJSON, err := json.Marshal(uploadData)
	if err != nil {
		log.Println(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, string(outgoingJSON))
	fmt.Println("------------------ UploadContactsToCampaignAndAttachSchedule. end ------------------", u1)
}

func SaveToDb(gopher_id int, contacts []string, uid uuid.UUID, data UploadData, gophers int, db *sql.DB) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in SaveToDb", r)
		}
	}()

	fmt.Print("Gopher Id : ", gopher_id)
	fmt.Println("  contacts : ", contacts)

	db.SetMaxOpenConns(gophers)

	// create string to pass
	var sStmt string = "WITH i AS (INSERT INTO \"DB_CAMP_ContactInfos\"(\"ContactId\", \"CategoryID\", \"TenantId\", \"CompanyId\",\"createdAt\" ,\"updatedAt\" ) VALUES ($1, $2, $3, $4,now(),now()) RETURNING \"CamContactId\") INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"ExtraData\",\"createdAt\" ,\"updatedAt\", \"CamContactId\")    VALUES ($5, $6, $7,now(),now(), (SELECT \"CamContactId\" FROM i))" // "WITH i AS (INSERT INTO tb1a (t) VALUES ($1) RETURNING id) INSERT INTO tb1b (t)SELECT id FROM i "
	//"insert into test (gopher_id, created) values ($1, $2)" //"insert into test (gopher_id, created) values ($1, $2)" //
	stmt, err := db.Prepare(sStmt)
	if err != nil {
		log.Panic(err)
	}

	for i := 0; i < len(contacts); i++ {
		//ContactId\", \"CategoryID\", \"TenantId\", \"CompanyId CampaignId\", \"CamScheduleId\", \"ExtraData
		res, err := stmt.Exec(contacts[i], data.CategoryId, data.TenantId, data.CompanyId, data.CampaignId, data.CamScheduleId, data.ExtraData)
		if err != nil || res == nil {
			log.Print(err)
			errorList[uid] = append(errorList[uid], contacts[i])
		}
		defer stmt.Close()
	}

	AppendToTrackList(uid)
}

//---------------------Assing Exssiting Numbers To Campaign--------------------------\\

func AssingExssitingNumbersToCampaign(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in AssingExssitingNumbersToCampaign", r)
		}
	}()

	res.Header().Set("Content-Type", "application/json")
	if req.Method != "POST" {
		fmt.Println("405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	db, ok := ctx.Value("db").(*sql.DB)
	if !ok {
		fmt.Println("Recovered in UploadContactsToCampaignAndAttachSchedule")
	}
	vars := mux.Vars(req)

	campaignId := vars["CampaignId"]

	existingData := new(ExistingData)
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&existingData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}

	db.SetMaxOpenConns(1)

	// create string to pass
	var sStmt string = "INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"CamContactId\", \"ExtraData\",\"createdAt\", \"updatedAt\") SELECT $1, $2,\"CamContactId\" , $3,now(),now() FROM \"DB_CAMP_ContactInfos\" WHERE  \"CategoryID\"=$4"

	//"INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"CamContactId\", \"ExtraData\",\"createdAt\", \"updatedAt\") SELECT  " + CampaignId + ", " + existingData.CamScheduleId + ", 'CamContactId', " + existingData.extraData + ",now(),now() FROM \"DB_CAMP_ContactInfos\" WHERE  \"CategoryID\"=" + existingData.CategoryID

	msg := "Process Complete."
	stmt, err := db.Prepare(sStmt)
	if err != nil {
		log.Panic(err)
		msg = "Error"
	}

	res.WriteHeader(http.StatusCreated)
	reply, err := stmt.Exec(campaignId, existingData.CamScheduleId, existingData.ExtraData, existingData.CategoryID)
	if err != nil || reply == nil {
		msg = "Error"
		fmt.Fprint(res, err.Error())
		log.Panic(err)
	}

	stmt.Close()

	fmt.Fprint(res, msg)

}

//---------------------End-Assing Exssiting Numbers To Campaign--------------------------\\

func GetTrackingInfo(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in GetTrackingInfo", r)
		}
	}()

	res.Header().Set("Content-Type", "application/json")
	if req.Method != "GET" {
		fmt.Println("405")
		http.Error(res, http.StatusText(405), 405)
		return
	}

	fmt.Println("------------- Tracking Info -----------------")
	fmt.Println(" workerCount : ", WorkersCount)
	fmt.Println(" TrackList : ", TrackList)
	fmt.Println(" errorList : ", errorList)
	fmt.Println("------------------------------")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, "{\"msg\": \"Done\"}")

}

func RemoveCompleteProcess(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in RemoveCompleteProcess", r)
		}
	}()
	res.Header().Set("Content-Type", "application/json")
	if req.Method != "POST" {
		fmt.Println("405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	list := []uuid.UUID{}
	for index, element := range WorkersCount {
		fmt.Println("index", index)
		fmt.Println("element", element)

		if WorkersCount[index] == len(TrackList[index]) {
			delete(WorkersCount, index)
			delete(TrackList, index)
			delete(errorList, index)
			list = append(list, index)
		}
	}

	outgoingJSON, err := json.Marshal(list)
	if err != nil {
		log.Println(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, string(outgoingJSON))
}

func TrackNumberUpload(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in TrackNumberUpload", r)
		}
	}()
	res.Header().Set("Content-Type", "application/json")
	if req.Method != "GET" {
		fmt.Println("405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	vars := mux.Vars(req)

	trackerId, err := uuid.FromString(vars["TrackerId"])
	if err != nil {
		fmt.Println("Something gone wrong: %s", err)
	}

	workerCount := WorkersCount[trackerId]
	trackList := TrackList[trackerId]

	trackInfo := TrackInfo{}
	trackInfo.Message = "Process Complete."
	trackInfo.ErrorList = errorList[trackerId]
	if workerCount == len(trackList) {
		delete(WorkersCount, trackerId)
		delete(TrackList, trackerId)
		delete(errorList, trackerId)
	} else {
		trackInfo.Message = "Invalid Track ID or Incomplete Process."

	}
	fmt.Println("------------- Track Number Upload -----------------", trackerId)
	fmt.Print(" workerCount : ", WorkersCount[trackerId])
	fmt.Print(" TrackList : ", TrackList[trackerId])
	fmt.Println(" errorList : ", errorList[trackerId])
	fmt.Println(" trackInfo : ", trackInfo)
	fmt.Println("------------------------------")

	res.Header().Set("Content-Type", "application/json")
	outgoingJSON, err := json.Marshal(trackInfo)
	if err != nil {
		log.Println(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, string(outgoingJSON))

}
