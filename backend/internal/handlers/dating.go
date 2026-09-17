package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"digcatalog/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// datingReq 送检单编辑请求。
// LinkedSampleID 为预留字段：Sample 表尚未建立，当前仅允许挂接 Find，
// 但仍执行“文物 / 采样恰选其一”的互斥校验。
type datingReq struct {
	LabName        string `json:"labName"`
	Method         string `json:"method"`
	LinkedFindID   *uint  `json:"linkedFindId"`
	LinkedSampleID *uint  `json:"linkedSampleId"`
	ResultText     string `json:"resultText"`
}

type datingTransitionReq struct {
	Action     string `json:"action"` // submit | result | void
	ResultText string `json:"resultText"`
}

// validateLink 校验送检对象恰选一个且当前支持挂接。
func validateLink(req *datingReq) (int, string) {
	hasFind := req.LinkedFindID != nil && *req.LinkedFindID != 0
	hasSample := req.LinkedSampleID != nil && *req.LinkedSampleID != 0
	if hasFind && hasSample {
		return http.StatusBadRequest, "送检对象只能挂接文物或采样中的一个，不可同时关联"
	}
	if !hasFind && !hasSample {
		return http.StatusBadRequest, "送检对象必须且只能关联一件出土文物（或采样记录）"
	}
	if hasSample {
		return http.StatusBadRequest, "系统尚未建立采样（Sample）记录，当前仅支持挂接出土文物"
	}
	return 0, ""
}

func validDatingMethod(m string) bool {
	return m == models.DatingMethodC14 || m == models.DatingMethodTL
}

func loadDatingSubmission(h *Handler, id uint) (*models.DatingSubmission, error) {
	var sub models.DatingSubmission
	err := h.DB.
		Preload("LinkedFind").
		Preload("LinkedFind.Unit").
		Preload("LinkedFind.Unit.Site").
		Preload("LinkedFind.Material").
		First(&sub, id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

// ---------- Handlers ----------

func (h *Handler) ListDatingSubmissions(c *gin.Context) {
	var subs []models.DatingSubmission
	q := h.DB.
		Preload("LinkedFind").
		Preload("LinkedFind.Unit").
		Preload("LinkedFind.Unit.Site").
		Order("id desc")
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	if method := c.Query("method"); method != "" {
		q = q.Where("method = ?", method)
	}
	if findID := c.Query("findId"); findID != "" {
		q = q.Where("linked_find_id = ?", findID)
	}
	if err := q.Find(&subs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, subs)
}

func (h *Handler) GetDatingSubmission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	sub, err := loadDatingSubmission(h, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "测年送检单不存在"})
		return
	}
	c.JSON(http.StatusOK, sub)
}

func (h *Handler) CreateDatingSubmission(c *gin.Context) {
	var req datingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if req.LabName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "承测实验室必填"})
		return
	}
	if !validDatingMethod(req.Method) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "测年方法仅支持 c14（碳十四）或 tl（热释光）"})
		return
	}
	if code, msg := validateLink(&req); code != 0 {
		c.JSON(code, gin.H{"error": msg})
		return
	}
	var find models.Find
	if err := h.DB.First(&find, *req.LinkedFindID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "关联的出土文物不存在"})
		return
	}

	sub := models.DatingSubmission{
		LabName:      req.LabName,
		Method:       req.Method,
		Status:       models.DatingStatusDraft,
		LinkedFindID: req.LinkedFindID,
		ResultText:   req.ResultText,
	}
	if err := h.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	loaded, err := loadDatingSubmission(h, sub.ID)
	if err != nil {
		c.JSON(http.StatusCreated, sub)
		return
	}
	c.JSON(http.StatusCreated, loaded)
}

// UpdateDatingSubmission 仅 draft 可编辑；已送检及终态单的关联与字段均锁定。
func (h *Handler) UpdateDatingSubmission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	sub, err := loadDatingSubmission(h, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "测年送检单不存在"})
		return
	}
	if sub.Status != models.DatingStatusDraft {
		c.JSON(http.StatusConflict, gin.H{"error": "仅草稿状态的送检单可编辑，当前状态为「" + datingStatusLabel(sub.Status) + "」"})
		return
	}

	var req datingReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}
	if req.LabName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "承测实验室必填"})
		return
	}
	if !validDatingMethod(req.Method) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "测年方法仅支持 c14（碳十四）或 tl（热释光）"})
		return
	}
	if code, msg := validateLink(&req); code != 0 {
		c.JSON(code, gin.H{"error": msg})
		return
	}
	var find models.Find
	if err := h.DB.First(&find, *req.LinkedFindID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "关联的出土文物不存在"})
		return
	}

	sub.LabName = req.LabName
	sub.Method = req.Method
	sub.LinkedFindID = req.LinkedFindID
	sub.ResultText = req.ResultText
	if err := h.DB.Save(sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	loaded, err := loadDatingSubmission(h, sub.ID)
	if err != nil {
		c.JSON(http.StatusOK, sub)
		return
	}
	c.JSON(http.StatusOK, loaded)
}

// TransitionDatingSubmission 驱动状态机：
// draft --submit--> submitted --result--> resulted
//                                  \--void--> void
func (h *Handler) TransitionDatingSubmission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	sub, err := loadDatingSubmission(h, uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "测年送检单不存在"})
		return
	}

	var req datingTransitionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数无效"})
		return
	}

	now := time.Now()
	switch req.Action {
	case "submit":
		if sub.Status != models.DatingStatusDraft {
			c.JSON(http.StatusConflict, gin.H{"error": "非法状态流转：仅草稿（draft）可送检，当前状态为「" + datingStatusLabel(sub.Status) + "」"})
			return
		}
		sub.Status = models.DatingStatusSubmitted
		sub.SubmittedAt = &now
	case "result":
		if sub.Status != models.DatingStatusSubmitted {
			c.JSON(http.StatusConflict, gin.H{"error": "非法状态流转：仅已送检（submitted）可登记结果，当前状态为「" + datingStatusLabel(sub.Status) + "」"})
			return
		}
		resultText := req.ResultText
		if resultText == "" {
			resultText = sub.ResultText
		}
		if resultText == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "登记测年结果时必须填写结果说明"})
			return
		}
		sub.Status = models.DatingStatusResulted
		sub.ResultText = resultText
		sub.ResultedAt = &now
	case "void":
		if sub.Status != models.DatingStatusSubmitted {
			c.JSON(http.StatusConflict, gin.H{"error": "非法状态流转：仅已送检（submitted）可作废，当前状态为「" + datingStatusLabel(sub.Status) + "」"})
			return
		}
		sub.Status = models.DatingStatusVoid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "未知流转动作，仅支持 submit / result / void"})
		return
	}

	if err := h.DB.Save(sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	loaded, err := loadDatingSubmission(h, sub.ID)
	if err != nil {
		c.JSON(http.StatusOK, sub)
		return
	}
	c.JSON(http.StatusOK, loaded)
}

// DeleteDatingSubmission 仅草稿可删除，避免抹除送检痕迹。
func (h *Handler) DeleteDatingSubmission(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	sub, err := loadDatingSubmission(h, uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "测年送检单不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if sub.Status != models.DatingStatusDraft {
		c.JSON(http.StatusConflict, gin.H{"error": "仅草稿状态的送检单可删除，当前状态为「" + datingStatusLabel(sub.Status) + "」"})
		return
	}
	if err := h.DB.Delete(&models.DatingSubmission{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "已删除"})
}

func datingStatusLabel(s string) string {
	switch s {
	case models.DatingStatusDraft:
		return "草稿"
	case models.DatingStatusSubmitted:
		return "已送检"
	case models.DatingStatusResulted:
		return "已出结果"
	case models.DatingStatusVoid:
		return "已作废"
	default:
		return s
	}
}
