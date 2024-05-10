package cmd

import (
	"database/sql/driver"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/minio/minio/internal/logger"
)

type JParameter struct {
	SnapStatus int    `json:"snap_status"`
	UUID       string `json:"uuid"`
}

func (c JParameter) Value() (driver.Value, error) {
	b, err := json.Marshal(c)
	return string(b), err
}

func (c *JParameter) Scan(input interface{}) error {
	return json.Unmarshal(input.([]byte), c)
}

type AdmGeneralBackupSnap struct {
	ID          int        `gorm:"column:general_backup_snap_id" json:"general_backup_snap_id"`
	Name        string     `gorm:"column:name" json:"name"`
	BackupSrcID int        `gorm:"column:general_backup_src_id" json:"general_backup_src_id"`
	UserID      int        `gorm:"column:user_id" json:"user_id"`
	SnapStatus  int        `gorm:"column:snap_status" json:"snap_status"`
	SnapType    int        `gorm:"column:snap_type" json:"snap_type"`
	Parameter   JParameter `gorm:"column:json_parameter;TYPE:json" json:"json_parameter"`
}

func (AdmGeneralBackupSnap) TableName() string {
	return "t_adm_general_backup_snap"
}

// update adm backup job status
func UpdateBackupStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "UpdateBackupStatus")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	backupId := r.Form.Get("backupId")
	status := r.Form.Get("status")
	if backupId == "" || status == "" {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInvalidUpdateBackupStatusParams), r.URL)
		return
	}
	ID, err := strconv.Atoi(backupId)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}
	snap_status, err := strconv.Atoi(status)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}
	log.Printf("backupId:%s,status:%s", backupId, status)
	globalDB.Exec("UPDATE t_adm_general_backup_snap SET json_parameter=JSON_SET(json_parameter, '$.snap_status', ?) WHERE general_backup_snap_id=?", snap_status, backupId)
	var backupSnap AdmGeneralBackupSnap
	globalDB.Where(&AdmGeneralBackupSnap{ID: ID}).First(&backupSnap)
	jsonBytes, err := json.Marshal(backupSnap)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}
	writeSuccessResponseJSON(w, jsonBytes)
}

// update adm restore job status
func UpdateRestoreStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "UpdateRestoreStatus")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	restoreId := r.Form.Get("restoreId")
	status := r.Form.Get("status")
	backupType := r.Form.Get("restoreType")
	log.Printf("restoreId:%s,backupType:%s,status:%s", restoreId, backupType, status)
	if restoreId == "" || status == "" || backupType == "" {
		writeErrorResponseJSON(ctx, w, errorCodes.ToAPIErr(ErrInvalidUpdateBackupStatusParams), r.URL)
		return
	}
	var backupSnap AdmGeneralBackupSnap
	jsonBytes, err := json.Marshal(backupSnap)
	if err != nil {
		writeErrorResponseJSON(ctx, w, toAPIError(ctx, err), r.URL)
		return
	}
	writeSuccessResponseJSON(w, jsonBytes)
}
