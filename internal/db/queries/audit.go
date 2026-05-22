package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/xiehqing/infra/pkg/ormx"
	"gorm.io/gorm"

	"github.com/xiehqing/hiauth/internal/audit"
	"github.com/xiehqing/hiauth/internal/db/entity"
)

type AuditLogListFilter struct {
	ormx.Pagination
	OperatorName string `json:"operatorName" form:"operatorName"`
	Module       string `json:"module" form:"module"`
	Action       string `json:"action" form:"action"`
	ResourceType string `json:"resourceType" form:"resourceType"`
	ResourceID   int64  `json:"resourceId" form:"resourceId"`
	Status       string `json:"status" form:"status"`
	StartTime    *time.Time
	EndTime      *time.Time
}

type auditRecordOptions struct {
	module       string
	action       string
	resourceType string
	resourceID   int64
	tableName    string
	operation    string
	description  string
	before       any
	after        any
}

type RequestAudit struct {
	Module       string
	Action       string
	ResourceType string
	ResourceID   int64
	Description  string
	Status       string
	ErrorMessage string
	DurationMs   int64
}

type auditFieldChange struct {
	Field  string `json:"field"`
	Before any    `json:"before"`
	After  any    `json:"after"`
}

func (q *Queries) GetAuditLog(ctx context.Context, id int64) (*entity.AuditLog, error) {
	var log entity.AuditLog
	err := q.db.WithContext(ctx).Preload("Changes").First(&log, "id = ?", id).Error
	if err != nil {
		return nil, ormx.NotFoundAsNil(err)
	}
	return &log, nil
}

func (q *Queries) ListAuditLogs(ctx context.Context, filter AuditLogListFilter) (ormx.PageResult[entity.AuditLog], error) {
	db := q.db.WithContext(ctx).Model(&entity.AuditLog{})
	if ormx.KeywordPresent(filter.Keyword) {
		keyword := ormx.LikeKeyword(filter.Keyword)
		db = db.Where("operator_name LIKE ? OR description LIKE ? OR path LIKE ?", keyword, keyword, keyword)
	}
	if filter.OperatorName != "" {
		db = db.Where("operator_name LIKE ?", ormx.LikeKeyword(filter.OperatorName))
	}
	if filter.Module != "" {
		db = db.Where("module = ?", filter.Module)
	}
	if filter.Action != "" {
		db = db.Where("action = ?", filter.Action)
	}
	if filter.ResourceType != "" {
		db = db.Where("resource_type = ?", filter.ResourceType)
	}
	if filter.ResourceID > 0 {
		db = db.Where("resource_id = ?", filter.ResourceID)
	}
	if filter.Status != "" {
		db = db.Where("status = ?", filter.Status)
	}
	if filter.StartTime != nil {
		db = db.Where("created_at >= ?", *filter.StartTime)
	}
	if filter.EndTime != nil {
		db = db.Where("created_at <= ?", *filter.EndTime)
	}

	return ormx.Paginate[entity.AuditLog](db, filter.Pagination, map[string]string{
		"id":           "id",
		"operatorName": "operator_name",
		"module":       "module",
		"action":       "action",
		"resourceType": "resource_type",
		"resourceId":   "resource_id",
		"status":       "status",
		"durationMs":   "duration_ms",
		"createdAt":    "created_at",
		"updatedAt":    "updated_at",
	})
}

func (q *Queries) CreateRequestAudit(ctx context.Context, request RequestAudit) error {
	auditContext, _ := audit.FromContext(ctx)
	log := entity.AuditLog{
		RequestID:    auditContext.RequestID,
		OperatorID:   auditContext.OperatorID,
		OperatorName: auditContext.OperatorName,
		Module:       request.Module,
		Action:       request.Action,
		ResourceType: request.ResourceType,
		ResourceID:   request.ResourceID,
		Description:  request.Description,
		Method:       auditContext.Method,
		Path:         auditContext.Path,
		IP:           auditContext.IP,
		UserAgent:    auditContext.UserAgent,
		Status:       request.Status,
		ErrorMessage: request.ErrorMessage,
		DurationMs:   request.DurationMs,
	}
	if log.Module == "" {
		log.Module = auditModule(request.ResourceType)
	}
	if log.Action == "" {
		log.Action = entity.AuditOperationQuery
	}
	if log.ResourceType == "" {
		log.ResourceType = "request"
	}
	if log.Status == "" {
		log.Status = entity.AuditStatusSuccess
	}
	if log.Description == "" {
		log.Description = auditRequestDescription(log.Action, log.ResourceType, log.Path)
	}
	return q.db.WithContext(ctx).Create(&log).Error
}

func (q *Queries) auditCreate(ctx context.Context, tx *gorm.DB, resourceType, tableName string, recordID int64, after any) error {
	return q.recordAudit(ctx, tx, auditRecordOptions{
		module:       auditModule(resourceType),
		action:       entity.AuditOperationCreate,
		resourceType: resourceType,
		resourceID:   recordID,
		tableName:    tableName,
		operation:    entity.AuditOperationCreate,
		description:  auditDescription(resourceType, entity.AuditOperationCreate, recordID),
		after:        after,
	})
}

func (q *Queries) auditUpdate(ctx context.Context, tx *gorm.DB, resourceType, tableName string, recordID int64, before, after any) error {
	return q.recordAudit(ctx, tx, auditRecordOptions{
		module:       auditModule(resourceType),
		action:       entity.AuditOperationUpdate,
		resourceType: resourceType,
		resourceID:   recordID,
		tableName:    tableName,
		operation:    entity.AuditOperationUpdate,
		description:  auditDescription(resourceType, entity.AuditOperationUpdate, recordID),
		before:       before,
		after:        after,
	})
}

func (q *Queries) auditDelete(ctx context.Context, tx *gorm.DB, resourceType, tableName string, recordID int64, before any) error {
	return q.recordAudit(ctx, tx, auditRecordOptions{
		module:       auditModule(resourceType),
		action:       entity.AuditOperationDelete,
		resourceType: resourceType,
		resourceID:   recordID,
		tableName:    tableName,
		operation:    entity.AuditOperationDelete,
		description:  auditDescription(resourceType, entity.AuditOperationDelete, recordID),
		before:       before,
	})
}

func (q *Queries) auditAuthorize(ctx context.Context, tx *gorm.DB, resourceType, tableName string, recordID int64, before, after any) error {
	return q.recordAudit(ctx, tx, auditRecordOptions{
		module:       auditModule(resourceType),
		action:       entity.AuditOperationAuthorize,
		resourceType: resourceType,
		resourceID:   recordID,
		tableName:    tableName,
		operation:    entity.AuditOperationUpdate,
		description:  auditDescription(resourceType, entity.AuditOperationAuthorize, recordID),
		before:       before,
		after:        after,
	})
}

func (q *Queries) recordAudit(ctx context.Context, tx *gorm.DB, options auditRecordOptions) error {
	auditContext, _ := audit.FromContext(ctx)
	beforeData, beforeMap := marshalAuditData(options.before)
	afterData, afterMap := marshalAuditData(options.after)
	changedFields := marshalChangedFields(beforeMap, afterMap)
	fieldChanges := marshalFieldChanges(beforeMap, afterMap)

	log := entity.AuditLog{
		RequestID:    auditContext.RequestID,
		OperatorID:   auditContext.OperatorID,
		OperatorName: auditContext.OperatorName,
		Module:       options.module,
		Action:       options.action,
		ResourceType: options.resourceType,
		ResourceID:   options.resourceID,
		Description:  options.description,
		Method:       auditContext.Method,
		Path:         auditContext.Path,
		IP:           auditContext.IP,
		UserAgent:    auditContext.UserAgent,
		Status:       entity.AuditStatusSuccess,
	}
	if log.Module == "" {
		log.Module = "系统"
	}
	if log.Description == "" {
		log.Description = auditDescription(options.resourceType, options.action, options.resourceID)
	}

	change := entity.AuditChange{
		DBTableName:   options.tableName,
		RecordID:      options.resourceID,
		Operation:     options.operation,
		BeforeData:    beforeData,
		AfterData:     afterData,
		ChangedFields: changedFields,
		FieldChanges:  fieldChanges,
	}

	if err := tx.Create(&log).Error; err != nil {
		return err
	}
	change.AuditLogID = log.ID
	return tx.Create(&change).Error
}

func marshalAuditData(value any) (string, map[string]any) {
	if value == nil {
		return "", nil
	}
	bytes, err := json.Marshal(value)
	if err != nil {
		return "", nil
	}

	var data map[string]any
	if err := json.Unmarshal(bytes, &data); err != nil {
		return string(bytes), nil
	}
	maskAuditMap(data)
	bytes, err = json.Marshal(data)
	if err != nil {
		return "", data
	}
	return string(bytes), data
}

func maskAuditMap(data map[string]any) {
	configKey := ""
	if value, ok := data["key"].(string); ok {
		configKey = strings.ToLower(value)
	}
	for key, value := range data {
		lowerKey := strings.ToLower(key)
		if sensitiveAuditField(lowerKey) || (key == "value" && sensitiveConfigKey(configKey)) {
			data[key] = "******"
			continue
		}
		switch item := value.(type) {
		case map[string]any:
			maskAuditMap(item)
		case []any:
			for _, child := range item {
				if childMap, ok := child.(map[string]any); ok {
					maskAuditMap(childMap)
				}
			}
		}
	}
}

func sensitiveAuditField(key string) bool {
	return strings.Contains(key, "password") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "privatekey") ||
		strings.Contains(key, "private_key") ||
		strings.Contains(key, "authorization")
}

func sensitiveConfigKey(key string) bool {
	return strings.Contains(key, "password") ||
		strings.Contains(key, "secret") ||
		strings.Contains(key, "token") ||
		strings.Contains(key, "private_key") ||
		strings.Contains(key, "privatekey") ||
		strings.Contains(key, "storage")
}

func marshalChangedFields(before, after map[string]any) string {
	changes := fieldChanges(before, after)
	fields := make([]string, 0, len(changes))
	for _, change := range changes {
		fields = append(fields, change.Field)
	}
	bytes, err := json.Marshal(fields)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

func marshalFieldChanges(before, after map[string]any) string {
	changes := fieldChanges(before, after)
	bytes, err := json.Marshal(changes)
	if err != nil {
		return "[]"
	}
	return string(bytes)
}

func fieldChanges(before, after map[string]any) []auditFieldChange {
	if before == nil && after == nil {
		return []auditFieldChange{}
	}

	keys := make(map[string]struct{}, len(before)+len(after))
	for key := range before {
		if !ignoreAuditCompareField(key) {
			keys[key] = struct{}{}
		}
	}
	for key := range after {
		if !ignoreAuditCompareField(key) {
			keys[key] = struct{}{}
		}
	}

	fields := make([]string, 0, len(keys))
	for field := range keys {
		fields = append(fields, field)
	}
	sort.Strings(fields)

	result := make([]auditFieldChange, 0, len(fields))
	for _, field := range fields {
		beforeValue := valueOfAuditField(before, field)
		afterValue := valueOfAuditField(after, field)
		if !reflect.DeepEqual(beforeValue, afterValue) {
			result = append(result, auditFieldChange{
				Field:  field,
				Before: beforeValue,
				After:  afterValue,
			})
		}
	}
	return result
}

func valueOfAuditField(data map[string]any, field string) any {
	if data == nil {
		return nil
	}
	return data[field]
}

func ignoreAuditCompareField(field string) bool {
	switch field {
	case "createdAt", "createdBy", "updatedAt", "updatedBy":
		return true
	default:
		return false
	}
}

func auditModule(resourceType string) string {
	switch resourceType {
	case "user", "user_roles":
		return "用户管理"
	case "role", "role_menus":
		return "角色管理"
	case "department":
		return "部门管理"
	case "menu":
		return "菜单管理"
	case "system_config":
		return "系统配置"
	default:
		return "系统"
	}
}

func auditDescription(resourceType, action string, resourceID int64) string {
	actionName := map[string]string{
		entity.AuditOperationCreate:    "新增",
		entity.AuditOperationUpdate:    "修改",
		entity.AuditOperationDelete:    "删除",
		entity.AuditOperationAuthorize: "授权",
	}[action]
	if actionName == "" {
		actionName = action
	}
	return fmt.Sprintf("%s%s，ID：%d", actionName, auditResourceName(resourceType), resourceID)
}

func auditResourceName(resourceType string) string {
	switch resourceType {
	case "user":
		return "用户"
	case "user_roles":
		return "用户角色"
	case "role":
		return "角色"
	case "role_menus":
		return "角色菜单"
	case "department":
		return "部门"
	case "menu":
		return "菜单"
	case "system_config":
		return "系统配置"
	default:
		return resourceType
	}
}

func auditRequestDescription(action, resourceType, path string) string {
	switch action {
	case entity.AuditOperationLogin:
		return "用户登录"
	case entity.AuditOperationLogout:
		return "退出登录"
	case entity.AuditOperationQuery:
		return fmt.Sprintf("查询%s", auditResourceName(resourceType))
	default:
		return fmt.Sprintf("%s接口：%s", action, path)
	}
}
