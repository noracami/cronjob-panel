package cronjob

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type ServiceFactory func(clusterID string) (*Service, error)

type Handler struct {
	getService ServiceFactory
}

func NewHandler(factory ServiceFactory) *Handler {
	return &Handler{getService: factory}
}

func (h *Handler) ListCronJobs(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cronjobs, err := svc.ListCronJobs(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cronjobs)
}

func (h *Handler) GetCronJob(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cj, err := svc.GetCronJob(c.Request.Context(), c.Param("ns"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cj)
}

func (h *Handler) ListJobs(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	jobs, err := svc.ListJobs(c.Request.Context(), c.Param("ns"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

func (h *Handler) GetPodLog(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log, err := svc.GetPodLog(c.Request.Context(), c.Param("ns"), c.Param("pod"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"log": log})
}

func (h *Handler) TriggerJob(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	job, err := svc.TriggerJob(c.Request.Context(), c.Param("ns"), c.Param("name"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, job)
}

func (h *Handler) SetSuspend(c *gin.Context) {
	svc, err := h.getService(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var body struct {
		Suspend bool `json:"suspend"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := svc.SetSuspend(c.Request.Context(), c.Param("ns"), c.Param("name"), body.Suspend); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
