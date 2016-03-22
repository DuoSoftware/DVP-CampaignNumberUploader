package models

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
	CamScheduleId int `json:"camScheduleId"`
	//CategoryID    int    `json:"CategoryID"`
	ExtraData string `json:"extraData"`
}

type TrackInfo struct {
	Message   string   `json:"message"`
	ErrorList []string `json:"errorList"`
}

const (
//	gophers = 10
//entries = 10
)

var TrackList = make(map[uuid.UUID][]int)

//var WorkersCount = make(map[uuid.UUID]int)
var errorList = make(map[uuid.UUID][]string)

func AppendToTrackList(uid uuid.UUID, gopher_id int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in AppendToTrackList", r)
		}
	}()
	TrackList[uid] = append(TrackList[uid], gopher_id)

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
		fmt.Println("UploadContactsToCampaignAndAttachSchedule - 405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	db, ok := ctx.Value("db").(*sql.DB)
	if !ok {
		fmt.Println("Recovered in UploadContactsToCampaignAndAttachSchedule")
	}

	var gophers = 100

	fmt.Println("------------------ UploadContactsToCampaignAndAttachSchedule ------------------")
	uploadData := UploadData{}
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&uploadData)
	if error != nil {
		log.Println(error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}

	//fmt.Println("Upload Data : ", uploadData)
	entries := len(uploadData.Contacts)
	page := entries / gophers

	u1 := uuid.NewV4()
	fmt.Printf("UUIDv4: %s\n", u1)
	TrackList[u1] = []int{}

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

	reply := UploadData{}
	reply.TrackerId = u1
	outgoingJSON, err := json.Marshal(reply)
	if err != nil {
		log.Println(err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, "{ \"Data\": "+string(outgoingJSON)+",\"IsSuccess\":\"true\"}")
	fmt.Println("------------------ UploadContactsToCampaignAndAttachSchedule. end ------------------", u1)
}

func SaveToDb(gopher_id int, contacts []string, uid uuid.UUID, data UploadData, gophers int, db *sql.DB) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in SaveToDb", r)
		}
	}()

	fmt.Print("Gopher Id : ", gopher_id)
	//fmt.Println("  contacts : ", contacts)

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
		fmt.Printf("contact : %s Save... [%d : %d] \n", contacts[i], gopher_id, i)
		defer stmt.Close()
	}

	AppendToTrackList(uid, gopher_id)
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
		fmt.Println("AssingExssitingNumbersToCampaign - 405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	db, ok := ctx.Value("db").(*sql.DB)
	if !ok {
		fmt.Println("Recovered in UploadContactsToCampaignAndAttachSchedule")
	}
	vars := mux.Vars(req)

	campaignId := vars["CampaignId"]
	CategoryID := vars["CategoryID"]

	existingData := new(ExistingData)
	decoder := json.NewDecoder(req.Body)
	error := decoder.Decode(&existingData)
	if error != nil {
		log.Println("req.Body : ", error.Error())
		http.Error(res, error.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("req.Body : ", decoder)

	// create string to pass
	var sStmt string = "INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"CamContactId\", \"ExtraData\",\"createdAt\", \"updatedAt\") SELECT $1, $2,\"CamContactId\" , $3,now(),now() FROM \"DB_CAMP_ContactInfos\" WHERE  \"CategoryID\"=$4"

	//"INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"CamContactId\", \"ExtraData\",\"createdAt\", \"updatedAt\") SELECT  " + CampaignId + ", " + existingData.CamScheduleId + ", 'CamContactId', " + existingData.extraData + ",now(),now() FROM \"DB_CAMP_ContactInfos\" WHERE  \"CategoryID\"=" + existingData.CategoryID

	msg := "{ \"msg\":\"Process Complete.\",\"IsSuccess\":\"true\"}"
	stmt, err := db.Prepare(sStmt)
	if err != nil {
		log.Panic(err)
		msg = "{ \"msg\":\"Error.\",\"IsSuccess\":\"false\"}"
	}

	res.WriteHeader(http.StatusCreated)
	reply, err := stmt.Exec(campaignId, existingData.CamScheduleId, existingData.ExtraData, CategoryID)
	if err != nil || reply == nil {
		msg = "{ \"msg\":\"Error.\",\"IsSuccess\":\"false\"}"
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
		fmt.Println("GetTrackingInfo - 405")
		http.Error(res, http.StatusText(405), 405)
		return
	}

	fmt.Println("------------- Tracking Info -----------------")
	fmt.Println(" TrackList : ", TrackList)
	fmt.Println(" errorList : ", errorList)
	fmt.Println("------------------------------")
	res.WriteHeader(http.StatusCreated)
	fmt.Fprint(res, "{ \"msg\":\"Error.\",\"IsSuccess\":\"true\"}")

}

func RemoveCompleteProcess(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in RemoveCompleteProcess", r)
		}
	}()
	res.Header().Set("Content-Type", "application/json")
	if req.Method != "POST" {
		fmt.Println("RemoveCompleteProcess - 405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	list := []uuid.UUID{}
	for index, element := range TrackList {
		fmt.Println("index", index)
		fmt.Println("element", element)

		if 100 == len(TrackList[index]) {
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
	fmt.Fprint(res, "{ \"msg\":"+string(outgoingJSON)+",\"IsSuccess\":\"true\"}")
}

func TrackNumberUpload(ctx context.Context, res http.ResponseWriter, req *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in TrackNumberUpload", r)
		}
	}()
	res.Header().Set("Content-Type", "application/json")
	if req.Method != "GET" {
		fmt.Println("TrackNumberUpload -405")
		http.Error(res, http.StatusText(405), 405)
		return
	}
	vars := mux.Vars(req)

	trackerId, err := uuid.FromString(vars["TrackerId"])
	if err != nil {
		fmt.Println("Something gone wrong: %s", err)
	}

	trackList := TrackList[trackerId]

	trackInfo := new(TrackInfo)
	trackInfo.Message = "Process Complete."
	trackInfo.ErrorList = errorList[trackerId]

	if 100 == len(trackList) {
		delete(TrackList, trackerId)
		delete(errorList, trackerId)
	} else if 0 == len(trackList) {
		trackInfo.Message = "Invalid Track ID or Complete Process."
		if len(errorList[trackerId]) > 0 {
			trackInfo.Message = "Try again few minutes."
		}
	} else {
		trackInfo.Message = "Incomplete Process."
	}
	fmt.Println("------------- Track Number Upload -----------------", trackerId)

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

	fmt.Fprint(res, "{ \"msg\":"+string(outgoingJSON)+",\"IsSuccess\":\"true\"}")
}
