package cmd

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/minio/minio/internal/logger"
)

var mutex sync.Mutex
var mapTimer map[string]*time.Timer

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

type Result struct {
	OK     int    `json:"ok"`
	ErrMsg string `json:"errmsg"`
}

type UpdateRecordStatusResult struct {
	Result
}

type ProgressResult struct {
	Result
	Completed    int    `json:"completed"`
	Succeed      int    `json:"succeed"`
	SnapId       int    `gorm:"column:snapid" json:"snapid"`
	Status       int    `gorm:"column:status" json:"status"`
	Rate         string `gorm:"column:compress_ratio_str" json:"rate"`
	RealData     string `gorm:"column:logical_used_str" json:"real_data"`
	TransferData string `gorm:"column:used_str" json:"transfer_data"`
}

func saveRecordStatus(recordStatus int, recordId string, BusinessType int) bool {
	fmt.Printf("save record(%s) status:%d\n", recordId, recordStatus)
	if BusinessType == 0 || BusinessType == 1 { //备份
		globalDB.Exec("UPDATE t_adm_general_backup_snap SET json_parameter=JSON_SET(json_parameter, '$.snap_status', ?) WHERE general_backup_snap_id=?", recordStatus, recordId)
		return true
	} else if BusinessType == 2 { //恢复
		globalDB.Exec("UPDATE t_adm_general_backup_recover SET json_parameters=JSON_SET(json_parameters, '$.recover_status', ?) WHERE general_backup_recover_id=?", recordStatus, recordId)
		return true
	} else if BusinessType == 3 {
		globalDB.Exec("UPDATE t_adm_general_backup_check_list SET json_parameter=JSON_SET(json_parameter, '$.check_status', ?) WHERE general_backup_check_list_id=?", recordStatus, recordId)
		return true
	} else if BusinessType == 4 { //CDM restore
		if recordStatus == 3 { //终止
			recordStatus = 12
		}
		if recordStatus == 2 { //失败
			recordStatus = 3
		}
		globalDB.Exec("UPDATE t_adm_vdb SET json_parameter=JSON_SET(json_parameter, '$.vdbstatus', ?) WHERE vdb_id=?", recordStatus, recordId)
		return true
	} else if BusinessType == 6 || BusinessType == 8 {
		globalDB.Exec("UPDATE t_adm_arch_list SET json_parameter=JSON_SET(json_parameter, '$.arch_status', ?) WHERE arch_list_id=?", recordStatus, recordId)
		return true
	} else if BusinessType == 7 || BusinessType == 9 {
		globalDB.Exec("UPDATE t_adm_arch_recover SET json_parameter=JSON_SET(json_parameter, '$.recover_status', ?) WHERE arch_recover_id=?", recordStatus, recordId)
		return true
	}
	fmt.Printf("unsupport businessType\n")
	return false
}

// update adm backup job status
func UpdateRecordStatusHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "UpdateRecordStatus")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	recordId := r.Form.Get("recordId")
	status := r.Form.Get("status")
	businessType := r.Form.Get("businessType")
	log.Printf("recordId:%s,status:%s,businessType:%s", recordId, status, businessType)
	var result UpdateRecordStatusResult
	for {
		if recordId == "" || status == "" || businessType == "" {
			result.OK = 1
			result.ErrMsg = "requires the recordId,status,businessType parameters"
			break
		}
		recordStatus, err := strconv.Atoi(status)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		BusinessType, err := strconv.Atoi(businessType)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		if saveRecordStatus(recordStatus, recordId, BusinessType) {
			// 移除定时器
			mutex.Lock()
			defer mutex.Unlock()
			key := businessType + "_" + recordId
			timer, ok := mapTimer[key]
			if ok {
				timer.Stop()
			} else {
				fmt.Println("找不到定时器!")
			}
			delete(mapTimer, key)
			result.OK = 0
		} else {
			result.OK = 1
			result.ErrMsg = "update status failed,please check log"
		}
		break
	}
	jsonBytes, _ := json.Marshal(result)
	// writeSuccessResponseJSON(w, jsonBytes)
	if result.OK == 0 {
		writeResponse(w, http.StatusOK, jsonBytes, mimeJSON)
	} else {
		writeResponse(w, http.StatusBadRequest, jsonBytes, mimeJSON)
	}
	return
}

func checkStatus(result *ProgressResult) {
	if result.Status > 0 {
		result.Completed = 1
	} else {
		result.Completed = 0
	}

	if result.Status == 1 {
		result.Succeed = 1
	} else {
		result.Succeed = 0
	}
}

func ProgressHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "Progress")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	snapid := r.Form.Get("snapid")
	querytype := r.Form.Get("querytype")
	// fmt.Printf("snapid:%s,querytype:%s\n", snapid, querytype)
	// halfphase := r.Form.Get("halfphase")
	var result ProgressResult
	sql := ""
	for {
		if snapid == "" || querytype == "" {
			result.OK = 10
			result.ErrMsg = "snapid or querytype is empty"
			break
		}
		qtype, err := strconv.Atoi(querytype)
		if err != nil {
			result.OK = 10
			result.ErrMsg = "Parameter error:querytype error"
			break
		}
		switch qtype {
		case 0, 1:
			sql = "SELECT general_backup_snap_id AS snapid,json_parameter->>'$.snap_status' AS status,used_str,logical_used_str,compress_ratio_str FROM t_adm_general_backup_snap WHERE general_backup_snap_id=" + snapid + " AND oper_type=" + querytype
			break
		case 2:
			sql = "SELECT general_backup_recover_id AS snapid,json_parameters->>'$.recover_status' AS status FROM t_adm_general_backup_recover WHERE general_backup_recover_id=" + snapid
			break
		case 3:
			sql = "SELECT general_backup_check_list_id AS snapid,json_parameter->>'$.check_status' AS status FROM t_adm_general_backup_check_list WHERE general_backup_check_list_id=" + snapid
			break
		case 4:
			sql = "SELECT vdb_id AS snapid, json_parameter->>'$.vdbstatus' AS status FROM t_adm_vdb WHERE vdb_id=" + snapid
			break
		case 5, 6, 8:
			sql = "SELECT arch_list_id AS snapid,json_parameter->>'$.arch_status' AS status FROM t_adm_arch_list WHERE arch_list_id=" + snapid
			break
		case 7, 9:
			sql = "SELECT arch_recover_id AS snapid, json_parameter->>'$.recover_status' AS status, percent FROM t_adm_arch_recover WHERE arch_recover_id=" + snapid
			break
		default:
			result.OK = 23
			result.ErrMsg = fmt.Sprintf("querytype error")
		}
		if sql != "" {
			if db := globalDB.Raw(sql).First(&result); db.Error == nil {
				result.OK = 0
				checkStatus(&result)
			} else {
				result.OK = 24
				result.ErrMsg = fmt.Sprintf("snapid %s does not exist", snapid)
			}
		}
		break
	}
	jsonBytes, _ := json.Marshal(result)
	writeSuccessResponseJSON(w, jsonBytes)
}

func AddHeartBeatTimerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "AddHeartBeatTimer")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	recordId := r.Form.Get("recordId")
	businessType := r.Form.Get("businessType")
	fmt.Printf("recordId:%s,businessType:%s\n", recordId, businessType)
	var result Result
	for {
		BusinessType, err := strconv.Atoi(businessType)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		duration := 5 * time.Minute
		key := businessType + "_" + recordId

		if timer, ok := mapTimer[key]; ok {
			timer.Reset(duration)
			fmt.Println("定时器已延长，新的期限:", time.Now().Add(duration))
			result.OK = 0
		} else {
			timer := time.AfterFunc(duration, func() {
				fmt.Println("更新状态为失败:", key, time.Now())
				//	更新状态为失败
				saveRecordStatus(2, recordId, BusinessType)
				mutex.Lock()
				defer mutex.Unlock()
				delete(mapTimer, key)
			})
			fmt.Println("添加定时器:", time.Now().Add(duration))
			mutex.Lock()
			defer mutex.Unlock()
			mapTimer[key] = timer
			result.OK = 0
		}
		break
	}

	jsonBytes, _ := json.Marshal(result)
	// writeSuccessResponseJSON(w, jsonBytes)
	if result.OK == 0 {
		writeResponse(w, http.StatusOK, jsonBytes, mimeJSON)
	} else {
		writeResponse(w, http.StatusBadRequest, jsonBytes, mimeJSON)
	}
}

func saveProcessId(procId int, recordId string, BusinessType int) bool {
	if BusinessType == 0 || BusinessType == 1 { //备份
		globalDB.Exec("UPDATE t_adm_general_backup_snap SET json_parameter=JSON_SET(json_parameter, '$.processId', ?) WHERE general_backup_snap_id=?", procId, recordId)
		return true
	} else if BusinessType == 2 { //恢复
		globalDB.Exec("UPDATE t_adm_general_backup_recover SET json_parameters=JSON_SET(json_parameters, '$.processId', ?) WHERE general_backup_recover_id=?", procId, recordId)
		return true
	} else if BusinessType == 3 {
		globalDB.Exec("UPDATE t_adm_general_backup_check_list SET json_parameter=JSON_SET(json_parameter, '$.processId', ?) WHERE general_backup_check_list_id=?", procId, recordId)
		return true
	} else if BusinessType == 4 { //CDM restore
		globalDB.Exec("UPDATE t_adm_vdb SET json_parameter=JSON_SET(json_parameter, '$.processId', ?) WHERE vdb_id=?", procId, recordId)
		return true
	} else if BusinessType == 6 || BusinessType == 8 {
		globalDB.Exec("UPDATE t_adm_arch_list SET json_parameter=JSON_SET(json_parameter, '$.processId', ?) WHERE arch_list_id=?", procId, recordId)
		return true
	} else if BusinessType == 7 || BusinessType == 9 {
		globalDB.Exec("UPDATE t_adm_arch_recover SET json_parameter=JSON_SET(json_parameter, '$.processId', ?) WHERE arch_recover_id=?", procId, recordId)
		return true
	}
	fmt.Printf("unsupport businessType\n")
	return false
}

func UpdateHeartBeatTimerHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "UpdateHeartBeatTimer")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	recordId := r.Form.Get("recordId")
	businessType := r.Form.Get("businessType")
	processId := r.Form.Get("processId")
	var result Result
	fmt.Printf("recordId:%s,businessType:%s,processId:%s\n", recordId, businessType, processId)

	duration := 5 * time.Minute
	key := businessType + "_" + recordId
	for {
		if recordId == "" || processId == "" || businessType == "" {
			result.OK = 1
			result.ErrMsg = "requires the recordId,processId,businessType parameters"
			break
		}
		procId, err := strconv.Atoi(processId)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		BusinessType, err := strconv.Atoi(businessType)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		mutex.Lock()
		defer mutex.Unlock()
		if timer, ok := mapTimer[key]; ok {
			if !timer.Stop() {
				result.OK = 1
			}
			timer.Reset(duration)
			fmt.Println("定时器已延长，新的期限:", time.Now().Add(duration))
			// 更新进程id到数据库
			if saveProcessId(procId, recordId, BusinessType) {
				result.OK = 0
			} else {
				result.OK = 1
			}
		} else {
			result.OK = 1
			result.ErrMsg = "Timer not found"
		}
		break
	}

	jsonBytes, _ := json.Marshal(result)
	// writeSuccessResponseJSON(w, jsonBytes)
	if result.OK == 0 {
		writeResponse(w, http.StatusOK, jsonBytes, mimeJSON)
	} else {
		writeResponse(w, http.StatusBadRequest, jsonBytes, mimeJSON)
	}
}

func formatBytes(bytes uint64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	exp := math.Floor(math.Log(float64(bytes)) / math.Log(1024))
	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(bytes) / math.Pow(1024, exp)
	return fmt.Sprintf("%.2f %s", value, units[int(exp)-1])
}

func saveStatistics(recordId string, BusinessType int, percent int, transferred uint64, speed uint64) bool {
	logical_used_str := formatBytes(transferred)
	speed_str := formatBytes(speed) + "/s"
	if BusinessType == 0 || BusinessType == 1 { //备份
		globalDB.Exec("UPDATE t_adm_general_backup_snap SET percent=?,logical_used=?,logical_used_str=?,json_parameter=JSON_SET(json_parameter, '$.average_speed', ?) WHERE general_backup_snap_id=?", percent, transferred, logical_used_str, speed_str, recordId)
		return true
	} else if BusinessType == 2 { //恢复
		globalDB.Exec("UPDATE t_adm_general_backup_recover SET percent=?,json_parameters=JSON_SET(json_parameters, '$.average_speed', ?)WHERE general_backup_recover_id=?", percent, speed_str, recordId)
		return true
	} else if BusinessType == 3 {
		globalDB.Exec("UPDATE t_adm_general_backup_check_list check_percent=? WHERE general_backup_check_list_id=?", percent, recordId)
		return true
	} else if BusinessType == 4 { //CDM restore
		globalDB.Exec("UPDATE t_adm_vdb SET percent=?,logicalused=?,logicalusedstr=?,json_parameter=JSON_SET(json_parameter, '$.average_speed', ?) WHERE vdb_id=?", percent, transferred, logical_used_str, speed_str, recordId)
		return true
	} else if BusinessType == 6 || BusinessType == 8 {
		globalDB.Exec("UPDATE t_adm_arch_list SET percent=? WHERE arch_list_id=?", percent, recordId)
		return true
	} else if BusinessType == 7 || BusinessType == 9 {
		globalDB.Exec("UPDATE t_adm_arch_recover SET percent=?,json_parameter=JSON_SET(json_parameter, '$.average_speed', ?) WHERE arch_recover_id=?", percent, speed_str, recordId)
		return true
	}
	fmt.Printf("unsupport businessType\n")
	return false
}

func UpdateStatisticsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := newContext(r, w, "UpdateStatisticsHandler")
	defer logger.AuditLog(ctx, w, r, mustGetClaimsFromToken(r))
	recordId := r.Form.Get("recordId")
	businessType := r.Form.Get("businessType")
	total := r.Form.Get("total")
	backuped := r.Form.Get("backuped")
	transferred := r.Form.Get("transferred")
	speed := r.Form.Get("speed")
	var result Result
	for {
		if recordId == "" || businessType == "" {
			result.OK = 1
			result.ErrMsg = "requires the recordId,businessType parameters"
			break
		}
		BusinessType, err := strconv.Atoi(businessType)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		Total, err := strconv.ParseUint(total, 10, 64)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		Backuped, err := strconv.ParseUint(backuped, 10, 64)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		Transferred, err := strconv.ParseUint(transferred, 10, 64)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		Speed, err := strconv.ParseUint(speed, 10, 64)
		if err != nil {
			result.OK = 1
			result.ErrMsg = err.Error()
			break
		}
		// fmt.Printf("recordId:%s,businessType:%s,total:%s,backuped:%s,transferred:%s,speed:%s\n", recordId, businessType, formatBytes(Total), backuped, formatBytes(Transferred), formatBytes(Speed)+"/s")
		percent := 0
		if Total != 0 {
			percent = (int)(Backuped * 100 / Total)
		}
		// 更新进程id到数据库
		if saveStatistics(recordId, BusinessType, percent, Transferred, Speed) {
			result.OK = 0
		} else {
			result.OK = 1
		}
		break
	}

	jsonBytes, _ := json.Marshal(result)
	// writeSuccessResponseJSON(w, jsonBytes)
	if result.OK == 0 {
		writeResponse(w, http.StatusOK, jsonBytes, mimeJSON)
	} else {
		writeResponse(w, http.StatusBadRequest, jsonBytes, mimeJSON)
	}
}
