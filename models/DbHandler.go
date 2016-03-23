package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-contrib/uuid"
	_ "github.com/lib/pq"
	"log"
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
	//CampaignId    int    `json:"CampaignId"`
	ExtraData string `json:"extraData"`
}

type TrackInfo struct {
	Message   string   `json:"message"`
	ErrorList []string `json:"errorList"`
}

var TrackList = make(map[uuid.UUID][]int)
var Items = make(map[uuid.UUID]int)
var errorList = make(map[uuid.UUID][]string)

func UploadContactsToCampaignAndAttachSchedule(db *sql.DB, uploadData UploadData, uid uuid.UUID) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in UploadContactsToCampaignAndAttachSchedule", r)
		}
	}()

	fmt.Println("------------------ UploadContactsToCampaignAndAttachSchedule ------------------")

	TrackList[uid] = []int{}

	var sStmt string = "WITH i AS (INSERT INTO \"DB_CAMP_ContactInfos\"(\"ContactId\", \"CategoryID\", \"TenantId\", \"CompanyId\",\"createdAt\" ,\"updatedAt\" ) VALUES ($1, $2, $3, $4,now(),now()) RETURNING \"CamContactId\") INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"ExtraData\",\"createdAt\" ,\"updatedAt\", \"CamContactId\")    VALUES ($5, $6, $7,now(),now(), (SELECT \"CamContactId\" FROM i))" // "WITH i AS (INSERT INTO tb1a (t) VALUES ($1) RETURNING id) INSERT INTO tb1b (t)SELECT id FROM i "
	stmt, err := db.Prepare(sStmt)
	if err != nil {
		log.Panic(err)
	}

	lenth := len(uploadData.Contacts)
	Items[uid] = lenth
	for i := 0; i < lenth; i++ {
		//ContactId\", \"CategoryID\", \"TenantId\", \"CompanyId CampaignId\", \"CamScheduleId\", \"ExtraData
		res, err := stmt.Exec(uploadData.Contacts[i], uploadData.CategoryId, uploadData.TenantId, uploadData.CompanyId, uploadData.CampaignId, uploadData.CamScheduleId, uploadData.ExtraData)
		if err != nil || res == nil {
			log.Print(err)
			errorList[uid] = append(errorList[uid], uploadData.Contacts[i])
		}
		fmt.Printf("Tracker : [%s] contact : %s Save... [%d : %d] \n", uid, uploadData.Contacts[i], lenth, i)
		TrackList[uid] = append(TrackList[uid], i)
	}

	fmt.Println("------------------ UploadContactsToCampaignAndAttachSchedule. end ------------------", uid)
}

func GetTrackingInfo() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in GetTrackingInfo", r)
		}
	}()

	fmt.Println("------------- Tracking Info -----------------")
	fmt.Println(" TrackList : ", TrackList)
	fmt.Println(" ErrorList : ", errorList)
	fmt.Println(" Items : ", Items)
	fmt.Println("------------------------------")

}

//---------------------Assing Exssiting Numbers To Campaign--------------------------\\

func AssingExssitingNumbersToCampaign(db *sql.DB, existingData ExistingData, uid uuid.UUID, campaignId int, categoryID int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in AssingExssitingNumbersToCampaign", r)
		}
	}()

	// create string to pass
	var sStmt string = "INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"CamContactId\", \"ExtraData\",\"createdAt\", \"updatedAt\") SELECT $1, $2,\"CamContactId\" , $3,now(),now() FROM \"DB_CAMP_ContactInfos\" WHERE  \"CategoryID\"=$4"

	//"INSERT INTO \"DB_CAMP_ContactSchedules\"(\"CampaignId\", \"CamScheduleId\", \"CamContactId\", \"ExtraData\",\"createdAt\", \"updatedAt\") SELECT  " + CampaignId + ", " + existingData.CamScheduleId + ", 'CamContactId', " + existingData.extraData + ",now(),now() FROM \"DB_CAMP_ContactInfos\" WHERE  \"CategoryID\"=" + existingData.CategoryID

	msg := "{ \"msg\":\"Process Complete.\",\"IsSuccess\":\"true\"}"
	stmt, err := db.Prepare(sStmt)
	if err != nil {
		log.Panic(err)
		msg = "{ \"msg\":\"Error.\",\"IsSuccess\":\"false\"}"
		errorList[uid] = append(errorList[uid], msg)
	}

	reply, err := stmt.Exec(campaignId, existingData.CamScheduleId, existingData.ExtraData, categoryID)
	if err != nil || reply == nil {
		errorList[uid] = append(errorList[uid], sStmt)
		log.Panic(err)
	}
	TrackList[uid] = append(TrackList[uid], -1)
	stmt.Close()

}

//---------------------End-Assing Exssiting Numbers To Campaign--------------------------\\

func RemoveCompleteProcess() []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in RemoveCompleteProcess", r)
		}
	}()

	list := []uuid.UUID{}
	for index, element := range TrackList {
		fmt.Println("index", index)
		fmt.Println("element", element)

		if 1 == len(TrackList[index]) {
			delete(TrackList, index)
			delete(errorList, index)
			list = append(list, index)
		}
	}

	outgoingJSON, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err.Error())
		return []byte(err.Error())
	}
	return outgoingJSON
}

func TrackNumberUpload(tracker string) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in TrackNumberUpload", r)
		}
	}()

	trackerId, err := uuid.FromString(tracker)
	if err != nil {
		fmt.Println("Something gone wrong: %s", err)
	}

	trackList := TrackList[trackerId]

	trackInfo := new(TrackInfo)
	trackInfo.Message = "Process Complete."
	trackInfo.ErrorList = errorList[trackerId]

	if Items[trackerId] == len(trackList) {
		delete(TrackList, trackerId)
		delete(errorList, trackerId)
		delete(Items, trackerId)
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
	fmt.Println(" ErrorList : ", errorList[trackerId])
	fmt.Println(" Items : ", Items[trackerId])
	fmt.Println(" TrackInfo : ", trackInfo)
	fmt.Println("------------------------------")

	outgoingJSON, err := json.Marshal(trackInfo)
	if err != nil {
		fmt.Println(err.Error())
		return []byte(err.Error())
	}
	return outgoingJSON

}
