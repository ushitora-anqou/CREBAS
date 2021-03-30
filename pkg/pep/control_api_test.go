package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/naoki9911/CREBAS/pkg/app"
	"github.com/stretchr/testify/assert"
)

func TestGetAllAppInfos(t *testing.T) {
	req := httptest.NewRequest("GET", "/all", nil)
	w := httptest.NewRecorder()
	fetchAllItems(w, req)
	resp := w.Result()
	resbody, _ := ioutil.ReadAll(resp.Body)
	var appInfos []app.AppInfo
	json.Unmarshal(resbody, &appInfos)

	var expectedAppInfos = apps.GetAllAppInfos()
	for id, appInfo := range appInfos {
		assert.Equal(t, appInfo.Id, expectedAppInfos[id].Id, "unmatched ID")
	}
}
