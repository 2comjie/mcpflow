package api

import "github.com/gin-gonic/gin"

func (a *API) GetStats(c *gin.Context) {
	stats, err := a.store.GetExecutionStats()
	if err != nil {
		fail(c, 500, err.Error())
		return
	}
	ok(c, stats)
}
