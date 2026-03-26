// backend/router/router.go
package router

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kerke/cronjob-panel/auth"
	"github.com/kerke/cronjob-panel/cluster"
	"github.com/kerke/cronjob-panel/cronjob"
)

func Setup(r *gin.Engine, db *sql.DB, discordCfg auth.DiscordConfig) {
	// Auth routes (no middleware)
	r.GET("/api/auth/discord", auth.DiscordLoginHandler(discordCfg))
	r.GET("/api/auth/discord/callback", auth.DiscordCallbackHandler(discordCfg, db))

	api := r.Group("/api")
	api.Use(auth.RequireAuth(db))

	// Logout
	api.POST("/auth/logout", func(c *gin.Context) {
		cookie, _ := c.Cookie("session")
		auth.DeleteSession(db, cookie)
		c.SetCookie("session", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// Cluster CRUD
	api.GET("/clusters", func(c *gin.Context) {
		clusters, err := cluster.ListClusters(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, clusters)
	})

	api.POST("/clusters", func(c *gin.Context) {
		var body struct {
			Name     string `json:"name"`
			Endpoint string `json:"endpoint"`
			AuthType string `json:"auth_type"`
			AuthData string `json:"auth_data"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
			return
		}
		userID := c.GetString("user_id")
		err := cluster.CreateCluster(db, cluster.Cluster{
			Name:      body.Name,
			Endpoint:  body.Endpoint,
			AuthType:  body.AuthType,
			AuthData:  []byte(body.AuthData),
			CreatedBy: userID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"ok": true})
	})

	api.DELETE("/clusters/:id", func(c *gin.Context) {
		if err := cluster.DeleteCluster(db, c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// CronJob routes
	serviceFactory := func(clusterID string) (*cronjob.Service, error) {
		cl, err := cluster.GetCluster(db, clusterID)
		if err != nil {
			return nil, err
		}
		k8sClient, err := cluster.NewK8sClient(cl.Endpoint, cl.AuthType, cl.AuthData)
		if err != nil {
			return nil, err
		}
		return cronjob.NewService(k8sClient), nil
	}

	h := cronjob.NewHandler(serviceFactory)
	api.GET("/clusters/:id/cronjobs", h.ListCronJobs)
	api.GET("/clusters/:id/namespaces/:ns/cronjobs/:name", h.GetCronJob)
	api.GET("/clusters/:id/namespaces/:ns/cronjobs/:name/jobs", h.ListJobs)
	api.GET("/clusters/:id/namespaces/:ns/pods/:pod/log", h.GetPodLog)
	api.POST("/clusters/:id/namespaces/:ns/cronjobs/:name/trigger", h.TriggerJob)
	api.PATCH("/clusters/:id/namespaces/:ns/cronjobs/:name/suspend", h.SetSuspend)
}
